package proio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"

	"github.com/decibelcooper/proio/go-proio/proto"
	"github.com/pierrec/lz4"
)

type Compression int

const (
	UNCOMPRESSED Compression = iota
	GZIP
	LZ4
)

type Writer struct {
	streamWriter io.Writer
	bucket       *bytes.Buffer
	bucketWriter io.Writer
	bucketEvents uint64
	bucketComp   proto.BucketHeader_CompType
}

func Create(filename string) (*Writer, error) {
	file, err := os.Create(filename)
	if err != nil {
		return nil, err
	}

	return NewWriter(file), nil
}

func (wrt *Writer) Flush() error {
	if wrt.bucket.Len() > 0 {
		err := wrt.writeBucket()
		if err != nil {
			return err
		}
	}
	return nil
}

func (wrt *Writer) Close() error {
	err := wrt.Flush()
	if err != nil {
		return err
	}
	closer, ok := wrt.streamWriter.(io.Closer)
	if ok {
		closer.Close()
	}
	return nil
}

func NewWriter(streamWriter io.Writer) *Writer {
	writer := &Writer{
		streamWriter: streamWriter,
		bucket:       &bytes.Buffer{},
	}

	writer.SetCompression(LZ4)

	return writer
}

func (wrt *Writer) SetCompression(comp Compression) error {
	wrt.Flush()

	switch comp {
	case GZIP:
		wrt.bucketWriter = gzip.NewWriter(wrt.bucket)
		wrt.bucketComp = proto.BucketHeader_GZIP
	case LZ4:
		wrt.bucketWriter = lz4.NewWriter(wrt.bucket)
		wrt.bucketComp = proto.BucketHeader_LZ4
	case UNCOMPRESSED:
		wrt.bucketWriter = wrt.bucket
		wrt.bucketComp = proto.BucketHeader_NONE
	default:
		return errors.New("invalid compression type")
	}

	return nil
}

func (wrt *Writer) Push(event *Event) error {
	event.flushCache()
	protoBuf, err := event.proto.Marshal()
	if err != nil {
		return err
	}

	protoSizeBuf := make([]byte, 4)
	binary.LittleEndian.PutUint32(protoSizeBuf, uint32(len(protoBuf)))

	if err := writeBytes(wrt.bucketWriter, protoSizeBuf); err != nil {
		return err
	}
	if err := writeBytes(wrt.bucketWriter, protoBuf); err != nil {
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

type writerResettable interface {
	Reset(io.Writer)
}

func (wrt *Writer) writeBucket() error {
	closer, ok := wrt.bucketWriter.(io.Closer)
	if ok {
		closer.Close()
	}

	bucketBytes := wrt.bucket.Bytes()
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
	wrtReset, ok := wrt.bucketWriter.(writerResettable)
	if ok {
		wrtReset.Reset(wrt.bucket)
	}

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
