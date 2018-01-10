package proio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"sync"

	"github.com/decibelcooper/proio/go-proio/proto"
	"github.com/pierrec/lz4"
)

type Compression int

const (
	UNCOMPRESSED Compression = iota
	GZIP
	LZ4
)

// Writer serves to write Events into the proio format.
type Writer struct {
	streamWriter io.Writer
	bucket       *bytes.Buffer
	bucketEvents uint64
	bucketComp   proto.BucketHeader_CompType

	deferredUntilClose []func() error

	sync.Mutex
}

// Create makes a new file specified by filename, overwriting any existing
// file, and returns a Writer for the file.  Either NewWriter or Create must be
// used to construct a Writer.
func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	writer := NewWriter(file)
	writer.deferUntilClose(file.Close)

	return writer, nil
}

// Flush flushes any of the Writer's bucket contents.
func (wrt *Writer) Flush() error {
	if wrt.bucket.Len() > 0 {
		err := wrt.writeBucket()
		if err != nil {
			return err
		}
	}
	return nil
}

// Close calls Flush and closes any file that was created by the library.
// Close does not close io.Writers passed directly to NewWriter.
func (wrt *Writer) Close() error {
	for _, thisFunc := range wrt.deferredUntilClose {
		if err := thisFunc(); err != nil {
			return err
		}
	}
	return nil
}

// NewWriter takes an io.Writer and wraps it in a new proio Writer.  Either
// NewWriter or Create must be used to construct a Writer.
func NewWriter(streamWriter io.Writer) *Writer {
	writer := &Writer{
		streamWriter: streamWriter,
		bucket:       &bytes.Buffer{},
	}

	writer.SetCompression(LZ4)
	writer.deferUntilClose(writer.Flush)

	return writer
}

// Set compression type, for example to GZIP or UNCOMPRESSED.  This can be
// called even after writing some events.
func (wrt *Writer) SetCompression(comp Compression) error {
	switch comp {
	case GZIP:
		wrt.bucketComp = proto.BucketHeader_GZIP
	case LZ4:
		wrt.bucketComp = proto.BucketHeader_LZ4
	case UNCOMPRESSED:
		wrt.bucketComp = proto.BucketHeader_NONE
	default:
		return errors.New("invalid compression type")
	}

	return nil
}

// Serialize the given Event.  Once this is performed, changes to the Event in
// memory are not reflected in the output stream.
func (wrt *Writer) Push(event *Event) error {
	wrt.Lock()
	defer wrt.Unlock()

	event.flushCache()
	protoBuf, err := event.proto.Marshal()
	if err != nil {
		return err
	}

	protoSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(protoSizeBuf, uint32(len(protoBuf)))

	if err := writeBytes(wrt.bucket, protoSizeBuf); err != nil {
		return err
	}
	if err := writeBytes(wrt.bucket, protoBuf); err != nil {
		return err
	}

	wrt.bucketEvents++

	if wrt.bucket.Len() > bucketDumpSize {
		if err := wrt.writeBucket(); err != nil {
			return err
		}
	}

	return nil
}

const bucketDumpSize = 0x1000000

var magicBytes = [...]byte{
	byte(0xe1),
	byte(0xc1),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
	byte(0x00),
}

func (wrt *Writer) writeBucket() error {
	bucketBytes := wrt.bucket.Bytes()
	switch wrt.bucketComp {
	case proto.BucketHeader_GZIP:
		buffer := &bytes.Buffer{}
		gzipWriter := gzip.NewWriter(buffer)
		gzipWriter.Write(bucketBytes)
		gzipWriter.Close()
		bucketBytes = buffer.Bytes()
	case proto.BucketHeader_LZ4:
		buffer := &bytes.Buffer{}
		lz4Writer := lz4.NewWriter(buffer)
		lz4Writer.HighCompression = true
		lz4Writer.Write(bucketBytes)
		lz4Writer.Close()
		bucketBytes = buffer.Bytes()
	}
	header := &proto.BucketHeader{
		NEvents:     wrt.bucketEvents,
		BucketSize:  uint64(len(bucketBytes)),
		Compression: wrt.bucketComp,
	}
	headerBuf, err := header.Marshal()
	if err != nil {
		return err
	}

	headerSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSizeBuf, uint32(len(headerBuf)))

	if err := writeBytes(wrt.streamWriter, magicBytes[:]); err != nil {
		return err
	}
	if err := writeBytes(wrt.streamWriter, headerSizeBuf); err != nil {
		return err
	}
	if err := writeBytes(wrt.streamWriter, headerBuf); err != nil {
		return err
	}
	if err := writeBytes(wrt.streamWriter, bucketBytes); err != nil {
		return err
	}

	wrt.bucketEvents = 0
	wrt.bucket.Reset()

	return nil
}

func writeBytes(wrt io.Writer, buf []byte) error {
	tot := 0
	for tot < len(buf) {
		n, err := wrt.Write(buf[tot:])
		tot += n
		if err != nil && tot != len(buf) {
			return err
		}
	}
	return nil
}

func (wrt *Writer) deferUntilClose(thisFunc func() error) {
	wrt.deferredUntilClose = append(wrt.deferredUntilClose, thisFunc)
}
