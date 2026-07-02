package runtime

import (
	"time"
	
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

func (rt *Runtime) Reload(cfg *config.Config) error {

	if err := rt.applyGlobals(cfg); err != nil {
		return err
	}

	old := rt.clusters.Load()

	newClusters, err := config.BuildClusters(cfg)
	if err != nil {
		return err
	}

	for _, listener := range cfg.Listeners {

		p, err := config.NewPicker(
			listener,
			newClusters,
		)
		if err != nil {
			return err
		}

		w := transport.NewWrapper(
			listener.Routing.Type,
			&transport.Picker{p},
		)

		if err := rt.reload(
			listener.Address,
			w.Picker,
			w.Inspector,
		); err != nil {
			return err
		}
	}

    
	rt.RegisterClusters(newClusters)
	timeout := time.Duration(rt.environment.HealthCheckTimeout.Load())
    rt.StartHealthChecks(timeout)

	timeout = time.Duration(rt.environment.DrainTimeout.Load())
	for _, lb := range old.clusters {
		for _, backend := range lb.Backends() {
			backend.StartDrain(
				timeout,
				health.Removed,
			)
		}
	}

	return nil
}