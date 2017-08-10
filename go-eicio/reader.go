package eicio

import (
	"encoding/binary"
	"io"
	"os"
)

type Reader struct {
	byteReader io.Reader
}

func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	return NewReader(file), nil
}

func (rdr *Reader) Close() {
	rdr.byteReader.(*os.File).Close()
}

func NewReader(byteReader io.Reader) *Reader {
	return &Reader{
		byteReader: byteReader,
	}
}

func (rdr *Reader) GetEvent() *Event {
	headerSizeBuf := make([]byte, 4)
	if err := readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return nil
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	headerBuf := make([]byte, headerSize)
	if err := readBytes(rdr.byteReader, headerBuf); err != nil {
		return nil
	}
	header := &EventHeader{}
	if err := header.Unmarshal(headerBuf); err != nil {
		return nil
	}

	payloadSize := uint32(0)
	for _, collHdr := range header.Collection {
		payloadSize += collHdr.PayloadSize
	}
	payload := make([]byte, payloadSize)
	if err := readBytes(rdr.byteReader, payload); err != nil {
		return nil
	}

	event := &Event{}
	event.Header = header
	event.setPayload(payload)
	return event
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
