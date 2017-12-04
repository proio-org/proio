package proio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"

	"github.com/decibelcooper/proio/go-proio/proto"
	"github.com/pierrec/lz4"
)

type Reader struct {
	streamReader       io.Reader
	bucket             *bytes.Reader
	bucketDecompressor io.Reader
	bucketReader       io.Reader
	bucketHeader       *proto.BucketHeader
	bucketEventsRead   uint64
}

func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewReader(file), nil
}

func NewReader(streamReader io.Reader) *Reader {
	rdr := &Reader{
		streamReader: streamReader,
		bucket:       &bytes.Reader{},
	}
	rdr.bucketReader = rdr.bucket

	return rdr
}

func (rdr *Reader) Close() {
	closer, ok := rdr.streamReader.(io.Closer)
	if ok {
		closer.Close()
	}
}

func (rdr *Reader) Next() (*Event, error) {
	protoSizeBuf := make([]byte, 4)
	if err := readBytes(rdr.bucketReader, protoSizeBuf); err != nil {
		if err == io.EOF {
			err = rdr.readBucket()
			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}

		if err := readBytes(rdr.bucketReader, protoSizeBuf); err != nil {
			return nil, err
		}
	}

	protoSize := binary.LittleEndian.Uint32(protoSizeBuf)

	protoBuf := make([]byte, protoSize)
	if err := readBytes(rdr.bucketReader, protoBuf); err != nil {
		return nil, err
	}
	eventProto := &proto.Event{}
	if err := eventProto.Unmarshal(protoBuf); err != nil {
		return nil, err
	}

	event := newEventFromProto(eventProto)

	return event, nil
}

func (rdr *Reader) readBucket() error {
	rdr.bucketEventsRead = 0

	_, err := rdr.syncToMagic()
	if err != nil {
		return err
	}

	headerSizeBuf := make([]byte, 4)
	if err := readBytes(rdr.streamReader, headerSizeBuf); err != nil {
		return err
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)

	headerBuf := make([]byte, headerSize)
	if err := readBytes(rdr.streamReader, headerBuf); err != nil {
		return err
	}
	rdr.bucketHeader = &proto.BucketHeader{}
	if err := rdr.bucketHeader.Unmarshal(headerBuf); err != nil {
		return err
	}

	bucketBytes := make([]byte, rdr.bucketHeader.BucketSize)
	if err := readBytes(rdr.streamReader, bucketBytes); err != nil {
		return err
	}
	rdr.bucket.Reset(bucketBytes)

	switch rdr.bucketHeader.Compression {
	case proto.BucketHeader_GZIP:
		gzipRdr, ok := rdr.bucketDecompressor.(*gzip.Reader)
		if ok {
			gzipRdr.Reset(rdr.bucket)
		} else {
			gzipRdr, err = gzip.NewReader(rdr.bucket)
			if err != nil {
				return err
			}
			rdr.bucketDecompressor = gzipRdr
		}
		bucketRdr, ok := rdr.bucketReader.(*bytes.Buffer)
		if ok {
			bucketRdr.Reset()
		} else {
			bucketRdr = &bytes.Buffer{}
		}
		bucketRdr.ReadFrom(gzipRdr)
		rdr.bucketReader = bucketRdr
	case proto.BucketHeader_LZ4:
		lz4Rdr, ok := rdr.bucketDecompressor.(*lz4.Reader)
		if ok {
			lz4Rdr.Reset(rdr.bucket)
		} else {
			lz4Rdr = lz4.NewReader(rdr.bucket)
			rdr.bucketDecompressor = lz4Rdr
		}
		bucketRdr, ok := rdr.bucketReader.(*bytes.Buffer)
		if ok {
			bucketRdr.Reset()
		} else {
			bucketRdr = &bytes.Buffer{}
		}
		bucketRdr.ReadFrom(lz4Rdr)
		rdr.bucketReader = bucketRdr
	default:
		rdr.bucketReader = rdr.bucket
	}

	return nil
}

func (rdr *Reader) syncToMagic() (int, error) {
	magicByteBuf := make([]byte, 1)
	nRead := 0
	for {
		err := readBytes(rdr.streamReader, magicByteBuf)
		if err != nil {
			return nRead, err
		}
		nRead++

		if magicByteBuf[0] == magicBytes[0] {
			var goodSeq = true
			for i := 1; i < len(magicBytes); i++ {
				err := readBytes(rdr.streamReader, magicByteBuf)
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

func readBytes(rdr io.Reader, buf []byte) error {
	tot := 0
	for tot < len(buf) {
		n, err := rdr.Read(buf[tot:])
		tot += n
		if err != nil && tot != len(buf) {
			return err
		}
	}
	return nil
}
