package ratelimit

import (
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/time/rate"
)

const (
	MaxConnections = 10_000

	RateLimitRPS   = 20
	RateLimitBurst = 40

	LimiterTTL      = time.Hour
	CleanupInterval = 10 * time.Minute
)

var (
	activeConnections atomic.Int64
	ipLimiters        sync.Map // map[string]*IPLimiter
)

type IPLimiter struct {
	Limiter  *rate.Limiter
	LastSeen atomic.Int64 // unix timestamp
}

func Init() {
	go StartLimiterCleanup()
}

func GetLimiter(ip string) *rate.Limiter {
	now := time.Now().Unix()

	// Fast path
	if v, ok := ipLimiters.Load(ip); ok {
		entry := v.(*IPLimiter)
		entry.LastSeen.Store(now)
		return entry.Limiter
	}

	// Create candidate
	newEntry := &IPLimiter{
		Limiter: rate.NewLimiter(
			rate.Limit(RateLimitRPS),
			RateLimitBurst,
		),
	}
	newEntry.LastSeen.Store(now)

	// Atomically insert if absent
	actual, _ := ipLimiters.LoadOrStore(ip, newEntry)

	entry := actual.(*IPLimiter)
	entry.LastSeen.Store(now)

	return entry.Limiter
}

func IncrementConnections() bool {
	for {
		current := activeConnections.Load()

		if current >= MaxConnections {
			return false
		}

		if activeConnections.CompareAndSwap(current, current+1) {
			return true
		}
	}
}

func DecrementConnections() {
	for {
		current := activeConnections.Load()

		if current <= 0 {
			return
		}

		if activeConnections.CompareAndSwap(current, current-1) {
			return
		}
	}
}

func ActiveConnectionCount() int64 {
	return activeConnections.Load()
}

func LimiterCount() int {
	count := 0

	ipLimiters.Range(func(_, _ any) bool {
		count++
		return true
	})

	return count
}

func StartLimiterCleanup() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		CleanupLimiters()
	}
}

func CleanupLimiters() {
	cutoff := time.Now().Add(-LimiterTTL).Unix()

	ipLimiters.Range(func(key, value any) bool {
		entry := value.(*IPLimiter)

		if entry.LastSeen.Load() < cutoff {
			ipLimiters.Delete(key)
		}

		return true
	})
}