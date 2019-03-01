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

// check that the server’s name in the certificate matches the hosst name
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
	for {
		m.Stamp = time.Now().UTC()
		n, err := sendMessage(&m)
		if err != nil {
			log.Fatalf("send message: %+v", err)
		}
		statUpdate(n)
	}
}

func sendMessage(m *dmon.Msg) (int, error) {
	var (
		err  error
		conn net.Conn
	)
	// open connection
	if *tlsFlag {
		var clientCert tls.Certificate
		clientCert, err = tls.LoadX509KeyPair(clientCRTFilename, clientKeyFilename)
		if err != nil {
			return 0, errors.Wrapf(err, "could not load X509 certificate")
		}
		config := tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: !serverDNSNameCheck,
		}
		conn, err = tls.Dial("tcp", *addressFlag, &config)
	} else {
		conn, err = net.Dial("tcp", *addressFlag)
	}
	if err != nil {
		return 0, errors.Wrap(err, "open connection failed")
	}
	defer conn.Close()

	// encode message
	buf := make([]byte, 8, 512)
	switch *msgCodecFlag {
	case "json":
		buf, err = m.JSONEncode(buf)
	case "binary":
		buf, err = m.BinaryEncode(buf)
	}

	// set message header
	copy(buf[:4], []byte("DMON"))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(len(buf)-8))

	// send message
	conn.SetWriteDeadline(time.Now().Add(timeOutDelay))
	n, err := conn.Write(buf)
	if err != nil {
		return 0, errors.Wrap(err, "send message")
	}
	if n != len(buf) {
		err = errors.Errorf("short write: expected %d, got %d", len(buf), n)
		return 0, errors.Wrap(err, "send message")
	}

	// receive acknowledgment
	var b [1]byte
	conn.SetReadDeadline(time.Now().Add(timeOutDelay))
	n, err = conn.Read(b[:])
	if n != 1 {
		if err != nil {
			return 0, errors.Wrap(err, "recv acknowledgment")
		}
		err = errors.Errorf("expected 1 byte, got %d", n)
		return 0, errors.Wrap(err, "recv acknowledgment")
	}
	if b[0] != ackCode {
		err = errors.Errorf("expected ack byte %+X, got %+X", ackCode, b[0])
		return 0, errors.Wrap(err, "recv acknowledgment")
	}
	return len(buf), nil
}
