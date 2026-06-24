package balancers

import (
	"sync"
	"time"
	"slices"
	"math/rand"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

type BaseBalancer struct {
	backends []*health.Backend
	mu       sync.RWMutex
}

func NewBaseBalancer(
	backends []*health.Backend,
) BaseBalancer {
	return BaseBalancer{
		backends: slices.Clone(backends),
	}
}


func (b *BaseBalancer) AddBackend(
	backend *health.Backend,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.backends = append(
		b.backends,
		backend,
	)
}

func (b *BaseBalancer) RemoveBackend(
	address string,
) {
	b.mu.Lock()
	defer b.mu.Unlock()

	filtered := b.backends[:0]

	for _, backend := range b.backends {

		if backend.Address != address {
			filtered = append(
				filtered,
				backend,
			)
		}
	}

	b.backends = filtered
}

func (b *BaseBalancer) Backends() []*health.Backend {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return slices.Clone(b.backends)
}

func (b *BaseBalancer) StartHealthChecks(
    interval time.Duration,
) {
    for _, backend := range b.Backends() {
        go func(backend *health.Backend) {

            if interval > 0 {
                jitter := time.Duration(
                    rand.Int63n(int64(interval / 5)),
                )

                select {
                case <-backend.Ctx.Done():
                    return
                case <-time.After(jitter):
                }
            }

            backend.HealthCheck()

            ticker := time.NewTicker(interval)
            defer ticker.Stop()

            for {
                select {
                case <-backend.Ctx.Done():
                    return

                case <-ticker.C:
                    backend.HealthCheck()
                }
            }

        }(backend)
    }
}

