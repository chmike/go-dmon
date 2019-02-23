package main

import (
	"database/sql"
	"log"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

var mysqlCredentials = "dmon:4dmonTest!@/dmon?charset=utf8"

func database(msgs chan msgInfo) {
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

	statStart(time.Duration(*periodFlag) * time.Second)

	if *dbFlag {
		for m := range msgs {
			_, err = db.Exec("INSERT dmon SET stamp=?,level=?,system=?,component=?,message=?",
				m.msg.Stamp, m.msg.Level, m.msg.System, m.msg.Component, m.msg.Message)
			if err != nil {
				log.Println("ERROR:", err, ": ignoring entry")
				continue
			}
			statUpdate(m.len)
		}
	} else {
		for m := range msgs {
			statUpdate(m.len)
		}
	}
}
