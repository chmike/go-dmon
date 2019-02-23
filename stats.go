package main

import (
	"log"
	"sync"
	"sync/atomic"
	"time"
)

type statInfo struct {
	mtx       sync.Mutex
	stamp     time.Time
	accMsgLen uint64
	nbrMsg    uint64
}

var stats = statInfo{stamp: time.Now()}

func statStart(period time.Duration) {
	stats.stamp = time.Now()
	go statDisplay(period)
}

func statUpdate(msgLen int) {
	atomic.AddUint64(&stats.accMsgLen, uint64(msgLen))
	atomic.AddUint64(&stats.nbrMsg, 1)
}

func statDisplay(period time.Duration) {
	for {
		time.Sleep(period)

		accMsgLen := atomic.SwapUint64(&stats.accMsgLen, 0)
		nbrMsg := atomic.SwapUint64(&stats.nbrMsg, 0)
		delay := time.Since(stats.stamp)
		stats.stamp = time.Now()
		mbs := float64(accMsgLen) / (1000000. * delay.Seconds())
		rate := float64(nbrMsg) / delay.Seconds()
		usmsg := 1000000. / rate
		mLen := float64(accMsgLen) / float64(nbrMsg)
		if rate == 0. {
			usmsg = 0
			mLen = 0
		}
		log.Printf("%.3f usec/msg, %.3f B/msg, %.3f kHz, %.3f MB/s\n", usmsg, mLen, rate/1000, mbs)
	}
}
