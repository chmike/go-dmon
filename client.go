package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log"
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
	buf := make([]byte, 3)
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
		log.Printf("send '%s' (%d bytes)", msg, n)

		n, err = io.ReadFull(conn, buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		if string(buf) != "ack" {
			log.Fatalf("expected \"ack\", got %s", string(buf))
		}
	}
	log.Printf("close conn")
}
