package eicio

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"strings"
)

type Reader struct {
	byteReader io.Reader
}

func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filename, ".gz") {
		return NewGzipReader(file)
	}

	return NewReader(file), nil
}

func (rdr *Reader) Close() error {
	return rdr.byteReader.(io.Closer).Close()
}

func NewReader(byteReader io.Reader) *Reader {
	return &Reader{
		byteReader: byteReader,
	}
}

func NewGzipReader(byteReader io.Reader) (*Reader, error) {
	gzReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return nil, err
	}
	return NewReader(gzReader), nil
}

func (rdr *Reader) GetEvent() (*Event, error) {
	headerSizeBuf := make([]byte, 4)
	if err := readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return nil, err
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	headerBuf := make([]byte, headerSize)
	if err := readBytes(rdr.byteReader, headerBuf); err != nil {
		return nil, err
	}
	header := &EventHeader{}
	if err := header.Unmarshal(headerBuf); err != nil {
		return nil, err
	}

	payloadSize := uint32(0)
	for _, collHdr := range header.Collection {
		payloadSize += collHdr.PayloadSize
	}
	payload := make([]byte, payloadSize)
	if err := readBytes(rdr.byteReader, payload); err != nil {
		return nil, err
	}

	event := &Event{}
	event.Header = header
	event.setPayload(payload)
	return event, nil
}

func readBytes(rdr io.Reader, buf []byte) error {
	tot := 0
	for tot < len(buf) {
		n, err := rdr.Read(buf[tot:])
		if err != nil {
			return err
		}

		tot += n
	}
	return nil
}
