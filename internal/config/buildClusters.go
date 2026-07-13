package config

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers/strategies"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"

	"context"
)

const (
	DefaultBackendWeight = int64(1)
	DefaultBackendTau    = 10
)

func BuildClusters(cfg *Config) (map[string]balancers.Balancer, error) {
	result := make(map[string]balancers.Balancer)

	for _, c := range cfg.Clusters {
		var backends []*health.Backend

		for _, b := range c.Backends {
			ctx, cancel := context.WithCancel(context.Background())

			weight := b.Weight
			if weight <= 0 {
				weight = DefaultBackendWeight
			}

			tau := b.Tau
			if tau <= 0 {
				tau = DefaultBackendTau
			}

			be, err := health.NewBackend(
				b.Address, 
				tau, 
				weight, 
				ctx, 
				cancel,
			)
			if err != nil {
				return nil, err
			}

			backends = append(backends, be)
		}

		var bl balancers.Balancer

		switch c.LB {
		case "least_connections":
			bl = strategies.NewLeastConnections(backends)
		case "p2c":
			bl = strategies.NewP2C(backends)
		case "sticky":
			bl = strategies.NewSticky(backends)
		default:
			bl = strategies.NewRoundRobin(backends)
		}

		result[c.Name] = bl
	}

	return result, nil
}
