package dmon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"
	"time"

	"github.com/pkg/errors"
)

// Msg is a monitoring log meessage.
type Msg struct {
	Stamp     time.Time `json:"stamp"`
	Level     string    `json:"level"`
	System    string    `json:"system"`
	Component string    `json:"component"`
	Message   string    `json:"message"`
}

// MsgWriter encode and write messages
type MsgWriter interface {
	Write(*Msg) (int, error)
}

// MsgReader reads and decode messages.
type MsgReader interface {
	Read(*Msg) (int, error)
}

// JSONWriter is a MsgWriter for JSON.
type JSONWriter struct {
	w io.Writer
	e *json.Encoder
	b bytes.Buffer
}

// NewJSONWriter intsantiate a JSONWriter.
func NewJSONWriter(w io.Writer) *JSONWriter {
	j := &JSONWriter{w: w}
	j.e = json.NewEncoder(&j.b)
	return j
}

func (j *JSONWriter) Write(m *Msg) (n int, err error) {
	defer errors.Wrap(err, "json encode")
	var buf [4]byte
	j.b.Reset()
	if _, err = j.b.Write(buf[:]); err != nil {
		return 0, err
	}
	if err = j.e.Encode(m); err != nil {
		return 0, err
	}
	data := j.b.Bytes()
	binary.LittleEndian.PutUint32(data, uint32(len(data)-4))
	n, err = j.w.Write(data)
	return n, err
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

func (j *JSONReader) Read(m *Msg) (n int, err error) {
	defer errors.Wrap(err, "json decode")
	var hdr [4]byte
	n, err = io.ReadFull(j.r, hdr[:])
	if err != nil {
		return
	}
	msgLen := int(binary.LittleEndian.Uint32(hdr[:]))
	if msgLen > len(j.buf) {
		j.buf = make([]byte, msgLen+100)
	}
	j.buf = j.buf[:msgLen]
	n, err = io.ReadFull(j.r, j.buf)
	n += 4
	if err != nil {
		return
	}
	err = json.Unmarshal(j.buf, m)
	return
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

func (w *BinaryWriter) Write(m *Msg) (n int, err error) {
	var b [8]byte
	defer errors.Wrap(err, "binary encode")
	// reserve space for message length
	w.buf = w.buf[:4]
	sub, err := m.Stamp.MarshalBinary()
	if err != nil {
		return 0, err
	}
	w.buf = append(w.buf, byte(len(sub)))
	w.buf = append(w.buf, sub...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(m.Level)))
	w.buf = append(w.buf, b[:4]...)
	w.buf = append(w.buf, []byte(m.Level)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(m.System)))
	w.buf = append(w.buf, b[:4]...)
	w.buf = append(w.buf, []byte(m.System)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(m.Component)))
	w.buf = append(w.buf, b[:4]...)
	w.buf = append(w.buf, []byte(m.Component)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(m.Message)))
	w.buf = append(w.buf, b[:4]...)
	w.buf = append(w.buf, []byte(m.Message)...)
	binary.LittleEndian.PutUint32(w.buf, uint32(len(w.buf)-4))
	n, err = w.w.Write(w.buf)
	return
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

func (r *BinaryReader) Read(m *Msg) (n int, err error) {
	defer errors.Wrap(err, "binary decode")
	var hdr [4]byte
	if n, err = io.ReadFull(r.r, hdr[:]); err != nil {
		return
	}
	msgLen := int(binary.LittleEndian.Uint32(hdr[:]))
	if msgLen > len(r.buf) {
		r.buf = make([]byte, msgLen+100)
	}
	r.buf = r.buf[:msgLen]
	n, err = io.ReadFull(r.r, r.buf)
	n += 4
	if err != nil {
		return
	}
	data := r.buf
	l := int(data[0])
	data = data[1:]
	if err = m.Stamp.UnmarshalBinary(data[:l]); err != nil {
		return
	}
	data = data[l:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	m.Level = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	m.System = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	m.Component = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	m.Message = string(data[:n])
	data = data[n:]
	return
}
