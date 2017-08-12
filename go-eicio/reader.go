package eicio

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
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

func (rdr *Reader) syncToMagic() (int, error) {
	magicByteBuf := make([]byte, 1)
	nRead := 0
	for {
		err := readBytes(rdr.byteReader, magicByteBuf)
		if err != nil {
			return nRead, err
		}
		nRead++

		if magicByteBuf[0] == magicBytes[0] {
			var goodSeq = true
			for i := 1; i < 4; i++ {
				err := readBytes(rdr.byteReader, magicByteBuf)
				if err != nil {
					return nRead, err
				}
				nRead++

				if magicByteBuf[0] != magicBytes[i] {
					goodSeq = false
					break
				}
			}

			if goodSeq {
				break
			}
		}
	}

	return nRead, nil
}

var (
	Resync    = errors.New("data stream had to be resynchronized")
	Truncated = errors.New("data stream is truncated early")
)

func (rdr *Reader) Next() (*Event, error) {
	n, err := rdr.syncToMagic()
	if err != nil {
		return nil, err
	}

	headerSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return nil, Truncated
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	headerBuf := make([]byte, headerSize)
	if err = readBytes(rdr.byteReader, headerBuf); err != nil {
		return nil, Truncated
	}
	header := &EventHeader{}
	if err = header.Unmarshal(headerBuf); err != nil {
		return nil, Truncated
	}

	payloadSize := uint32(0)
	for _, collHdr := range header.Collection {
		payloadSize += collHdr.PayloadSize
	}
	payload := make([]byte, payloadSize)
	if err = readBytes(rdr.byteReader, payload); err != nil {
		return nil, Truncated
	}

	event := &Event{}
	event.Header = header
	event.setPayload(payload)

	if n != 4 {
		err = Resync
	} else {
		err = nil
	}

	return event, err
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
