package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"net"
	"sync"

	"github.com/pkg/errors"
)

var monEntryPool = sync.Pool{New: func() interface{} { return new(monEntry) }}

func runAsServer() {
	log.SetPrefix("server ")

	monEntryChan := make(chan *monEntry, 1000)
	go database(monEntryChan)

	var listener net.Listener
	var err error
	if *tlsFlag {
		// listen for a TLS connection
		var serverCert tls.Certificate
		serverCert, err = tls.LoadX509KeyPair(serverCRTFilename, serverKeyFilename)
		if err != nil {
			log.Fatal(err)
		}

		config := tls.Config{
			Certificates: []tls.Certificate{serverCert},
			ClientAuth:   tls.RequireAndVerifyClientCert,
			ClientCAs:    certPool,
		}
		config.Rand = rand.Reader
		listener, err = tls.Listen("tcp", *addressFlag, &config)
		if err != nil {
			log.Fatalln("failed listen:", err)
		}
	} else {
		listener, err = net.Listen("tcp", *addressFlag)
	}

	log.Println("listen:", *addressFlag)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleClient(conn, monEntryChan)
	}
}

func handleClient(conn net.Conn, monEntryChan chan *monEntry) {
	defer conn.Close()

	for {
		m := monEntryPool.New().(*monEntry)
		err := recvMsg(m, conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("error:", err)
		}

		monEntryChan <- m

		_, err = io.WriteString(conn, "ack")
		if err != nil {
			log.Println("send error:", err)
			break
		}

	}
	log.Println("conn closed")
}

type buffer struct {
	buf []byte
	len int
}

func newBuffer() *buffer {
	return &buffer{
		buf: make([]byte, 512),
		len: 0,
	}
}

func recvMsg(m *monEntry, conn net.Conn) error {
	var hdr [4]byte
	_, err := io.ReadFull(conn, hdr[:])
	if err != nil {
		return errors.Wrapf(err, "could not read message header")
	}
	buf := make([]byte, binary.LittleEndian.Uint32(hdr[:]))
	_, err = io.ReadFull(conn, buf)
	if err != nil {
		return errors.Wrapf(err, "could not read message payload")
	}

	err = json.Unmarshal(buf, m)
	if err != nil {
		return errors.Wrapf(err, "could not decode message payload: %s", string(buf))
	}
	return nil
}
