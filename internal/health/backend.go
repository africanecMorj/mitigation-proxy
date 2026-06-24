package health

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/pkg"
)

type BackendState int32

const (
	Healthy BackendState = iota
	Suspect
	Draining
	Unhealthy
	Recovering
	Removed
)

type BackendPool struct {
	backend *Backend
	idle    chan int
}

type Backend struct {
	Address string

	State           atomic.Int32
	LastStateChange atomic.Int64

	Weight        atomic.Int64
	currentWeight atomic.Int64

	Tau float64

	ActiveConnections atomic.Int64

	TotalFailures  atomic.Uint64
	TotalSuccesses atomic.Uint64

	ConsecutiveFailures  atomic.Int64
	ConsecutiveSuccesses atomic.Int64

	DrainStartedAt atomic.Int64

	SockAddr unix.Sockaddr
	Family   int

	// latency metrics
	ewmaMu         sync.Mutex
	ewma           float64
	lastEWMAUpdate time.Time

	ttfb atomic.Int64

	Requests atomic.Uint64

	TotalLatency atomic.Uint64

	Ctx    context.Context
	cancel context.CancelFunc

	// drain protection
	draining atomic.Bool

	//TODO: pool connection system:
	pool *BackendPool
}

func NewBackend(
	address string,
	tau float64,
	weight int64,
	ctx context.Context,
	cancel context.CancelFunc,
) (*Backend, error) {
	family, sa, err := pkg.ResolveSockaddr(address)

	if err != nil {
		return nil, err
	}

	b := Backend{
		Address:  address,
		Tau:      tau,
		Family:   family,
		SockAddr: sa,
		Ctx:      ctx,
		cancel:   cancel,
	}
	b.SetWeight(weight)

	return &b, nil

}
