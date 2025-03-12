package xgfile

import (
	"errors"
	"io"
	"os"
	"path/filepath"
)

type Import struct {
	Filename string
}

type Segment struct {
	Filename   string
	File       *os.File
	Type       int
	AutoDelete bool
}

const (
	GDF_HDR = iota
	GDF_IMAGE
	XG_GAMEHDR
	XG_GAMEFILE
	XG_ROLLOUTS
	XG_COMMENT
	ZLIBARC_IDX
	XG_UNKNOWN
)

var EXTENSIONS = []string{"_gdh.bin", ".jpg", "_gamehdr.bin", "_gamefile.bin", "_rollouts.bin", "_comments.bin", "_idx.bin", ""}

func NewSegment(segmentType int, autoDelete bool) (*Segment, error) {
	tmpFile, err := os.CreateTemp("", "tmpXGI")
	if err != nil {
		return nil, err
	}
	return &Segment{
		Filename:   tmpFile.Name(),
		File:       tmpFile,
		Type:       segmentType,
		AutoDelete: autoDelete,
	}, nil
}

func (s *Segment) Close() error {
	if s.File != nil {
		if err := s.File.Close(); err != nil {
			return err
		}
	}
	if s.AutoDelete && s.Filename != "" {
		if err := os.Remove(s.Filename); err != nil {
			return err
		}
	}
	return nil
}

func (s *Segment) CopyTo(dest string) error {
	srcFile, err := os.Open(s.Filename)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	destFile, err := os.Create(dest)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return err
}

func NewImport(filename string) *Import {
	return &Import{Filename: filename}
}

func (imp *Import) GetFileSegment() ([]*Segment, error) {
	var segments []*Segment

	xginfile, err := os.Open(imp.Filename)
	if err != nil {
		return nil, err
	}
	defer xginfile.Close()

	gdfheader := &GameDataFormatHdrRecord{}
	if err := gdfheader.FromStream(xginfile); err != nil {
		return nil, errors.New("not a game data format file")
	}

	segment, err := NewSegment(GDF_HDR, true)
	if err != nil {
		return nil, err
	}
	defer segment.Close()

	xginfile.Seek(0, io.SeekStart)
	block := make([]byte, gdfheader.HeaderSize)
	if _, err := xginfile.Read(block); err != nil {
		return nil, err
	}
	if _, err := segment.File.Write(block); err != nil {
		return nil, err
	}
	segment.File.Sync()
	segments = append(segments, segment)

	if gdfheader.ThumbnailSize > 0 {
		segment, err := NewSegment(GDF_IMAGE, true)
		if err != nil {
			return nil, err
		}
		defer segment.Close()

		xginfile.Seek(gdfheader.ThumbnailOffset, io.SeekStart)
		imgbuf := make([]byte, gdfheader.ThumbnailSize)
		if _, err := xginfile.Read(imgbuf); err != nil {
			return nil, err
		}
		if _, err := segment.File.Write(imgbuf); err != nil {
			return nil, err
		}
		segment.File.Sync()
		segments = append(segments, segment)
	}

	archiveobj, err := NewZlibArchive(imp.Filename)
	if err != nil {
		return nil, err
	}

	for _, filerec := range archiveobj.ArcRegistry {
		segmentFile, segFilename, err := archiveobj.GetArchiveFile(filerec)
		if err != nil {
			return nil, err
		}
		defer segmentFile.Close()
		defer os.Remove(segFilename)

		xgFileType := XG_FILEMAP[filepath.Base(filerec.Name)]
		xgFileSegment := &Segment{
			Filename:   segFilename,
			File:       segmentFile,
			Type:       xgFileType,
			AutoDelete: false,
		}

		if xgFileType == XG_GAMEFILE {
			segmentFile.Seek(XG_GAMEHDR_LEN, io.SeekStart)
			magicStr := make([]byte, 4)
			if _, err := segmentFile.Read(magicStr); err != nil {
				return nil, err
			}
			if string(magicStr) != "DMLI" {
				return nil, errors.New("not a valid XG gamefile")
			}
		}

		segments = append(segments, xgFileSegment)
	}

	return segments, nil
}

var XG_FILEMAP = map[string]int{
	"temp.xgi": XG_GAMEHDR,
	"temp.xgr": XG_ROLLOUTS,
	"temp.xgc": XG_COMMENT,
	"temp.xg":  XG_GAMEFILE,
}

const XG_GAMEHDR_LEN = 556

type GameDataFormatHdrRecord struct {
	// ...existing fields...
}

func (hdr *GameDataFormatHdrRecord) FromStream(stream io.Reader) error {
	// ...existing code...
	return nil
}

type ZlibArchive struct {
	ArcRegistry []FileRecord
	// ...existing fields...
}

func NewZlibArchive(filename string) (*ZlibArchive, error) {
	// ...existing code...
	return nil
}

func (za *ZlibArchive) GetArchiveFile(filerec FileRecord) (*os.File, string, error) {
	// ...existing code...
	return nil, "", nil
}

type FileRecord struct {
	Name string
	// ...existing fields...
}
