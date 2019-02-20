package main

import (
	"crypto/tls"
	"encoding/binary"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/chmike/go-dmon/dmon"
	_ "github.com/go-sql-driver/mysql"
)

var (
	serverDNSNameCheck = false
)

func runAsClient() {
	log.SetPrefix("client ")
	log.Println("target:", *addressFlag)

	var (
		conn    io.ReadWriteCloser
		netConn net.Conn
		err     error
		id      int64
	)
	if *tlsFlag {
		var clientCert tls.Certificate
		clientCert, err = tls.LoadX509KeyPair(clientCRTFilename, clientKeyFilename)
		if err != nil {
			log.Fatalf("could not load X509 certificate: %v", err)
		}
		config := tls.Config{
			Certificates:       []tls.Certificate{clientCert},
			InsecureSkipVerify: !serverDNSNameCheck,
		}
		netConn, err = tls.Dial("tcp", *addressFlag, &config)
	} else {
		netConn, err = net.Dial("tcp", *addressFlag)
		if *bufLenFlag == 0 {
			netConn.(*net.TCPConn).SetNoDelay(false)
		}
	}
	if err != nil {
		log.Fatal(err)
	}
	log.Println("connected to:", netConn.RemoteAddr())
	if *bufLenFlag != 0 {
		conn = newSender(netConn, time.Duration(*bufPeriodFlag)*time.Millisecond, *bufLenFlag)
	} else {
		conn = netConn
	}
	defer conn.Close()

	ack := make(chan struct{}, 5000)
	go getAcks(conn, ack)

	statStart(time.Duration(*periodFlag) * time.Second)

	buf := make([]byte, 512)
	for {
		id++
		m := dmon.Msg{
			ID:        id,
			Stamp:     time.Now().UTC(),
			Level:     "info",
			System:    "dmon",
			Component: "test",
			Message:   "no problem",
		}

		var (
			data []byte
			err  error
		)
		switch msgCodec {
		case JSON:
			data, err = m.MarshalJSON()
		case BINARY:
			data, err = m.MarshalBinary()
		}
		if err != nil {
			log.Fatalf("could not encode message: %v", err)
		}
		buf = buf[:4]
		binary.LittleEndian.PutUint32(buf, uint32(len(data)))
		buf = append(buf, data...)

		n, err := conn.Write(buf)
		if err != nil {
			log.Fatalln("send error:", err)
		}
		if n != len(buf) {
			println("truncated write")
		}

		statUpdate(n)

		ack <- struct{}{}
	}
}

func getAcks(conn io.ReadWriteCloser, ack chan struct{}) {
	buf := make([]byte, 3)
	defer conn.Close()

	for range ack {
		// do read ack from connection
		_, err := io.ReadFull(conn, buf)
		if err != nil {
			if err == io.EOF {
				log.Printf("close conn")
				os.Exit(0)
			}
			log.Fatal(err)
		}
		if string(buf) != "ack" {
			log.Fatalf("expected \"ack\", got %s", string(buf))
		}
	}
}
