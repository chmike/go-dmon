package main

import (
	"crypto/tls"
	"encoding/binary"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

	"github.com/chmike/go-dmon/dmon"
	_ "github.com/go-sql-driver/mysql"
)

var (
	serverDNSNameCheck = false
)

func runAsClient() {
	log.SetPrefix("client ")
	if strings.HasPrefix(*addressFlag, "0.0.0.0") {
		log.Fatal("invalid address: ", *addressFlag)
	}
	log.Println("target:", *addressFlag)

	var (
		conn net.Conn
		err  error
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
		conn, err = tls.Dial("tcp", *addressFlag, &config)
	} else {
		conn, err = net.Dial("tcp", *addressFlag)
	}
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	log.Println("connected to:", conn.RemoteAddr())

	ack := make(chan struct{}, 5000)
	go getAcks(conn, ack)

	prevTime := time.Now()
	prevCount := 0
	lastCount := 0
	buf := make([]byte, 4, 512)
	for {
		m := dmon.Msg{
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
		binary.LittleEndian.PutUint32(buf[:4], uint32(len(data)))
		if len(data) > len(buf)-4 {
			buf = append(buf, make([]byte, len(data))...)
		}
		copy(buf[4:], data) // FIXME(sbinet): segregate writes b/w hdr/payload

		n, err := conn.Write(buf[:4+len(data)])
		if err != nil {
			log.Fatalln("send error:", err)
		}
		if lastCount-prevCount == statCount {
			duration := time.Since(prevTime)
			microSec := (duration.Seconds() * 1000000) / float64(statCount)
			str := string(buf[:n])
			if msgCodec == BINARY {
				str = fmt.Sprintf("%q", str)
			}
			log.Printf("send '%s' (%d bytes)", str, n)
			log.Printf("%.3f usec/msg, %.3f Hz\n", microSec, 1000000/microSec)
			prevCount = lastCount
			prevTime = time.Now()
		}
		lastCount++

		ack <- struct{}{}
		buf = buf[:4]
	}
}

func getAcks(conn net.Conn, ack chan struct{}) {
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
