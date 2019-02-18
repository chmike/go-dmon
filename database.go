package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var (
	mysqlCredentials = "dmon:4dmonTest!@/dmon?charset=utf8"
	statCount        = 1000
)

func database(monEntryChan chan *monEntry) {
	db, err := sql.Open("mysql", mysqlCredentials)
	if err != nil {
		log.Fatalln(err)
	}
	defer db.Close()

	_, err = db.Exec(`
			CREATE TABLE IF NOT EXISTS dmon (
				mid BIGINT NOT NULL AUTO_INCREMENT,
				stamp DATETIME(6) NOT NULL,
				level VARCHAR(5) NOT NULL,
				system VARCHAR(128) NOT NULL,
				component VARCHAR(64) NOT NULL,
				message VARCHAR(256) NOT NULL,
				PRIMARY KEY (mid)
			) ENGINE=INNODB
		`)
	if err != nil {
		log.Fatalln(err)
	}

	prevTime := time.Now()
	prevCount := 0
	lastCount := 0
	for {
		m := <-monEntryChan

		_, err = db.Exec("INSERT dmon SET stamp=?,level=?,system=?,component=?,message=?",
			m.Stamp, m.Level, m.System, m.Component, m.Message)
		if err != nil {
			log.Println("ERROR:", err, ": ignoring entry")
			continue
		}

		monEntryPool.Put(m)

		if lastCount-prevCount == statCount {
			duration := time.Since(prevTime)
			log.Printf("%f sec for %d messages\n", duration.Seconds(), statCount)
			prevCount = lastCount
			prevTime = time.Now()
		}
		lastCount++
	}
}
