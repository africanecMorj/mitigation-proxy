package runtime

import (
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
	"github.com/africanecMorj/mitigation-proxy.git/internal/logger"
	"github.com/africanecMorj/mitigation-proxy.git/internal/metrics"
	"github.com/africanecMorj/mitigation-proxy.git/pkg"
)

func (rt *Runtime) Build(cfg *config.Config, metri metrics.Server) error {

	if err := rt.applyGlobals(cfg); err != nil {
		return err
	}

	clusters, err := config.BuildClusters(cfg)
	if err != nil {
		return err
	}

	rt.RegisterClusters(clusters)

	for _, listener := range cfg.Listeners {

		fd, err := pkg.BuildListener(listener.Address)
		if err != nil {
			return err
		}

		p, err := config.NewPicker(
			listener,
			clusters,
		)
		if err != nil {
			return err
		}

		w := transport.NewWrapper(
			listener.Routing.Type,
			&transport.Picker{p},
		)

		log := logger.NewPretty(rt.environment.Dev.Load())

		tr, err := transport.New(&w, log)
		if err != nil {
			return err
		}

		rt.Register(listener.Address, tr)
		go tr.Run(fd)
	}

	metri.Run()
	timeout := time.Duration(rt.environment.HealthCheckTimeout.Load())
    rt.StartHealthChecks(timeout)
	
	return nil
}

