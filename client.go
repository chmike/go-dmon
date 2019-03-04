package main

import (
	"crypto/tls"
	"encoding/binary"
	"log"
	"net"
	"time"

	"github.com/chmike/go-dmon/dmon"
	_ "github.com/go-sql-driver/mysql"
	"github.com/pkg/errors"
)

var timeOutDelay = 30 * time.Second

// check that the server’s name in the certificate matches the host name
const serverDNSNameCheck = false

func runAsClient() {
	log.SetPrefix("client ")
	log.Println("target:", *addressFlag)

	m := dmon.Msg{
		Stamp:     time.Now().UTC(),
		Level:     "info",
		System:    "dmon",
		Component: "test",
		Message:   "no problem",
	}
	statStart(time.Duration(*periodFlag) * time.Second)
	lms := &MsgLogSrv{Address: *addressFlag}
	for {
		m.Stamp = time.Now().UTC()
		n := lms.SendMessage(&m)
		if lms.err != nil {
			log.Printf("send message: %+v, wait 2 seconds", lms.err)
			time.Sleep(2 * time.Second)
		}
		statUpdate(n)
	}
}

// MsgLogSrv holds a cached connection to the logging server.
type MsgLogSrv struct {
	Address string
	conn    net.Conn
	err     error
}

// Error returns the last error.
func (lms *MsgLogSrv) Error() error {
	return lms.err
}

// SendMessage send the message m to the logging server.
func (lms *MsgLogSrv) SendMessage(m *dmon.Msg) (n int) {
	defer func() {
		if lms.conn != nil && lms.err != nil {
			lms.conn.Close()
			lms.conn = nil
		}
	}()

	if lms.conn == nil || lms.err != nil {
		lms.tryConnect()
	}
	if lms.conn == nil || lms.err != nil {
		return 0
	}

	// encode message
	buf := make([]byte, 8, 512)
	switch *msgCodecFlag {
	case "json":
		buf, lms.err = m.JSONEncode(buf)
	case "binary":
		buf, lms.err = m.BinaryEncode(buf)
	}
	if lms.err != nil {
		lms.err = errors.Wrap(lms.err, "send message")
		return 0
	}

	// set message header
	copy(buf[:4], []byte("DMON"))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)-8))

	// send message
	lms.err = lms.conn.SetWriteDeadline(time.Now().Add(timeOutDelay))
	if lms.err != nil {
		lms.err = errors.Wrap(lms.err, "send message")
		return 0
	}

	n, lms.err = lms.conn.Write(buf)
	if lms.err != nil {
		lms.err = errors.Wrap(lms.err, "send message")
		return 0
	}
	if n != len(buf) {
		lms.err = errors.Errorf("short write: expected %d, got %d", len(buf), n)
		lms.err = errors.Wrap(lms.err, "send message")
		return 0
	}

	// receive acknowledgment
	var b [1]byte
	lms.err = lms.conn.SetReadDeadline(time.Now().Add(timeOutDelay))
	if lms.err != nil {
		lms.err = errors.Wrap(lms.err, "recv acknowledgment")
		return 0
	}
	n, lms.err = lms.conn.Read(b[:])
	if n != 1 {
		if lms.err != nil {
			lms.err = errors.Wrap(lms.err, "recv acknowledgment")
			return 0
		}
		lms.err = errors.Errorf("expected 1 byte, got %d", n)
		lms.err = errors.Wrap(lms.err, "recv acknowledgment")
		return 0
	}
	if b[0] != ackCode {
		lms.err = errors.Errorf("expected ack byte %+X, got %+X", ackCode, b[0])
		lms.err = errors.Wrap(lms.err, "recv acknowledgment")
		return 0
	}
	return len(buf)
}

func (lms *MsgLogSrv) tryConnect() {
	var err error
	if *tlsFlag {
		var clientCert tls.Certificate
		clientCert, err = tls.LoadX509KeyPair(clientCRTFilename, clientKeyFilename)
		if err != nil {
			lms.conn = nil
			lms.err = errors.Wrapf(err, "could not load X509 certificate")
			return
		}
		config := tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: !serverDNSNameCheck,
		}
		lms.conn, lms.err = tls.Dial("tcp", *addressFlag, &config)
	} else {
		lms.conn, lms.err = net.Dial("tcp", *addressFlag)
	}
}
