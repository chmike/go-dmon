package main

import (
	"bufio"
	"io"
	"net"
	"sync"
	"time"
)

type bufConn struct {
	r    *bufio.Reader
	w    *bufio.Writer
	mtx  sync.Mutex
	conn net.Conn
}

func bufferizedConn(conn net.Conn, period int, bufLen int) io.ReadWriteCloser {
	bconn := &bufConn{
		r:    bufio.NewReaderSize(conn, bufLen),
		w:    bufio.NewWriterSize(conn, bufLen),
		conn: conn,
	}
	go func() {
		var err error
		for err == nil {
			time.Sleep(time.Duration(period) * time.Millisecond)
			bconn.mtx.Lock()
			err = bconn.w.Flush()
			bconn.mtx.Unlock()
		}
	}()
	return bconn
}

func (b *bufConn) Read(p []byte) (int, error) {
	return b.conn.Read(p)
}

func (b *bufConn) Write(p []byte) (int, error) {
	return b.conn.Write(p)
}

func (b *bufConn) Close() error {
	return b.conn.Close()
}
