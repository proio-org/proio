package eicio

import (
	"compress/gzip"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"

	"github.com/decibelcooper/eicio/go-eicio/model"
)

type Reader struct {
	byteReader         io.Reader
	deferredUntilClose []func() error
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
	} else {
		reader = NewReader(file)
	}
	reader.deferUntilClose(file.Close)

	return reader, nil
}

// Closes anything created by Open() or NewGzipReader()
func (rdr *Reader) Close() error {
	for _, thisFunc := range rdr.deferredUntilClose {
		if err := thisFunc(); err != nil {
			return err
		}
	}
	return nil
}

func (rdr *Reader) deferUntilClose(thisFunc func() error) {
	rdr.deferredUntilClose = append(rdr.deferredUntilClose, thisFunc)
}

// Returns a new Reader for reading events from a stream
func NewReader(byteReader io.Reader) *Reader {
	return &Reader{
		byteReader: byteReader,
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

// Returns the next even upon success.  If the data stream is not aligned with
// the beginning of an event, the stream will be resynchronized to the next
// event, and ErrResync will be returned along with the event.
func (rdr *Reader) Get() (*Event, error) {
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
	header := &model.EventHeader{}
	if err = header.Unmarshal(headerBuf); err != nil {
		return nil, ErrTruncated
	}

	payload := make([]byte, payloadSize)
	if err = readBytes(rdr.byteReader, payload); err != nil {
		return nil, ErrTruncated
	}

	event := NewEvent()
	event.Header = header
	event.setPayload(payload)

	if n != 4 {
		err = ErrResync
	}

	return event, err
}

// Get the next Header only from the stream, and seek past the collection
// payload if possible.  This is useful for parsing the metadata of a file or
// stream.
func (rdr *Reader) GetHeader() (*model.EventHeader, error) {
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
	header := &model.EventHeader{}
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
