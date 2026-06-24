package runtime

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

type Runtime struct {
	mu       sync.RWMutex
	clusters map[string]balancers.Balancer
	loops    map[string]*transport.Transport
	wg       sync.WaitGroup
}

func New() *Runtime {
	return &Runtime{
		loops:    make(map[string]*transport.Transport),
		clusters: make(map[string]balancers.Balancer),
	}
}

func (rt *Runtime) Register(
	addr string,
	tr *transport.Transport,
) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.loops[addr] = tr
}

func (rt *Runtime) RegisterClusters(
	clusters map[string]balancers.Balancer,
) {
	rt.mu.Lock()
	defer rt.mu.Unlock()

	rt.clusters = clusters
}

func (rt *Runtime) Get(
	addr string,
) (*transport.Transport, bool) {

	rt.mu.RLock()
	defer rt.mu.RUnlock()

	tr, ok := rt.loops[addr]

	return tr, ok
}

func (rt *Runtime) reload(
	addr string,
	picker transport.BackendPicker,
	ins inspector.Inspector,
) error {
	tr, ok := rt.Get(addr)
	if !ok {
		return fmt.Errorf(
			"listener %s not found",
			addr,
		)
	}

	rt.mu.Lock()
	defer rt.mu.Unlock()

	w := transport.Wrapper{Picker: picker, Inspector: ins}

	tr.Reload(&w)

	return nil
}

func (rt *Runtime) WatchBackend(
	balancer balancers.Balancer,
	backend *health.Backend,
) {
	defer rt.wg.Done()

	<-backend.Ctx.Done()

	balancer.RemoveBackend(
		backend.Address,
	)
}

func (rt *Runtime) Shutdown(timeout time.Duration) {
	log.Printf("Shutdown started")

	rt.mu.RLock()
	for _, tr := range rt.loops {
		tr.Close()
	}

	for _, bl := range rt.clusters {

		for _, backend := range bl.Backends() {
			log.Printf("backend=%s", backend.Address)

			backend.StartDrain(timeout, health.Removed)
		}
	}

	rt.mu.RUnlock()

	rt.wg.Wait()

	log.Printf("shutdown complete")
}

