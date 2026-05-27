package limiter

import (
	"sync"

	"golang.org/x/time/rate"
)

const (
	LISTEN_ADDR = ":9000"
	BACKEND_ADDR = "127.0.0.1:8080"

	MAX_CONNECTIONS = 10000
	RATE_LIMIT_RPS = 20
	RATE_LIMIT_BURST = 40
)

var (
	activeConnections int
	connMutex         sync.Mutex

	ipLimiters sync.Map
)

func GetLimiter(ip string) *rate.Limiter {
	limiter, exists := ipLimiters.Load(ip)
	if exists {
		return limiter.(*rate.Limiter)
	}

	newLimiter := rate.NewLimiter(RATE_LIMIT_RPS, RATE_LIMIT_BURST)
	ipLimiters.Store(ip, newLimiter)

	return newLimiter
}

func IncrementConnections() bool {
	connMutex.Lock()
	defer connMutex.Unlock()

	if activeConnections >= MAX_CONNECTIONS {
		return false
	}

	activeConnections++
	return true
}

func DecrementConnections() {
	connMutex.Lock()
	activeConnections--
	connMutex.Unlock()
}