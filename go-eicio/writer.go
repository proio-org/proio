package eicio

import (
	"encoding/binary"
	"github.com/golang/protobuf/proto"
	"io"
)

type Writer struct {
	byteWriter io.Writer
}

func NewWriter(byteWriter io.Writer) *Writer {
	return &Writer{
		byteWriter: byteWriter,
	}
}

func (wrt *Writer) PushEvent(event *Event) (err error) {
	headerBuf, err := proto.Marshal(event.Header)
	if err != nil {
		return
	}

	headerSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSizeBuf, uint32(len(headerBuf)))

	wrt.byteWriter.Write(headerSizeBuf)
	wrt.byteWriter.Write(headerBuf)
	wrt.byteWriter.Write(event.getPayload())

	return
}
