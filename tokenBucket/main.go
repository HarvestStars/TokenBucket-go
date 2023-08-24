package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var (
	connMap     map[int]bool
	mapLock     sync.Mutex
	accpetCount int
	reqCount    int
)

type TokenBucket struct {
	tokens        int
	tokenCapacity int
	refillRate    time.Duration
	lastRefill    time.Time
	tokenLock     sync.Mutex
}

func NewTokenBucket(tokenCapacity int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:        tokenCapacity,
		tokenCapacity: tokenCapacity,
		refillRate:    refillRate,
		lastRefill:    time.Now(),
	}
}

func (tb *TokenBucket) refillTokens() {
	now := time.Now()
	timeElapsed := now.Sub(tb.lastRefill)
	tokensToAdd := int(timeElapsed.Seconds() / tb.refillRate.Seconds())

	if tokensToAdd > 0 {
		tb.tokens = min(tb.tokens+tokensToAdd, tb.tokenCapacity)
		tb.lastRefill = now
	}
}

func (tb *TokenBucket) Consume(tokens int) bool {
	tb.tokenLock.Lock()
	defer tb.tokenLock.Unlock()

	tb.refillTokens()

	if tb.tokens >= tokens {
		tb.tokens -= tokens
		return true
	}

	return false
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func main() {
	connMap = make(map[int]bool, 2000)
	tokenBucket := NewTokenBucket(500, time.Millisecond*10) // generate 100 tokens per second
	for i := 0; i < 2000; i++ {
		if i == 1000 {
			time.Sleep(5 * time.Second)
		}
		go func() {
			//new(rand.NewSource(100))
			randomNumber := rand.Intn(1000)
			time.Sleep(time.Millisecond * time.Duration(randomNumber))
			if tokenBucket.Consume(1) {
				mapLock.Lock()
				connMap[i] = true
				reqCount++
				accpetCount++
				fmt.Println("Request processed.")
				mapLock.Unlock()
			} else {
				mapLock.Lock()
				reqCount++
				connMap[i] = false
				fmt.Println("Request dropped.")
				mapLock.Unlock()
			}

		}()
	}
	c := make(chan os.Signal, 5)
	signal.Notify(c, os.Interrupt, os.Kill, syscall.SIGABRT, syscall.SIGTERM)
	for {
		select {
		case <-c:
			return
		default:
			time.Sleep(time.Duration(2) * time.Second)
			mapLock.Lock()
			fmt.Println(fmt.Sprintf("accept: %v / %v", accpetCount, reqCount))
			mapLock.Unlock()
		}
	}
}
