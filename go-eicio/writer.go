package eicio

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"strings"
)

type Writer struct {
	byteWriter io.Writer
}

func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	if strings.HasSuffix(filename, ".gz") {
		return NewGzipWriter(file)
	}

	return NewWriter(file), nil
}

func (wrt *Writer) Close() error {
	flusher, ok := wrt.byteWriter.(Flusher)
	if ok {
		if err := flusher.Flush(); err != nil {
			return err
		}
	}

	return wrt.byteWriter.(io.Closer).Close()
}

func NewWriter(byteWriter io.Writer) *Writer {
	return &Writer{
		byteWriter: byteWriter,
	}
}

func NewGzipWriter(byteWriter io.Writer) (*Writer, error) {
	gzWriter, err := gzip.NewWriterLevel(byteWriter, gzip.BestSpeed)
	if err != nil {
		return nil, err
	}
	return NewWriter(gzWriter), nil
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

type Flusher interface {
	Flush() error
}
