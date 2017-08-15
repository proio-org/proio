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

// creates a new file (overwriting existing file) and adds the file as an
// io.Writer to a new Writer that is returned.  If the file name ends with
// ".gz", the file is wrapped with gzip.NewWriterLevel() with level
// gzip.BestSpeed.  If the function returns successful (err == nil), the
// Close() function should be called when finished.
func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	var writer *Writer
	if strings.HasSuffix(filename, ".gz") {
		writer = NewGzipWriter(file)
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

func NewGzipWriter(byteWriter io.Writer) *Writer {
	gzWriter := gzip.NewWriter(byteWriter)
	writer := NewWriter(gzWriter)
	writer.deferUntilClose(gzWriter.Close)

	return writer
}

var magicBytes = [...]byte{
	byte(0xe1),
	byte(0xc1),
	byte(0x00),
	byte(0x00),
}

func (wrt *Writer) PushEvent(event *Event) (err error) {
	if err := event.flushCollCache(); err != nil {
		return err
	}

	headerBuf, err := event.Header.Marshal()
	if err != nil {
		return
	}

	headerSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSizeBuf, uint32(len(headerBuf)))

	payload := event.getPayload()
	payloadSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(payloadSizeBuf, uint32(len(payload)))

	wrt.byteWriter.Write(magicBytes[:])
	wrt.byteWriter.Write(headerSizeBuf)
	wrt.byteWriter.Write(payloadSizeBuf)
	wrt.byteWriter.Write(headerBuf)
	wrt.byteWriter.Write(payload)

	return
}
