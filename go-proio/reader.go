package proio

import (
	"bytes"
	"compress/gzip"
	"encoding/binary"
	"io"
	"os"
	"sync"

	"github.com/decibelcooper/proio/go-proio/proto"
	"github.com/pierrec/lz4"
)

type Reader struct {
	streamReader       io.Reader
	bucket             *bytes.Reader
	bucketDecompressor io.Reader
	bucketReader       io.Reader
	bucketHeader       *proto.BucketHeader
	bucketEventsRead   int

	Err                   chan error
	EventScanBufferSize   int
	deferredUntilStopScan []func()
	getMutex              sync.Mutex
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
		streamReader:        streamReader,
		bucket:              &bytes.Reader{},
		Err:                 make(chan error, 100),
		EventScanBufferSize: 100,
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
	return rdr.next(true)
}

func (rdr *Reader) NextHeader() (*proto.BucketHeader, error) {
	if _, err := rdr.readBucket(1 << 62); err != nil {
		return nil, err
	}
	return rdr.bucketHeader, nil
}

func (rdr *Reader) Skip(nEvents int) (nSkipped int, err error) {
	bucketEventsLeft := 0
	if rdr.bucketHeader != nil {
		bucketEventsLeft = int(rdr.bucketHeader.NEvents) - rdr.bucketEventsRead
	}
	if nEvents > bucketEventsLeft {
		var n int
		for n != 0 && err == nil {
			n, err = rdr.readBucket(nEvents - bucketEventsLeft - nSkipped)
			if err != nil {
				return
			}
			nSkipped += n
		}
	}

	for nSkipped < nEvents {
		_, err = rdr.next(false)
		if err != nil {
			return
		}
		nSkipped++
	}

	return
}

//ScanEvents returns a buffered channel of type Event where all of the events
//in the stream will be pushed.  The channel buffer size is defined by
//Reader.EventScanBufferSize which defaults to 100.  The goroutine responsible
//for fetching events will not break until there are no more events,
//Reader.StopScan() is called, or Reader.Close() is called.  In this scenario,
//errors are pushed to the Reader.Err channel.
func (rdr *Reader) ScanEvents() <-chan *Event {
	events := make(chan *Event, rdr.EventScanBufferSize)
	quit := make(chan int)

	go func() {
		defer close(events)
		for {
			event, err := rdr.Next()
			if err != nil {
				select {
				case rdr.Err <- err:
				default:
				}
			}
			if event == nil {
				return
			}

			select {
			case events <- event:
			case <-quit:
				return
			}
		}
	}()

	rdr.deferUntilStopScan(
		func() {
			close(quit)
		},
	)

	return events
}

// StopScan stops all scans initiated by Reader.ScanEvents()
func (rdr *Reader) StopScan() {
	for _, thisFunc := range rdr.deferredUntilStopScan {
		thisFunc()
	}
	rdr.deferredUntilStopScan = make([]func(), 0)
}

func (rdr *Reader) deferUntilStopScan(thisFunc func()) {
	rdr.deferredUntilStopScan = append(rdr.deferredUntilStopScan, thisFunc)
}

func (rdr *Reader) next(doUnmarshal bool) (*Event, error) {
	protoSizeBuf := make([]byte, 4)
	if err := readBytes(rdr.bucketReader, protoSizeBuf); err != nil {
		if err == io.EOF {
			_, err = rdr.readBucket(0)
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
	rdr.bucketEventsRead++

	var event *Event
	if doUnmarshal {
		eventProto := &proto.Event{}
		if err := eventProto.Unmarshal(protoBuf); err != nil {
			return nil, err
		}

		event = newEventFromProto(eventProto)
	}

	return event, nil
}

func (rdr *Reader) readBucket(maxSkipEvents int) (eventsSkipped int, err error) {
	rdr.bucketEventsRead = 0

	_, err = rdr.syncToMagic()
	if err != nil {
		return
	}

	headerSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.streamReader, headerSizeBuf); err != nil {
		return
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)

	headerBuf := make([]byte, headerSize)
	if err = readBytes(rdr.streamReader, headerBuf); err != nil {
		return
	}
	rdr.bucketHeader = &proto.BucketHeader{}
	if err = rdr.bucketHeader.Unmarshal(headerBuf); err != nil {
		return
	}

	if int(rdr.bucketHeader.NEvents) > maxSkipEvents {
		bucketBytes := make([]byte, rdr.bucketHeader.BucketSize)
		if err = readBytes(rdr.streamReader, bucketBytes); err != nil {
			return
		}
		rdr.bucket.Reset(bucketBytes)
	} else {
		rdr.bucketReader = nil
		eventsSkipped = int(rdr.bucketHeader.NEvents)
		seeker, ok := rdr.streamReader.(io.Seeker)
		if ok {
			seekBytes(seeker, int64(rdr.bucketHeader.BucketSize))
		} else {
			bucketBytes := make([]byte, rdr.bucketHeader.BucketSize)
			err = readBytes(rdr.streamReader, bucketBytes)
		}
		return
	}

	switch rdr.bucketHeader.Compression {
	case proto.BucketHeader_GZIP:
		gzipRdr, ok := rdr.bucketDecompressor.(*gzip.Reader)
		if ok {
			gzipRdr.Reset(rdr.bucket)
		} else {
			gzipRdr, err = gzip.NewReader(rdr.bucket)
			if err != nil {
				return
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

	return
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

func seekBytes(seeker io.Seeker, nBytes int64) error {
	start, err := seeker.Seek(0, 1 /*io.SeekCurrent*/)
	if err != nil {
		return err
	}

	tot := int64(0)
	for tot < nBytes {
		n, err := seeker.Seek(int64(nBytes-tot), 1 /*io.SeekCurrent*/)
		tot += n - start
		if err != nil && tot != nBytes {
			return err
		}
	}
	return nil
}
