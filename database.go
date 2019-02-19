package main

import (
	"database/sql"
	"log"

	"github.com/chmike/go-dmon/dmon"
	_ "github.com/go-sql-driver/mysql"
)

var mysqlCredentials = "dmon:4dmonTest!@/dmon?charset=utf8"

func database(msgs chan dmon.Msg) {
	var (
		db  *sql.DB
		err error
	)

	if *dbFlag {
		db, err = sql.Open("mysql", mysqlCredentials)
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
	}

	stats := newStats(statUpdatePeriod, statWindowSize)
	for {
		m := <-msgs

		if *dbFlag {
			_, err = db.Exec("INSERT dmon SET stamp=?,level=?,system=?,component=?,message=?",
				m.Stamp, m.Level, m.System, m.Component, m.Message)
			if err != nil {
				log.Println("ERROR:", err, ": ignoring entry")
				continue
			}
		}

		stats.update(130)
	}
}
