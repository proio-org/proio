package proio

import (
	"bytes"
	"io"
	"os"
)

type Reader struct {
	streamReader     io.Reader
	bucket           *bytes.Buffer
	bucketReader     io.Reader
	bucketEventsRead uint64
}

func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewReader(file), nil
}

func NewReader(streamReader io.Reader) *Reader {
	return &Reader{
		streamReader: streamReader,
	}
}

func (rdr *Reader) Close() {
	closer, ok := rdr.streamReader.(io.Closer)
	if ok {
		closer.Close()
	}
}

func (rdr *Reader) NextEvent() (*Event, error) {
	if rdr.bucket.Len() == 0 {
		if err := rdr.readBucket(); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (rdr *Reader) readBucket() error {
	return io.EOF
}
