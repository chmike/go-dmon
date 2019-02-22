package dmon

import (
	"io"
	"sync"
	"time"
)

// BufWriter is an io.Writer buffering output data.
type BufWriter struct {
	mtx sync.Mutex
	buf []byte
	n   int
	w   io.Writer
	err error
}

const minBufLen = 256

// NewBufWriter returns an io.Writer with a buffer of size bufLen. The buffer will be
// flushed every period milliseconds.
func NewBufWriter(w io.Writer, bufLen int, period time.Duration) *BufWriter {
	if bufLen < minBufLen {
		bufLen = minBufLen
	}
	b := &BufWriter{
		buf: make([]byte, bufLen),
		w:   w,
	}
	delay := time.Duration(period)
	go func() {
		b.mtx.Lock()
		defer b.mtx.Unlock()
		for b.flush() != nil {
			b.mtx.Unlock()
			time.Sleep(delay)
			b.mtx.Lock()
		}
	}()
	return b
}

// Error return the last error.
func (b *BufWriter) Error() error {
	return b.err
}

// Flush writes the content of the buffer and return the error if any.
// The mutex is required to be locked when called.
func (b *BufWriter) flush() error {
	if b.err == nil && b.n != 0 {
		_, b.err = b.w.Write(b.buf[:b.n])
		b.n = 0
	}
	return b.err
}

// Write bufferize the writing operations.
func (b *BufWriter) Write(p []byte) (int, error) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.err != nil {
		return 0, b.err
	}
	var tot int
	for len(p) > 0 {
		if b.n == len(b.buf) && b.flush() != nil {
			break
		}
		n := copy(b.buf[b.n:], p)
		p = p[n:]
		b.n += n
		tot += n
	}
	return tot, b.err
}

// WriteByte writes a byte in the bufferized writer.
func (b *BufWriter) WriteByte(p byte) error {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	if b.err != nil || (b.n == len(b.buf) && b.flush() != nil) {
		return b.err
	}
	b.buf[b.n] = p
	b.n++
	return nil
}

// BufReader is an io.Reader buffering input data.
type BufReader struct {
	buf []byte
	beg int
	end int
	r   io.Reader
	err error
}

// NewBufReader returns an io.Reader with a buffer of size bufLen.
func NewBufReader(r io.Reader, bufLen int) *BufReader {
	if bufLen < minBufLen {
		bufLen = minBufLen
	}
	b := &BufReader{
		buf: make([]byte, bufLen),
		r:   r,
	}
	return b
}

// Error return the last error.
func (b *BufReader) Error() error {
	return b.err
}

// Read bufferize the read operations.
func (b *BufReader) fetch() error {
	b.beg = 0
	b.end, b.err = b.r.Read(b.buf)
	return b.err
}

// Read bufferize the read operations.
func (b *BufReader) Read(p []byte) (int, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.beg == b.end && b.fetch() != nil {
		return 0, b.err
	}
	n := copy(p, b.buf[b.beg:b.end])
	b.beg += n
	return n, nil
}

// ReadByte reads a byte from the bufferizer reader.
func (b *BufReader) ReadByte() (byte, error) {
	if b.err != nil {
		return 0, b.err
	}
	if b.beg == b.end && b.fetch() != nil {
		return 0, b.err
	}
	p := b.buf[b.beg]
	b.beg++
	return p, nil
}

// ReadFull fills p with bytes from the bufferized reader.
func (b *BufReader) ReadFull(p []byte) (int, error) {
	pLen := len(p)
	for len(p) > 0 {
		if b.beg == b.end && b.fetch() != nil {
			return 0, b.err
		}
		n := copy(p, b.buf[b.beg:b.end])
		b.beg += n
		p = p[n:]
	}
	return pLen, nil
}
