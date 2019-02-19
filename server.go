package main

import (
	"crypto/rand"
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"

	"github.com/chmike/go-dmon/dmon"
	"github.com/pkg/errors"
)

func runAsServer() {
	log.SetPrefix("server ")

	msgs := make(chan dmon.Msg, 1000)
	defer close(msgs)
	go database(msgs)

	var (
		listener net.Listener
		err      error
	)
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
		go handleClient(conn, msgs)
	}
}

func handleClient(conn net.Conn, msgs chan dmon.Msg) {
	defer conn.Close()

	var (
		hdr [4]byte
		ack = []byte("ack")
		buf = make([]byte, 512)
	)

	for {
		var m dmon.Msg

		_, err := io.ReadFull(conn, hdr[:])
		if err != nil {
			log.Fatal(errors.Wrapf(err, "could not read message header"))
		}
		n := binary.LittleEndian.Uint32(hdr[:])
		if len(buf) < int(n) {
			buf = make([]byte, n)
		} else {
			buf = buf[:n]
		}
		_, err = io.ReadFull(conn, buf)
		if err != nil {
			log.Fatal(errors.Wrapf(err, "could not read message payload"))
		}

		switch msgCodec {
		case JSON:
			err = m.UnmarshalJSON(buf)
		case BINARY:
			err = m.UnmarshalBinary(buf)
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("error:", err)
		}

		msgs <- m

		_, err = conn.Write(ack)
		if err != nil {
			log.Println("send error:", err)
			break
		}
	}
	log.Println("conn closed")
}
