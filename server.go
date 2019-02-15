package main

import (
	"crypto/rand"
	"crypto/tls"
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mysqlCredentials = "dmon:4dmonTest!@/dmon?charset=utf8"
)

func runAsServer() {
	log.SetPrefix("server ")

	// make sure the database table is created
	db, err := sql.Open("mysql", mysqlCredentials)
	if err != nil {
		log.Fatalln(err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS mon (
			mid BIGINT NOT NULL AUTO_INCREMENT,
			stamp DATETIME(6) NOT NULL,
			level VARCHAR(5) NOT NULL,
			system VARCHAR(128) NOT NULL,
			component VARCHAR(64) NOT NULL,
			message VARCHAR(256) NOT NULL,
			PRIMARY KEY (mid)
		) ENGINE=INNODB
	`)
	db.Close()
	if err != nil {
		log.Fatalln(err)
	}

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
		go handleClient(conn)
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	db, err := sql.Open("mysql", mysqlCredentials)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	b := newBuffer()
	var m monEntry
	for {
		err := recvMsg(&m, b, conn)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatalln("error:", err)
		}
		_, err = db.Exec("INSERT mon SET stamp=?,level=?,system=?,component=?,message=?",
			m.Stamp, m.Level, m.System, m.Component, m.Message)
		if err != nil {
			log.Println("ERROR:", err, ": ignoring entry")
			continue
		}
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
