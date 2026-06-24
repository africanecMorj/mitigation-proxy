package runtime

import (

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"time"
)


type BackendStats struct {
	Address string `json:"address"`

	State string `json:"state"`

	ActiveConnections int64 `json:"active"`

	Requests  uint64 `json:"requests"`
	Successes uint64 `json:"successes"`
	Failures  uint64 `json:"failures"`

	EWMALatency string `json:"ewma_latency"`
	AvgLatency  string `json:"avg_latency"`
	EWMATTFB    string `json:"ewma_ttfb"`
}

func (rt *Runtime) Stats() map[string][]BackendStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	stats := make(map[string][]BackendStats)

	for clusterName, balancer := range rt.clusters {
		for _, b := range balancer.Backends() {
			stats[clusterName] = append(
				stats[clusterName],
				buildBackendStats(b),
			)
		}
	}

	return stats
}

func (rt *Runtime) StatsCluster(cluster string) map[string][]BackendStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	stats := make(map[string][]BackendStats)

	balancer, ok := rt.clusters[cluster]
	if !ok {
		return stats
	}

		for _, b := range balancer.Backends() {
		stats[cluster] = append(
			stats[cluster],
			buildBackendStats(b),
		)
	}


	return stats
}

func (rt *Runtime) StatsBackend(cluster, address string) map[string][]BackendStats {
	rt.mu.RLock()
	defer rt.mu.RUnlock()

	stats := make(map[string][]BackendStats)

	balancer, ok := rt.clusters[cluster]
	if !ok {
		return stats
	}


	for _, b := range balancer.Backends() {
		if b.Address == address {
			stats[cluster] = append(
				stats[cluster],
				buildBackendStats(b),
			)
			break
		}
	}


	return stats
}


func buildBackendStats(b *health.Backend) BackendStats {
	return BackendStats{
		Address: b.Address,

		State: b.StateValue().String(),

		ActiveConnections: b.ActiveConnections.Load(),

		Requests: b.Requests.Load(),

		Successes: b.Successes(),

		Failures: b.Failures(),

		EWMALatency: time.Duration(
			b.EWMA(),
		).String(),

		AvgLatency: time.Duration(
			b.AvgLatency(),
		).String(),

		EWMATTFB: time.Duration(
			b.TTFBValue(),
		).String(),
	}
}




