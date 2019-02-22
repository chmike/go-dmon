package dmon

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"io"

	"github.com/pkg/errors"
)

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
