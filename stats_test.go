package main

import (
	"math"
	"math/rand"
	"testing"
)

func TestStats(t *testing.T) {
	const wSz = 10000
	const epsilon = 1e-10

	var (
		rMean   float64
		rStdDev float64
		stats   statVal
		val     = make([]float64, wSz)
	)
	stats.init(wSz)

	rMean = 0
	rStdDev = 1
	for i := range val {
		val[i] = rand.NormFloat64()*rStdDev + rMean
		stats.update(val[i])
	}
	vAve, vVar := computeStats(val)
	if math.Abs(vAve-stats.average()) > epsilon {
		t.Errorf("average mismatch: ref=%g stats=%g diff=%g",
			vAve, stats.average(), math.Abs(vAve-stats.average()))
	}
	if math.Abs(vVar-stats.variance()) > epsilon {
		t.Errorf("variance mismatch: ref=%g stats=%g diff=%g",
			vVar, stats.variance(), math.Abs(vVar-stats.variance()))
	}

	rMean = 100
	rStdDev = 2
	for i := range val {
		val[i] = rand.NormFloat64()*rStdDev + rMean
		stats.update(val[i])
	}
	vAve, vVar = computeStats(val)
	if math.Abs(vAve-stats.average()) > epsilon {
		t.Errorf("average mismatch: ref=%g stats=%g diff=%g",
			vAve, stats.average(), math.Abs(vAve-stats.average()))
	}
	if math.Abs(vVar-stats.variance()) > epsilon {
		t.Errorf("variance mismatch: ref=%g stats=%g diff=%g",
			vVar, stats.variance(), math.Abs(vVar-stats.variance()))
	}

	rMean = -100
	rStdDev = 0.2
	for i := range val {
		val[i] = rand.NormFloat64()*rStdDev + rMean
		stats.update(val[i])
	}
	vAve, vVar = computeStats(val)
	if math.Abs(vAve-stats.average()) > epsilon {
		t.Errorf("average mismatch: ref=%g stats=%g diff=%g",
			vAve, stats.average(), math.Abs(vAve-stats.average()))
	}
	if math.Abs(vVar-stats.variance()) > epsilon {
		t.Errorf("variance mismatch: ref=%g stats=%g diff=%g",
			vVar, stats.variance(), math.Abs(vVar-stats.variance()))
	}

}

func computeStats(val []float64) (vAve float64, vVar float64) {
	for i := 0; i < len(val); i++ {
		n := float64(i + 1)
		vDiff := val[i] - vAve
		vVar += (((n-1)/n)*vDiff*vDiff - vVar) / n
		vAve += vDiff / n
	}
	return
}
