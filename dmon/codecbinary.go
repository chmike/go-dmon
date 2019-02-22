package dmon

import (
	"encoding/binary"
	"io"

	"github.com/pkg/errors"
)

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
	defer errors.Wrap(err, "binary encode")
	w.buf = w.buf[:4]
	if w.buf, err = m.MarshalBinary(w.buf); err != nil {
		return
	}
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
	return n, m.UnmarshalBinary(r.buf)
}

// MarshalBinary appends binary encoded Msg to data.
func (o *Msg) MarshalBinary(data []byte) ([]byte, error) {
	var b [8]byte
	sub, err := o.Stamp.MarshalBinary()
	if err != nil {
		return data, err
	}
	data = append(data, byte(len(sub)))
	data = append(data, sub...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(o.Level)))
	data = append(data, b[:4]...)
	data = append(data, []byte(o.Level)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(o.System)))
	data = append(data, b[:4]...)
	data = append(data, []byte(o.System)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(o.Component)))
	data = append(data, b[:4]...)
	data = append(data, []byte(o.Component)...)
	binary.LittleEndian.PutUint32(b[:4], uint32(len(o.Message)))
	data = append(data, b[:4]...)
	data = append(data, []byte(o.Message)...)
	return data, nil
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler
func (o *Msg) UnmarshalBinary(data []byte) (err error) {
	n := int(data[0])
	data = data[1:]
	if err = o.Stamp.UnmarshalBinary(data[:n]); err != nil {
		return
	}
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	o.Level = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	o.System = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	o.Component = string(data[:n])
	data = data[n:]
	n = int(binary.LittleEndian.Uint32(data[:4]))
	data = data[4:]
	o.Message = string(data[:n])
	data = data[n:]
	return err
}
