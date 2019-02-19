package main

import (
	"log"
	"time"
)

type statVal struct {
	val  []float64
	nbr  int
	pos  int
	vAve float64
	vVar float64
	wSz  float64
}

func (v *statVal) init(windowSize int) {
	v.val = make([]float64, windowSize)
	v.nbr = 0
	v.pos = 0
	v.vAve = 0
	v.vVar = 0
	v.wSz = float64(windowSize)
}

func (v *statVal) update(val float64) {
	if v.nbr < len(v.val) {
		vDiff := val - v.vAve
		n := float64(v.nbr + 1)
		v.vVar += (((n-1)/n)*vDiff*vDiff - v.vVar) / n
		v.vAve += vDiff / n
		v.val[v.nbr] = val
		v.nbr++
		return
	}

	vDiff := v.val[v.pos] - v.vAve
	v.vVar -= (v.wSz/(v.wSz-1)*vDiff*vDiff - v.vVar) / (v.wSz - 1)
	v.vAve -= vDiff / (v.wSz - 1)
	vDiff = val - v.vAve
	v.vVar += (((v.wSz-1)/v.wSz)*vDiff*vDiff - v.vVar) / v.wSz
	v.vAve += vDiff / v.wSz
	v.val[v.pos] = val
	v.pos++
	if v.pos == len(v.val) {
		v.pos = 0
	}
}

func (v *statVal) average() float64 {
	return v.vAve
}

func (v *statVal) variance() float64 {
	return v.vVar
}

type statInfo struct {
	max        int
	cnt        int
	start      time.Time
	dataLenSum int
	delay      statVal
	dataLen    statVal
}

// newStat return a stat object were period is the number of calls
// to update() required to update the stats.
func newStats(period int, windowSize int) *statInfo {
	s := &statInfo{
		max:   period,
		start: time.Now(),
	}
	s.delay.init(windowSize)
	s.dataLen.init(windowSize)
	return s
}

// update update stats
func (s *statInfo) update(dataLen int) {
	s.cnt++
	if s.cnt != s.max {
		s.dataLenSum += dataLen
		return
	}
	println("coucou")
	delay := time.Since(s.start).Seconds() / float64(s.max)
	s.start = time.Now()
	s.delay.update(delay / float64(s.max))
	s.dataLen.update(float64(s.dataLenSum))
	s.dataLenSum = 0

	log.Printf("%.3f usec/msg, %.3f Hz, %.3f MB/s\n",
		s.delay.average()*1000000, 1./s.delay.average(),
		s.dataLen.average()/(s.delay.average()*1000000.))
}
