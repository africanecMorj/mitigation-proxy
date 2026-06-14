package strategies

import (
	"math/rand"
	"math"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
)

type LeastConnections struct {
	balancers.BaseBalancer
}

func NewLeastConnections(
	backends []*health.Backend,
) *LeastConnections {

	return &LeastConnections{
		BaseBalancer: balancers.NewBaseBalancer(
			backends,
		),	
	}
}

func (lb *LeastConnections) Next() *health.Backend {

	backends := lb.Backends()

	if len(backends) == 0 {
		return nil
	}

	var (
		selected  *health.Backend
		bestScore = math.MaxFloat64
	)

	for _, backend := range backends {
		state := health.BackendState(backend.State.Load())

		penalty := 1.0

		switch state {
		case health.Unhealthy, health.Draining:
			continue

		case health.Recovering:
			penalty = 1.5

		case health.Suspect:
			penalty = 2
		}

		latency := 1.0

		if v := backend.EWMA(); v != 0 {
			latency = v
		}

		active := backend.ActiveConnections.Load()

		weight := backend.WeightValue()
	    score := float64(latency) * math.Sqrt(float64(active+1)) * penalty
		score += rand.Float64() * 0.01
		score = score / float64(weight)

		if score < bestScore {
			bestScore = score
			selected = backend
		}
	}

	if selected != nil {
		selected.ActiveConnections.Add(1)
	}

	return selected
}

