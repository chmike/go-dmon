package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"time"

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

	clientCert, err := tls.LoadX509KeyPair(clientCRTFilename, clientKeyFilename)
	config := tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		InsecureSkipVerify: !serverDNSNameCheck,
	}
	conn, err := tls.Dial("tcp", *addressFlag, &config)
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
	for {
		m := monEntry{
			Stamp:     time.Now(),
			Level:     "info",
			System:    "dmon",
			Component: "test",
			Message:   "no problem",
		}

		data, err := json.Marshal(&m)
		msg := fmt.Sprintf("%d:%s", len(data), string(data))
		n, err := io.WriteString(conn, msg)
		if err != nil {
			log.Fatalln("send error:", err)
		}
		if lastCount-prevCount == statCount {
			duration := time.Since(prevTime)
			microSec := duration.Seconds() * 1000000 / float64(statCount)
			log.Printf("send '%s' (%d bytes)", msg, n)
			log.Printf("%.3f usec/msg, %.3f Hz\n", microSec, 1000000/microSec)
			prevCount = lastCount
			prevTime = time.Now()
		}
		lastCount++

		ack <- struct{}{}
	}
}

func getAcks(conn net.Conn, ack chan struct{}) {
	buf := make([]byte, 3)
	defer conn.Close()

	for {
		// wait for ack request
		_ = <-ack

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
