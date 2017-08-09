package eicio

import (
	"encoding/binary"
	"io"
	"os"
)

type Writer struct {
	byteWriter io.Writer
}

func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}
	return NewWriter(file), nil
}

func (wrt *Writer) Close() {
	wrt.byteWriter.(*os.File).Close()
}

func NewWriter(byteWriter io.Writer) *Writer {
	return &Writer{
		byteWriter: byteWriter,
	}
}

func (wrt *Writer) PushEvent(event *Event) (err error) {
	headerBuf, err := event.Header.Marshal()
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
