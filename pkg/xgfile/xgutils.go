package xgfile

import (
	"bytes"
	"encoding/binary"
	"io"
	"time"
)

// StreamCRC32 computes the CRC32 on a given stream.
func StreamCRC32(stream io.ReadSeeker, numBytes int64, startPos int64, blkSize int) (uint32, error) {
	var crc32 uint32
	curStreamPos, err := stream.Seek(0, io.SeekCurrent)
	if err != nil {
		return 0, err
	}

	if startPos != 0 {
		_, err = stream.Seek(startPos, io.SeekStart)
		if err != nil {
			return 0, err
		}
	}

	buf := make([]byte, blkSize)
	if numBytes == 0 {
		for {
			n, err := stream.Read(buf)
			if n > 0 {
				crc32 = crc32Update(crc32, buf[:n])
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				return 0, err
			}
		}
	} else {
		bytesLeft := numBytes
		for bytesLeft > 0 {
			if int64(blkSize) > bytesLeft {
				blkSize = int(bytesLeft)
			}
			n, err := stream.Read(buf[:blkSize])
			if n > 0 {
				crc32 = crc32Update(crc32, buf[:n])
				bytesLeft -= int64(n)
			}
			if err != nil {
				return 0, err
			}
		}
	}

	_, err = stream.Seek(curStreamPos, io.SeekStart)
	if err != nil {
		return 0, err
	}
	return crc32, nil
}

// crc32Update updates the CRC32 checksum with the given data.
func crc32Update(crc uint32, data []byte) uint32 {
	return crc32 ^ binary.LittleEndian.Uint32(data)
}

// UTF16IntArrayToStr converts an array of UTF-16 integers to a string.
func UTF16IntArrayToStr(intArray []uint16) string {
	var buf bytes.Buffer
	for _, intval := range intArray {
		if intval == 0 {
			break
		}
		buf.WriteRune(rune(intval))
	}
	return buf.String()
}

// DelphiDateTimeConv converts a Delphi datetime to a Go time.Time.
func DelphiDateTimeConv(delphiDatetime float64) time.Time {
	days := int64(delphiDatetime)
	seconds := int64((delphiDatetime - float64(days)) * 86400)
	return time.Date(1899, 12, 30, 0, 0, 0, 0, time.UTC).AddDate(0, 0, int(days)).Add(time.Duration(seconds) * time.Second)
}

// DelphiShortStrToStr converts a Delphi short string to a Go string.
func DelphiShortStrToStr(shortStr []byte) string {
	length := int(shortStr[0])
	return string(shortStr[1 : length+1])
}
