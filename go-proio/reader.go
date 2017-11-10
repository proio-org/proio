package proio

import (
	"bufio"
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"sync"

	"github.com/decibelcooper/proio/go-proio/proto"
	"github.com/pierrec/lz4"
)

type Reader struct {
	Err                   chan error
	EventScanBufferSize   int
	byteReader            io.Reader
	deferredUntilClose    []func() error
	deferredUntilStopScan []func()
	getMutex              sync.Mutex
}

// Opens a file and adds the file as an io.Reader to a new Reader that is
// returned.  If the file name ends with ".gz", the file is wrapped with
// gzip.NewReader().  If the function returns successful (err == nil), the
// Close() function should be called when finished.
func Open(filename string) (*Reader, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	var reader *Reader
	if strings.HasSuffix(filename, ".gz") {
		reader, err = NewGzipReader(file)
		if err != nil {
			file.Close()
			return nil, err
		}
	} else if strings.HasSuffix(filename, ".lz4") {
		reader = NewLZ4Reader(file)
	} else {
		reader = NewReader(file)
	}
	reader.deferUntilClose(file.Close)

	return reader, nil
}

// Closes anything created by Open() or NewGzipReader()
func (rdr *Reader) Close() error {
	rdr.StopScan()
	for _, thisFunc := range rdr.deferredUntilClose {
		if err := thisFunc(); err != nil {
			return err
		}
	}
	close(rdr.Err)

	return nil
}

func (rdr *Reader) deferUntilClose(thisFunc func() error) {
	rdr.deferredUntilClose = append(rdr.deferredUntilClose, thisFunc)
}

// Returns a new Reader for reading events from a stream
func NewReader(byteReader io.Reader) *Reader {
	return &Reader{
		byteReader:          byteReader,
		Err:                 make(chan error, 100),
		EventScanBufferSize: 100,
	}
}

// Opens a gzip stream and adds it as an io.Reader to a new Reader that is
// returned.  The Close() function should be called before closing the stream.
func NewGzipReader(byteReader io.Reader) (*Reader, error) {
	gzReader, err := gzip.NewReader(byteReader)
	if err != nil {
		return nil, err
	}

	reader := NewReader(gzReader)
	reader.deferUntilClose(gzReader.Close)

	return reader, nil
}

func NewLZ4Reader(byteReader io.Reader) *Reader {
	lz4Reader := lz4.NewReader(byteReader)
	buffReader := bufio.NewReaderSize(lz4Reader, 0x1000000)
	reader := NewReader(buffReader)

	return reader
}

func (rdr *Reader) syncToMagic() (int, error) {
	magicByteBuf := make([]byte, 1)
	nRead := 0
	for {
		err := readBytes(rdr.byteReader, magicByteBuf)
		if err != nil {
			return nRead, err
		}
		nRead++

		if magicByteBuf[0] == magicBytes[0] {
			var goodSeq = true
			for i := 1; i < 4; i++ {
				err := readBytes(rdr.byteReader, magicByteBuf)
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

var (
	ErrResync    = errors.New("data stream had to be resynchronized")
	ErrTruncated = errors.New("data stream is truncated early")
)

// Get() returns the next even upon success.  If the data stream is not aligned
// with the beginning of an event, the stream will be resynchronized to the
// next event, and ErrResync will be returned along with the event.
func (rdr *Reader) Get() (*Event, error) {
	rdr.getMutex.Lock()
	defer rdr.getMutex.Unlock()

	n, err := rdr.syncToMagic()
	if err != nil {
		return nil, err
	}

	headerSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return nil, ErrTruncated
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	payloadSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, payloadSizeBuf); err != nil {
		return nil, ErrTruncated
	}
	payloadSize := binary.LittleEndian.Uint32(payloadSizeBuf)

	headerBuf := make([]byte, headerSize)
	if err = readBytes(rdr.byteReader, headerBuf); err != nil {
		return nil, ErrTruncated
	}
	header := &proto.EventHeader{}
	if err = header.Unmarshal(headerBuf); err != nil {
		return nil, ErrTruncated
	}

	payload := make([]byte, payloadSize)
	if err = readBytes(rdr.byteReader, payload); err != nil {
		return nil, ErrTruncated
	}

	event := NewEvent()
	event.header = header
	event.payload = payload

	if n != 4 {
		err = ErrResync
	}

	return event, err
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
			event, err := rdr.Get()
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

func (rdr *Reader) deferUntilStopScan(thisFunc func()) {
	rdr.deferredUntilStopScan = append(rdr.deferredUntilStopScan, thisFunc)
}

// StopScan stops all scans initiated by Reader.ScanEvents()
func (rdr *Reader) StopScan() {
	for _, thisFunc := range rdr.deferredUntilStopScan {
		thisFunc()
	}
	rdr.deferredUntilStopScan = make([]func(), 0)
}

// Get the next Header only from the stream, and seek past the collection
// payload if possible.  This is useful for parsing the metadata of a file or
// stream.
func (rdr *Reader) GetHeader() (*proto.EventHeader, error) {
	rdr.getMutex.Lock()
	defer rdr.getMutex.Unlock()

	n, err := rdr.syncToMagic()
	if err != nil {
		return nil, err
	}

	headerSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, headerSizeBuf); err != nil {
		return nil, ErrTruncated
	}
	headerSize := binary.LittleEndian.Uint32(headerSizeBuf)
	payloadSizeBuf := make([]byte, 4)
	if err = readBytes(rdr.byteReader, payloadSizeBuf); err != nil {
		return nil, ErrTruncated
	}
	payloadSize := binary.LittleEndian.Uint32(payloadSizeBuf)

	headerBuf := make([]byte, headerSize)
	if err = readBytes(rdr.byteReader, headerBuf); err != nil {
		return nil, ErrTruncated
	}
	header := &proto.EventHeader{}
	if err = header.Unmarshal(headerBuf); err != nil {
		return nil, ErrTruncated
	}

	seeker, ok := rdr.byteReader.(io.Seeker)
	if ok {
		if err = seekBytes(seeker, int64(payloadSize)); err != nil {
			return header, ErrTruncated
		}
	} else {
		payload := make([]byte, payloadSize)
		if err = readBytes(rdr.byteReader, payload); err != nil {
			return header, ErrTruncated
		}
	}

	if n != 4 {
		err = ErrResync
	} else {
		err = nil
	}

	return header, err
}

// Skip the next nEvents events
func (rdr *Reader) Skip(nEvents int) (int, error) {
	rdr.getMutex.Lock()
	defer rdr.getMutex.Unlock()

	seeker, isSeeker := rdr.byteReader.(io.Seeker)
	wasResynced := false

	nSkipped := 0
	for i := 0; i < nEvents; i++ {
		n, err := rdr.syncToMagic()
		if err != nil {
			return nSkipped, err
		}
		if n != 4 {
			wasResynced = true
		}

		headerSizeBuf := make([]byte, 4)
		if err = readBytes(rdr.byteReader, headerSizeBuf); err != nil {
			return nSkipped, ErrTruncated
		}
		headerSize := binary.LittleEndian.Uint32(headerSizeBuf)

		payloadSizeBuf := make([]byte, 4)
		if err = readBytes(rdr.byteReader, payloadSizeBuf); err != nil {
			return nSkipped, ErrTruncated
		}
		payloadSize := binary.LittleEndian.Uint32(payloadSizeBuf)

		if isSeeker {
			if err = seekBytes(seeker, int64(headerSize+payloadSize)); err != nil {
				return nSkipped, ErrTruncated
			}
		} else {
			headerBuf := make([]byte, headerSize+payloadSize)
			if err = readBytes(rdr.byteReader, headerBuf); err != nil {
				return nSkipped, ErrTruncated
			}
		}

		nSkipped++
	}

	var err error = nil
	if wasResynced {
		err = ErrResync
	}

	return nSkipped, err
}

var ErrNotSeekable = errors.New("data stream is not seekable")

// If the stream implements io.Seeker (typically a file), reset back to the
// beginning of the file.
func (rdr *Reader) SeekToStart() error {
	rdr.getMutex.Lock()
	defer rdr.getMutex.Unlock()

	seeker, ok := rdr.byteReader.(io.Seeker)
	if !ok {
		return ErrNotSeekable
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

	return nil
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
