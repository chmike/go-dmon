package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"log"
	"net"
	"time"

	"github.com/chmike/go-dmon/dmon"
)

const ackCode byte = 0xA5

type msgInfo struct {
	len int
	msg dmon.Msg
}

func runAsServer() {
	log.SetPrefix("server ")

	msgs := make(chan msgInfo, *dbBufLenFlag*10)
	defer close(msgs)
	go database(msgs)

	var (
		listener net.Listener
		err      error
	)
	// listen for a connection
	if *tlsFlag {
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
			log.Fatalln("accept error:", err)
		}
		go handleClient(conn, msgs)
	}
}

func handleClient(conn net.Conn, msgs chan msgInfo) {
	var (
		hdr [8]byte
		err error
		n   int
		m   msgInfo
	)
	defer conn.Close()

	for {
		// decode and check message header
		conn.SetReadDeadline(time.Now().Add(timeOutDelay))
		_, err = conn.Read(hdr[:])
		if err != nil {
			log.Println("recv message header error:", err)
			return
		}
		if string(hdr[:4]) != "DMON" {
			log.Printf("recv header error: expected 'DMON', got '%s' (0x%s)", string(hdr[:4]), hex.EncodeToString(hdr[:4]))
			return
		}
		dataLen := int(binary.LittleEndian.Uint32(hdr[4:]))

		// decode message data
		conn.SetReadDeadline(time.Now().Add(timeOutDelay))
		buf := make([]byte, dataLen)
		_, err = conn.Read(buf)
		if err != nil {
			log.Println("recv message payload error:", err)
			return
		}
		if *jsonFlag {
			if *msgFlag {
				log.Println("recv:", string(buf))
			}
			err = m.msg.JSONDecode(buf)
		} else {
			err = m.msg.BinaryDecode(buf)
		}
		if err != nil {
			log.Println("decode message error:", err)
			return
		}

		// send acknowledgment
		var b = [1]byte{ackCode}
		conn.SetWriteDeadline(time.Now().Add(15 * time.Second))
		n, err = conn.Write(b[:])
		if err != nil {
			log.Println("send acknowledgment error:", err)
			return
		}
		if n != 1 {
			log.Printf("send acknowledgment error: expected 1 byte send, got %d", n)
			return
		}

		// pass message to database writer
		m.len = dataLen + len(hdr)
		msgs <- m
	}
}
