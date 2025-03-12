package xgfile

import (
	"bytes"
	"compress/zlib"
	"encoding/binary"
	"errors"
	"io"
	"os"
)

type Error struct {
	Value string
}

func (e *Error) Error() string {
	return e.Value
}

type ArchiveRecord struct {
	CRC                uint32
	FileCount          int32
	Version            int32
	RegistrySize       int32
	ArchiveSize        int32
	CompressedRegistry bool
	Reserved           [12]byte
}

func (ar *ArchiveRecord) FromStream(stream io.Reader) error {
	return binary.Read(stream, binary.LittleEndian, ar)
}

type FileRecord struct {
	Name             string
	Path             string
	OSize            int32
	CSize            int32
	Start            int32
	CRC              uint32
	Compressed       bool
	CompressionLevel byte
}

func (fr *FileRecord) FromStream(stream io.Reader) error {
	var nameBytes [256]byte
	var pathBytes [256]byte
	if err := binary.Read(stream, binary.LittleEndian, &nameBytes); err != nil {
		return err
	}
	if err := binary.Read(stream, binary.LittleEndian, &pathBytes); err != nil {
		return err
	}
	fr.Name = string(bytes.Trim(nameBytes[:], "\x00"))
	fr.Path = string(bytes.Trim(pathBytes[:], "\x00"))
	return binary.Read(stream, binary.LittleEndian, fr)
}

type ZlibArchive struct {
	ArcRec         ArchiveRecord
	ArcRegistry    []FileRecord
	StartOfArcData int64
	EndOfArcData   int64
	Filename       string
	Stream         *os.File
	MaxBufSize     int
}

func NewZlibArchive(filename string) (*ZlibArchive, error) {
	stream, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	za := &ZlibArchive{
		Filename:   filename,
		Stream:     stream,
		MaxBufSize: 32768,
	}
	if err := za.getArchiveIndex(); err != nil {
		return nil, err
	}
	return za, nil
}

func (za *ZlibArchive) extractSegment(isCompressed bool, numBytes int) (string, error) {
	tmpFile, err := os.CreateTemp("", "tmpXGI")
	if err != nil {
		return "", err
	}
	defer tmpFile.Close()

	if isCompressed {
		decomp, err := zlib.NewReader(za.Stream)
		if err != nil {
			return "", err
		}
		defer decomp.Close()
		if _, err := io.Copy(tmpFile, decomp); err != nil {
			return "", err
		}
	} else {
		if numBytes <= 0 {
			return "", errors.New("invalid number of bytes for uncompressed segment")
		}
		if _, err := io.CopyN(tmpFile, za.Stream, int64(numBytes)); err != nil {
			return "", err
		}
	}
	return tmpFile.Name(), nil
}

func (za *ZlibArchive) getArchiveIndex() error {
	curStreamPos, err := za.Stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return err
	}
	defer za.Stream.Seek(curStreamPos, io.SeekStart)

	if _, err := za.Stream.Seek(-int64(binary.Size(ArchiveRecord{})), io.SeekEnd); err != nil {
		return err
	}
	za.EndOfArcData, _ = za.Stream.Seek(0, io.SeekCurrent)
	if err := za.ArcRec.FromStream(za.Stream); err != nil {
		return err
	}

	if _, err := za.Stream.Seek(-int64(binary.Size(ArchiveRecord{}))-int64(za.ArcRec.RegistrySize), io.SeekEnd); err != nil {
		return err
	}
	za.StartOfArcData, _ = za.Stream.Seek(0, io.SeekCurrent)
	za.StartOfArcData -= int64(za.ArcRec.ArchiveSize)

	idxFilename, err := za.extractSegment(za.ArcRec.CompressedRegistry, 0)
	if err != nil {
		return err
	}
	defer os.Remove(idxFilename)

	idxFile, err := os.Open(idxFilename)
	if err != nil {
		return err
	}
	defer idxFile.Close()

	for i := 0; i < int(za.ArcRec.FileCount); i++ {
		var filerec FileRecord
		if err := filerec.FromStream(idxFile); err != nil {
			return err
		}
		za.ArcRegistry = append(za.ArcRegistry, filerec)
	}
	return nil
}

func (za *ZlibArchive) GetArchiveFile(filerec FileRecord) (*os.File, error) {
	if _, err := za.Stream.Seek(int64(filerec.Start)+za.StartOfArcData, io.SeekStart); err != nil {
		return nil, err
	}
	tmpFilename, err := za.extractSegment(filerec.Compressed, int(filerec.CSize))
	if err != nil {
		return nil, err
	}
	tmpFile, err := os.Open(tmpFilename)
	if err != nil {
		return nil, err
	}
	return tmpFile, nil
}

func (za *ZlibArchive) SetBlockSize(blksize int) {
	za.MaxBufSize = blksize
}
