package runtime

import (
	"fmt"
	"time"
	"sync/atomic"
	"sync"
	"context"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/logger"
	"github.com/africanecMorj/mitigation-proxy.git/internal/metrics"
)

type Runtime struct {
    clusters atomic.Pointer[clusterSnapshot]
    loops    atomic.Pointer[loopSnapshot]

	environment *EnvironmentalVars

    wg sync.WaitGroup
}

type EnvironmentalVars struct {
	MaxConnections atomic.Int32
	Burst 		   atomic.Int32
	RPS 		   atomic.Int32

	DrainTimeout   atomic.Int64
	ShutdownTimeout   atomic.Int64
	HealthCheckTimeout   atomic.Int64

	Dev atomic.Bool
}

type loopSnapshot struct {
	loops map[string]*transport.Transport
}

type clusterSnapshot struct {
    clusters map[string]balancers.Balancer
}

func New() *Runtime {
	rt := &Runtime{
		environment: &EnvironmentalVars{},
	}


	rt.clusters.Store(&clusterSnapshot{
		clusters: make(map[string]balancers.Balancer),
	})

	rt.loops.Store(&loopSnapshot{
		loops: make(map[string]*transport.Transport),
	})

	return rt
}

func (rt *Runtime) clustersSnapshot() *clusterSnapshot {
    return rt.clusters.Load()
}

func (rt *Runtime) ShutdownTimeout() time.Duration {
	return time.Duration(rt.environment.ShutdownTimeout.Load())
}

func (rt *Runtime) loopsSnapshot() *loopSnapshot {
    return rt.loops.Load()
}

func (rt *Runtime) Register(
	addr string,
	tr *transport.Transport,
) {
	old := rt.loops.Load()

	loops := cloneMap(old.loops)
	loops[addr] = tr

	rt.loops.Store(&loopSnapshot{
		loops: loops,
	})
}

func cloneMap[K comparable, V any](src map[K]V) map[K]V {
	dst := make(map[K]V, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func (rt *Runtime) RegisterClusters(clusters map[string]balancers.Balancer) {
    rt.clusters.Store(&clusterSnapshot{
        clusters: clusters,
    })
}

func (rt *Runtime) Clusters(fn func(string, *health.Backend)) {
	snap := rt.clusters.Load()

	for name, lb := range snap.clusters {
		for _, b := range lb.Backends() {
			fn(name, b)
		}
	}
}

func (rt *Runtime) Get(addr string) (*transport.Transport, bool) {
    snap := rt.loops.Load()

    tr, ok := snap.loops[addr]

    return tr, ok
}

func (rt *Runtime) reload(
	addr string,
	picker transport.BackendPicker,
	ins transport.InspectorFactory,
) error {

	tr, ok := rt.Get(addr)
	if !ok {
		return fmt.Errorf(
			"listener %s not found",
			addr,
		)
	}

	tr.Reload(&transport.Wrapper{
		Picker:    picker,
		Inspector: ins,
	})

	return nil
}

func (rt *Runtime) Shutdown(timeout time.Duration, metrics metrics.Server) error {
	 ctx, cancel := context.WithTimeout(
        context.Background(),
        rt.ShutdownTimeout(),
    )
    defer cancel()

	if metrics != nil {
		err := metrics.Close(ctx)
		if err != nil {
			return err
		}
	}
	

    var log = logger.NewPretty(false)

    loops := rt.loops.Load()

    for _, tr := range loops.loops {
        tr.Close()
    }

    clusters := rt.clusters.Load()

   	for _, lb := range clusters.clusters {
		for _, b := range lb.Backends() {
			b.StartDrain(timeout, health.Shutdown)
		}
	}

	var wg sync.WaitGroup

	for _, lb := range clusters.clusters {
		for _, b := range lb.Backends() {
			wg.Add(1)

			go func(b *health.Backend) {
				defer wg.Done()
				<-b.Ctx.Done()
			}(b)
		}
	}

	wg.Wait()

    log.Info("Shutdown complete", nil)

	return nil
}

func (rt *Runtime) StartHealthChecks(interval time.Duration) {
	snapshot := rt.clusters.Load()

	for _, balancer := range snapshot.clusters {
		rt.wg.Add(1)

		go func(bl balancers.Balancer) {
			defer rt.wg.Done()

			ticker := time.NewTicker(interval)
			defer ticker.Stop()

			for {
				removed := 0
				backends := bl.Backends()

				for _, b := range backends {

					if b.StateValue() == health.Removed {
						removed++
						continue
					}

					b.HealthCheck()
				}

				if removed == len(backends) {
					return
				}

				<-ticker.C
			}
		}(balancer)
	}
}

