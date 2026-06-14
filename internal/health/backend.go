package health

import (
	"time"
	"sync"
	"sync/atomic"

	"golang.org/x/sys/unix"
)

type BackendState int32

const (
	Healthy BackendState = iota
	Suspect
	Draining
	Unhealthy
	Recovering
)

type BackendPool struct {
	backend *Backend
	idle    chan int
}

type Backend struct {
	Address string

	State           atomic.Int32
	LastStateChange atomic.Int64

	Weight atomic.Int64
	currentWeight atomic.Int64

	Tau           float64

	ActiveConnections atomic.Int64

	ConsecutiveFailures  atomic.Int64
	ConsecutiveSuccesses atomic.Int64

	DrainStartedAt atomic.Int64

	SockAddr unix.Sockaddr
	Family   int

	// latency metrics
	ewmaMu         sync.Mutex
	ewma           float64
	lastEWMAUpdate time.Time

	ttfb 		   atomic.Int64

	AvgLatency atomic.Int64

	// drain protection
	draining atomic.Bool

	//TODO: pool connection system:
	pool *BackendPool
}
