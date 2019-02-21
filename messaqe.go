package main

import (
	"encoding/binary"
	"io"

	"github.com/chmike/go-dmon/dmon"
	"github.com/pkg/errors"
)

// MsgWriter encode and write messages
type MsgWriter interface {
	Write(*dmon.Msg) (int, error)
}

// MsgReader reads and decode messages.
type MsgReader interface {
	Read(*dmon.Msg) (int, error)
}

// JSONWriter is a MsgWriter for JSON.
type JSONWriter struct {
	w   io.Writer
	buf []byte
}

// NewJSONWriter intsantiate a JSONWriter.
func NewJSONWriter(w io.Writer) *JSONWriter {
	return &JSONWriter{w: w, buf: make([]byte, 256)}
}

func (w *JSONWriter) Write(m *dmon.Msg) (int, error) {
	data, err := m.MarshalJSON()
	if err != nil {
		return 0, errors.Wrapf(err, "json encoding error")
	}
	w.buf = w.buf[:4]
	binary.LittleEndian.PutUint32(w.buf, uint32(len(data)))
	w.buf = append(w.buf, data...)

	n, err := w.w.Write(w.buf)
	if err != nil {
		return n, errors.Wrapf(err, "json writer error")
	}
	if n != len(w.buf) {
		return n, errors.Wrapf(err, "short write")
	}
	return n, nil
}

// JSONReader is  MsgReader for JSON.
type JSONReader struct {
	r   io.Reader
	buf []byte
}

// NewJSONReader intsantiate a JSONReader.
func NewJSONReader(r io.Reader) *JSONReader {
	return &JSONReader{r: r, buf: make([]byte, 256)}
}

func (r *JSONReader) Read(m *dmon.Msg) (int, error) {
	var hdr [4]byte
	n, err := io.ReadFull(r.r, hdr[:])
	if err != nil {
		return n, errors.Wrapf(err, "could not read message header")
	}
	msgLen := int(binary.LittleEndian.Uint32(hdr[:]))
	if msgLen > len(r.buf) {
		r.buf = make([]byte, msgLen+100)
	}
	r.buf = r.buf[:msgLen]
	n, err = io.ReadFull(r.r, r.buf)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read message payload")
	}
	msgLen += 4

	err = m.UnmarshalJSON(r.buf)
	if err != nil {
		return msgLen, errors.Wrapf(err, "json encoding error")
	}
	return msgLen, nil
}

// BinaryWriter is a MsgWriter for binary encoding.
type BinaryWriter struct {
	w   io.Writer
	buf []byte
}

// NewBinaryWriter intsantiate a BinaryWriter.
func NewBinaryWriter(w io.Writer) *BinaryWriter {
	return &BinaryWriter{w: w, buf: make([]byte, 256)}
}

func (w *BinaryWriter) Write(m *dmon.Msg) (int, error) {
	data, err := m.MarshalBinary()
	if err != nil {
		return 0, errors.Wrapf(err, "binary encoding error")
	}
	w.buf = w.buf[:4]
	binary.LittleEndian.PutUint32(w.buf, uint32(len(data)))
	w.buf = append(w.buf, data...)

	n, err := w.w.Write(w.buf)
	if err != nil {
		return n, errors.Wrapf(err, "binary writer error")
	}
	if n != len(w.buf) {
		return n, errors.Wrapf(err, "short write")
	}
	return n, nil
}

// BinaryReader is  MsgReader for JSON.
type BinaryReader struct {
	r   io.Reader
	buf []byte
}

// NewBinaryReader intsantiate a BinaryReader.
func NewBinaryReader(r io.Reader) *BinaryReader {
	return &BinaryReader{r: r, buf: make([]byte, 256)}
}

func (r *BinaryReader) Read(m *dmon.Msg) (int, error) {
	var hdr [4]byte
	n, err := io.ReadFull(r.r, hdr[:])
	if err != nil {
		return n, errors.Wrapf(err, "could not read message header")
	}
	msgLen := int(binary.LittleEndian.Uint32(hdr[:]))
	if msgLen > len(r.buf) {
		r.buf = make([]byte, msgLen+100)
	}
	r.buf = r.buf[:msgLen]
	n, err = io.ReadFull(r.r, r.buf)
	if err != nil {
		return 0, errors.Wrapf(err, "could not read message payload")
	}
	msgLen += 4

	err = m.UnmarshalBinary(r.buf)
	if err != nil {
		return msgLen, errors.Wrapf(err, "json encoding error")
	}
	return msgLen, nil
}
