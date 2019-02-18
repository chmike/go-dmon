package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"
	"sync"
)

var monEntryPool = sync.Pool{New: func() interface{} { return new(monEntry) }}

func runAsServer() {
	log.SetPrefix("server ")

	monEntryChan := make(chan *monEntry)
	go database(monEntryChan)

	// listen for a TLS connection
	serverCert, err := tls.LoadX509KeyPair(serverCRTFilename, serverKeyFilename)
	if err != nil {
		log.Fatal(err)
	}

	config := tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    certPool,
	}
	config.Rand = rand.Reader
	listener, err := tls.Listen("tcp", *addressFlag, &config)
	if err != nil {
		log.Fatalln("failed listen:", err)
	}
	log.Println("listen:", *addressFlag)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("server: accept: %s", err)
			break
		}
		defer conn.Close()
		log.Printf("server: accepted from %s", conn.RemoteAddr())
		go handleClient(conn, monEntryChan)
	}
}

func handleClient(conn net.Conn, monEntryChan chan *monEntry) {
	defer conn.Close()

	b := newBuffer()
	for {
		m := monEntryPool.New().(*monEntry)

		err := recvMsg(m, b, conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("error:", err)
		}

		monEntryChan <- m

		_, err = io.WriteString(conn, "ack")
		if err != nil {
			log.Fatalln("send error:", err)
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

func recvMsg(m *monEntry, b *buffer, conn net.Conn) error {
	pos := 0
loop:
	for {
		for pos < b.len {
			if b.buf[pos] == ':' {
				break loop
			}
			pos++
		}
		if err := recvBytes(b, conn); err != nil {
			return err
		}
	}
	strLen, err := strconv.Atoi(string(b.buf[:pos]))
	if err != nil {
		return fmt.Errorf("readMsg: decode length: %s", err)
	}
	strBeg := pos + 1
	strEnd := strBeg + strLen
	for b.len < strEnd {
		if err := recvBytes(b, conn); err != nil {
			return err
		}
	}
	err = json.Unmarshal(b.buf[strBeg:strEnd], m)
	if err != nil {
		return fmt.Errorf("readMsg: json decode: %s", err)
	}
	b.len = copy(b.buf, b.buf[strEnd:b.len])
	return nil
}

func recvBytes(b *buffer, conn net.Conn) error {
	if b.len == len(b.buf) {
		tmp := make([]byte, b.len*2)
		copy(tmp, b.buf)
		b.buf = tmp
	}
	n, err := conn.Read(b.buf[b.len:])
	if err != nil {
		return err
	}
	b.len += n
	return nil
}
