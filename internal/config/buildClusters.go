package config

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers/strategies"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/pkg"
)

func BuildClusters(cfg *Config) (map[string]balancers.Balancer, error) {
	result := make(map[string]balancers.Balancer)

	for _, c := range cfg.Clusters {
		var backends []*health.Backend

		for _, b := range c.Backends {
			family, sa, err := pkg.ResolveSockaddr(b.Address)
			
			if err != nil {
				return nil, err
			}
			
			be := health.Backend{
				Address: b.Address,
				Tau: b.Tau,
				Family:family,
				SockAddr:sa,
			}
			be.SetWeight(b.Weight)

			go be.StartHealthChecks()

			backends = append(backends, &be)
		}

		var bl balancers.Balancer

		switch c.LB {
		case "ewma":
			bl = strategies.NewLeastConnections(backends)
		// case "p2c":
		// 	bl = NewP2C(backends)
		default:
			bl = strategies.NewRoundRobin(backends)
		}

		result[c.Name] = bl
	}

	return result, nil
}




