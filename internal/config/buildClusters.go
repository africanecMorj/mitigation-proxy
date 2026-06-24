package config

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers/strategies"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"

	"context"
	"time"
)

func BuildClusters(cfg *Config) (map[string]balancers.Balancer, error) {
	result := make(map[string]balancers.Balancer)

	for _, c := range cfg.Clusters {
		var backends []*health.Backend

		for _, b := range c.Backends {
			ctx, cancel := context.WithCancel(context.Background())

			be, err := health.NewBackend(b.Address, b.Tau, b.Weight, ctx, cancel)
			if err != nil {
				return nil, err
			}

			backends = append(backends, be)
		}

		var bl balancers.Balancer
		timeout, _ := time.ParseDuration("10s")

		switch c.LB {
		case "ewma":
			bl = strategies.NewLeastConnections(backends)
			bl.StartHealthChecks(timeout)
		case "p2c":
			bl = strategies.NewP2C(backends)
			bl.StartHealthChecks(timeout)
		default:
			bl = strategies.NewRoundRobin(backends)
			bl.StartHealthChecks(timeout)

		}

		result[c.Name] = bl
	}

	return result, nil
}
