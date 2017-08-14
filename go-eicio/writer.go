package eicio

import (
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"strings"
)

type Writer struct {
	byteWriter         io.Writer
	deferredUntilClose []func() error
}

func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	var writer *Writer
	if strings.HasSuffix(filename, ".gz") {
		writer, err = NewGzipWriter(file)
		if err != nil {
			file.Close()
			return nil, err
		}
	} else {
		writer = NewWriter(file)
	}
	writer.deferUntilClose(file.Close)

	return writer, nil
}

// closes anything created by Create() or NewGzipWriter()
func (wrt *Writer) Close() error {
	for _, thisFunc := range wrt.deferredUntilClose {
		if err := thisFunc(); err != nil {
			return err
		}
	}
	return nil
}

func (wrt *Writer) deferUntilClose(thisFunc func() error) {
	wrt.deferredUntilClose = append(wrt.deferredUntilClose, thisFunc)
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

	writer := NewWriter(gzWriter)
	writer.deferUntilClose(gzWriter.Flush)
	writer.deferUntilClose(gzWriter.Close)

	return writer, nil
}

var magicBytes = [...]byte{
	byte(0xe1),
	byte(0xc1),
	byte(0x00),
	byte(0x00),
}

func (wrt *Writer) PushEvent(event *Event) (err error) {
	headerBuf, err := event.Header.Marshal()
	if err != nil {
		return
	}

	headerSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSizeBuf, uint32(len(headerBuf)))

	wrt.byteWriter.Write(magicBytes[:])
	wrt.byteWriter.Write(headerSizeBuf)
	wrt.byteWriter.Write(headerBuf)
	wrt.byteWriter.Write(event.getPayload())

	return
}
