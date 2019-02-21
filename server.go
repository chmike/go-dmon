package main

import (
	"crypto/rand"
	"crypto/tls"
	"log"
	"net"
	"time"

	"github.com/chmike/go-dmon/dmon"
	"github.com/pkg/errors"
)

const ackByte byte = 0xA5

type msgInfo struct {
	len int
	msg dmon.Msg
}

func runAsServer() {
	log.SetPrefix("server ")

	msgs := make(chan msgInfo, 1000)
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

func handleClient(conn net.Conn, msgs chan msgInfo) {
	var (
		ack   = []byte{ackByte}
		rConn MsgReader
		err   error
	)
	defer conn.Close()

	wConn := NewBufWriter(conn, *bufLenFlag, time.Duration(*bufPeriodFlag)*time.Millisecond)
	switch *msgCodecFlag {
	case "json":
		rConn = NewJSONReader(NewBufReader(conn, *bufLenFlag))
	case "binary":
		rConn = NewBinaryReader(NewBufReader(conn, *bufLenFlag))
	}

	for {
		var m msgInfo

		m.len, err = rConn.Read(&m.msg)
		if err != nil {
			log.Println(errors.Wrapf(err, "could not receive message"))
			break
		}

		msgs <- m

		_, err = wConn.Write(ack)
		if err != nil {
			log.Println("send error:", err)
			break
		}
	}
	log.Println("conn closed")
}
