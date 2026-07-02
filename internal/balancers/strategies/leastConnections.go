package strategies

import (
	"math"
	"math/rand"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
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
		case health.Unhealthy, health.Draining, health.Removed:
			continue

		case health.Recovering:
			penalty = 1.5

		case health.Suspect:
			penalty = 2
		}

		ttfb := math.Max(
			backend.TTFBValue(),
			1,
		)

		lat := math.Max(
			backend.EWMA(),
			1,
		)

		active := backend.ActiveConnections.Load()

		weight := max(float64(backend.WeightValue()), 1)
	
		score :=
			(float64(lat) + float64(ttfb)) *
			math.Sqrt(float64(active+1)) *
			penalty

		score /= weight
		
		if score < bestScore {
			bestScore = score
			selected = backend
		} else if math.Abs(score-bestScore) < 1e-9 {
			if rand.Intn(2) == 0 {
				selected = backend
			}
		}
	}

	return selected
}

