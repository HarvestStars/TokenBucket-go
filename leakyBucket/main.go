package main

import (
	"fmt"
	"sync"
	"time"
)

type Funnel struct {
	capacity      int
	leakingRate   int
	availableData int
	lastLeak      time.Time
	mu            sync.Mutex
}

func NewFunnel(capacity, leakingRate int) *Funnel {
	return &Funnel{
		capacity:      capacity,
		leakingRate:   leakingRate,
		availableData: 0,
		lastLeak:      time.Now(),
	}
}

func (f *Funnel) leak() {
	now := time.Now()
	elapsed := now.Sub(f.lastLeak)
	leakedData := int(elapsed.Seconds()) * f.leakingRate

	if leakedData > 0 {
		f.availableData = max(f.availableData-leakedData, 0)
		f.lastLeak = now
	}
}

func (f *Funnel) Pour(data int) bool {
	f.mu.Lock()
	defer f.mu.Unlock()

	f.leak()

	if f.availableData+data <= f.capacity {
		f.availableData += data
		return true
	}

	return false
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func main() {
	funnel := NewFunnel(10, 2) // Allow 2 units of data per second

	for i := 0; i < 15; i++ {
		if funnel.Pour(1) {
			fmt.Println("Data passed through the funnel.")
		} else {
			fmt.Println("Data blocked by the funnel.")
		}
		time.Sleep(200 * time.Millisecond) // Simulate incoming data
	}
}
