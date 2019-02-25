package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/c9s/goprocinfo/linux"
)

type statInfo struct {
	mtx        sync.Mutex
	stamp      time.Time
	accMsgLen  uint64
	nbrMsg     uint64
	cpuTicks   uint64
	idleTicks  uint64
	totalTicks uint64
}

var stats = statInfo{stamp: time.Now()}

func statStart(period time.Duration) {
	stats.stamp = time.Now()
	cpuTicks, idleTicks, totalTicks := getCPUStats()
	stats.cpuTicks = cpuTicks
	stats.idleTicks = idleTicks
	stats.totalTicks = totalTicks
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
		cpuTicks, idleTicks, totalTicks := getCPUStats()
		cpu := 100 * float64(cpuTicks-stats.cpuTicks) / float64(totalTicks-stats.totalTicks)
		idle := 100 * float64(idleTicks-stats.idleTicks) / float64(totalTicks-stats.totalTicks)
		stats.cpuTicks = cpuTicks
		stats.idleTicks = idleTicks
		stats.totalTicks = totalTicks
		log.Printf("%.3f usec/msg, %.3f B/msg, %.3f kHz, %.3f MB/s, cpu: %.1f%% idle: %.1f%%\n",
			usmsg, mLen, rate/1000, mbs, cpu, idle)
	}
}

var pidStr = strconv.Itoa(os.Getpid())

func getCPUStats() (cpuTicks, idleTicks, totalTicks uint64) {
	cpuStat, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return
	}
	cpuStatStr := string(cpuStat)
	cpuStatStr = cpuStatStr[:strings.IndexByte(cpuStatStr, '\n')]
	fields := strings.Fields(cpuStatStr)
	for i := 1; i < len(fields); i++ {
		val, _ := strconv.ParseUint(fields[i], 10, 64)
		if i == 4 {
			idleTicks = val
		}
		totalTicks += val
	}

	pStats, err := linux.ReadProcessStat("/proc/" + pidStr + "/stat")
	if err != nil {
		return
	}
	cpuTicks = pStats.Utime + pStats.Stime
	return
}
