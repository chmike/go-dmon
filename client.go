package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/chmike/go-dmon/dmon"
	_ "github.com/go-sql-driver/mysql"
)

var serverDNSNameCheck = false

func runAsClient() {
	log.SetPrefix("client ")
	log.Println("target:", *addressFlag)

	var (
		conn      net.Conn
		err       error
		msgWriter dmon.MsgWriter
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

	bufWriter := NewBufWriter(conn, *bufLenFlag, time.Duration(*bufPeriodFlag)*time.Millisecond)
	switch *msgCodecFlag {
	case "json":
		msgWriter = dmon.NewJSONWriter(bufWriter)
	case "binary":
		msgWriter = dmon.NewBinaryWriter(bufWriter)
	}
	reqAcks := make(chan struct{}, 5000)
	go getAcks(NewBufReader(conn, *bufLenFlag), reqAcks)
	statStart(time.Duration(*periodFlag) * time.Second)
	for {
		m := dmon.Msg{
			Stamp:     time.Now().UTC(),
			Level:     "info",
			System:    "dmon",
			Component: "test",
			Message:   "no problem",
		}
		n, err := msgWriter.Write(&m)
		if err != nil {
			log.Fatalf("msg send: %v", err)
		}
		statUpdate(n)
		reqAcks <- struct{}{}
	}
}

func getAcks(bufReader *BufReader, reqAcks chan struct{}) {
	for range reqAcks {
		b, err := bufReader.ReadByte()
		if err != nil {
			if err == io.EOF {
				log.Printf("close conn")
				os.Exit(0)
			}
			log.Fatal(err)
		}
		if b != ackCode {
			log.Fatalf("expected %+X, got %+X", ackCode, b)
		}
	}
}
