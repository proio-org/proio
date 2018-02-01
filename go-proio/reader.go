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

// Reader serves to read Events from a stream in the proio format.
type Reader struct {
	streamReader     io.Reader
	bucket           *bytes.Reader
	bucketReader     io.Reader
	bucketHeader     *proto.BucketHeader
	bucketEventsRead int

	Err                   chan error
	EvtScanBufSize        int
	deferredUntilStopScan []func()
	scanLock              sync.Mutex

	deferredUntilClose []func() error

	sync.Mutex
}

// Open opens the given existing file (in read-only mode), returning an error
// where appropriate.  Upon success, a new Reader is created to wrap the file,
// and returned.  Either Open or NewReader should be called to construct a new
// Reader.
func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	return NewReader(file), nil
}

// NewReader wraps an existing io.Reader for reading proio Events.  Either Open
// or NewReader should be called to construct a new Reader.
func NewReader(streamReader io.Reader) *Reader {
	rdr := &Reader{
		streamReader:   streamReader,
		bucket:         &bytes.Reader{},
		bucketReader:   &bytes.Buffer{},
		Err:            make(chan error, evtScanBufferSize),
		EvtScanBufSize: evtScanBufferSize,
	}

	return rdr
}

// Close closes any file that was opened by the library, and stops any
// unfinished scans.  Close does not close io.Readers passed directly to
// NewReader.
func (rdr *Reader) Close() {
	rdr.Lock()
	defer rdr.Unlock()

	rdr.StopScan()
	closer, ok := rdr.streamReader.(io.Closer)
	if ok {
		closer.Close()
	}
}

// Next retrieves the next event from the stream.
func (rdr *Reader) Next() (*Event, error) {
	rdr.Lock()
	defer rdr.Unlock()

	return rdr.readFromBucket(true)
}

// NextHeader returns the next bucket header from the stream, and discards the
// bucket payload.
func (rdr *Reader) NextHeader() (*proto.BucketHeader, error) {
	rdr.Lock()
	defer rdr.Unlock()

	if _, err := rdr.readBucket(1 << 62); err != nil {
		return nil, err
	}
	return rdr.bucketHeader, nil
}

// Skip skips nEvents events.  If the return error is nil, nEvents have been
// skipped.
func (rdr *Reader) Skip(nEvents int) (nSkipped int, err error) {
	rdr.Lock()
	defer rdr.Unlock()

	bucketEventsLeft := 0
	if rdr.bucketHeader != nil {
		bucketEventsLeft = int(rdr.bucketHeader.NEvents) - rdr.bucketEventsRead
	}
	if nEvents > bucketEventsLeft {
		nSkipped += bucketEventsLeft
		for {
			var n int
			n, err = rdr.readBucket(nEvents - nSkipped)
			if err != nil {
				return
			}
			if n == 0 {
				break
			}
			nSkipped += n
		}
	}

	for nSkipped < nEvents {
		_, err = rdr.readFromBucket(false)
		if err != nil {
			return
		}
		nSkipped++
	}

	return
}

// SeekToStart seeks seekable streams to the beginning, and prepares the stream
// to read from there.
func (rdr *Reader) SeekToStart() error {
	rdr.Lock()
	defer rdr.Unlock()

	seeker, ok := rdr.streamReader.(io.Seeker)
	if !ok {
		return errors.New("stream not seekable")
	}

	for {
		n, err := seeker.Seek(0, 0 /*io.SeekStart*/)
		if err != nil {
			return err
		}
		if n == 0 {
			break
		}
	}

	rdr.readBucket(0)

	return nil
}

//ScanEvents returns a buffered channel of type Event where all of the events
//in the stream will be pushed.  The channel buffer size is defined by
//Reader.EvtScanBufSize which defaults to 100.  The goroutine responsible
//for fetching events will not break until there are no more events,
//Reader.StopScan() is called, or Reader.Close() is called.  In this scenario,
//errors are pushed to the Reader.Err channel.
func (rdr *Reader) ScanEvents() <-chan *Event {
	rdr.Lock()
	defer rdr.Unlock()

	events := make(chan *Event, rdr.EvtScanBufSize)
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
	rdr.scanLock.Lock()
	defer rdr.scanLock.Unlock()

	for _, thisFunc := range rdr.deferredUntilStopScan {
		thisFunc()
	}
	rdr.deferredUntilStopScan = make([]func(), 0)
}

var evtScanBufferSize int = 100

func (rdr *Reader) deferUntilStopScan(thisFunc func()) {
	rdr.scanLock.Lock()
	defer rdr.scanLock.Unlock()

	rdr.deferredUntilStopScan = append(rdr.deferredUntilStopScan, thisFunc)
}

func (rdr *Reader) readFromBucket(doUnmarshal bool) (*Event, error) {
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
		rdr.bucketReader = &bytes.Buffer{}
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
		gzipRdr, ok := rdr.bucketReader.(*gzip.Reader)
		if ok {
			gzipRdr.Reset(rdr.bucket)
		} else {
			gzipRdr, err = gzip.NewReader(rdr.bucket)
			if err != nil {
				return
			}
		}
		rdr.bucketReader = gzipRdr
	case proto.BucketHeader_LZ4:
		lz4Rdr, ok := rdr.bucketReader.(*lz4.Reader)
		if ok {
			lz4Rdr.Reset(rdr.bucket)
		} else {
			lz4Rdr = lz4.NewReader(rdr.bucket)
		}
		rdr.bucketReader = lz4Rdr
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

func (rdr *Reader) deferUntilClose(thisFunc func() error) {
	rdr.deferredUntilClose = append(rdr.deferredUntilClose, thisFunc)
}
