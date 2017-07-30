package eicio

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
)

type Reader struct {
	byteReader io.Reader
}

func NewReader(byteReader io.Reader) *Reader {
	return &Reader{
		byteReader: byteReader,
	}
}

func (rdr *Reader) GetEvent() (event *Event, err error) {
	headerSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	headerBuf := make([]byte, headerSize)
	if err = readBytes(rdr.byteReader, headerBuf); err != nil {
		return
	}
	header := &EventHeader{}
	if err = proto.Unmarshal(headerBuf, header); err != nil {
		return
	}

	payloadSize := uint32(0)
	for _, collHdr := range header.Collection {
		payloadSize += collHdr.PayloadSize
	}
	payload := make([]byte, payloadSize)
	if err = readBytes(rdr.byteReader, payload); err != nil {
		return
	}

	event = &Event{}
	event.Header = header
	event.setPayload(payload)
	return
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
