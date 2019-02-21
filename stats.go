package main

import (
	"log"
	"sync"
	"time"
)

type statInfo struct {
	mtx       sync.Mutex
	accMsgLen uint64
	nbrMsg    uint64
	stamp     time.Time
}

var stats = statInfo{stamp: time.Now()}

func statStart(period time.Duration) {
	stats.stamp = time.Now()
	go statDisplay(period)
}

func statUpdate(msgLen int) {
	stats.mtx.Lock()
	stats.accMsgLen += uint64(msgLen)
	stats.nbrMsg++
	stats.mtx.Unlock()
}

func statDisplay(period time.Duration) {
	for {
		time.Sleep(period)

		stats.mtx.Lock()
		accMsgLen := stats.accMsgLen
		stats.accMsgLen = 0
		nbrMsg := stats.nbrMsg
		stats.nbrMsg = 0
		stats.mtx.Unlock()
		delay := time.Since(stats.stamp)
		stats.stamp = time.Now()
		mbs := float64(accMsgLen) / (1000000. * delay.Seconds())
		rate := float64(nbrMsg) / delay.Seconds()
		usmsg := 1000000. / rate
		if rate == 0. {
			usmsg = 0
		}
		log.Printf("%.3f usec/msg, %.3f kHz, %.3f MB/s\n", usmsg, rate/1000, mbs)
	}
}
