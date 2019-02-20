package main

import (
	"io"
	"log"
	"net"
	"sync"
	"time"
)

// buffer is a io.ReadWriteCloser encapsulating read and write
// buffering with periodic buffer flush.
type buffer struct {
	recvMtx sync.Mutex
	sendMtx sync.Mutex
	rBuf    []byte
	sBuf    []byte
	rBeg    int
	rEnd    int
	sBeg    int
	sEnd    int
	conn    net.Conn
	period  time.Duration
}

func newSender(conn net.Conn, period time.Duration, bufLen int) *buffer {
	s := &buffer{
		rBuf:   make([]byte, bufLen),
		sBuf:   make([]byte, bufLen),
		conn:   conn,
		period: period,
	}
	go periodicWriter(s)
	return s
}

func (s *buffer) Read(p []byte) (int, error) {
	s.recvMtx.Lock()
	defer s.recvMtx.Unlock()
	if s.rBeg == s.rEnd {
		s.rBeg = 0
		s.recvMtx.Unlock()
		n, err := s.conn.Read(s.rBuf)
		s.recvMtx.Lock()
		if err != nil {
			return n, err
		}
		s.rEnd = n
	}
	n := copy(p, s.rBuf[s.rBeg:s.rEnd])
	s.rBeg += n
	return n, nil
}

func (s *buffer) Write(p []byte) (int, error) {
	s.sendMtx.Lock()
	defer s.sendMtx.Unlock()

	tot := 0
	for len(p) > 0 {
		n := copy(s.sBuf[s.sEnd:], p)
		s.sEnd += n
		p = p[n:]
		tot += n
		if s.sEnd == len(s.sBuf) {
			if s.period == 0 {
				return tot, io.EOF
			}
			_, err := s.conn.Write(s.sBuf[s.sBeg:s.sEnd])
			if err != nil {
				return tot, err
			}
			s.sEnd = 0
		}
	}
	return tot, nil
}

func (s *buffer) Close() error {
	s.sendMtx.Lock()
	s.recvMtx.Lock()
	s.period = 0
	s.sendMtx.Unlock()
	s.recvMtx.Unlock()
	return nil
}

func periodicWriter(s *buffer) {
	for {
		time.Sleep(s.period)
		s.sendMtx.Lock()
		if s.period == 0 {
			s.sendMtx.Unlock()
			s.conn.Close()
			break
		}
		if s.sEnd > 0 {
			_, err := s.conn.Write(s.sBuf[s.sBeg:s.sEnd])
			if err != nil {
				log.Fatalln("periodic sender error:", err)
			}
			s.sEnd = 0
		}
		s.sendMtx.Unlock()
	}
}
