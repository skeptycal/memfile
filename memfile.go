package memfile

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
)

type (
	MemFile interface {
		fmt.Stringer
		io.ReadWriteCloser
	}
	memFile struct {
		filename string
		io.ReadWriteCloser
		mu sync.RWMutex
	}
)

// NewMemFile returns an interface to an
// in-memory implementation of the file
// named filename.
//
// The MemFile implements io.ReadWriteCloser
// Closing the MemFile destroys the MemFile
// and frees any memory allocated.
//
// If the file does not exist or is not
// readable, an error is returned.
func NewMemFile(filename string) (MemFile, error) {

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	mf := &memFile{
		filename:        fi.Name(),
		ReadWriteCloser: newBufferSize(fi.Size()),
	}

	// this Lock() is likely not needed since this object does
	// not exist until it is returned by this function ...
	mf.mu.Lock()
	_, err = io.Copy(mf, f)
	mf.mu.Unlock()

	if err != nil {
		mf = nil
		return nil, err
	}

	return mf, err
}

func (m *memFile) Close() error {
	if m == nil || m.ReadWriteCloser == nil {
		return os.ErrInvalid
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.filename = ""
	m.ReadWriteCloser.Close()
	m.ReadWriteCloser = nil
	return nil
}

func (m *memFile) String() string {
	return "MemFile: " + m.filename
}

// buffer contains a bytes.Buffer and adds a
// Close() function to allow better compatibility
// with file system implementations
type buffer struct {
	bytes.Buffer
}

// Close closes the buffer by discarding all bytes from
// the underlying bytes.Buffer and setting the buffer to
// a new, empty buffer with a capacity of zero.
func (buf *buffer) Close() error {
	if buf.Cap() == 0 {
		return nil
	}

	buf.Truncate(0)
	buf = newBufferSize(0)
	buf.Reset()

	return nil
}

// newBufferSize returns a new, empty buffer
// with the specified capacity
func newBufferSize(cap int64) *buffer {
	return &buffer{*bytes.NewBuffer(make([]byte, 0, cap))}
}
