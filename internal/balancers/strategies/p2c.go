package strategies

import (
	"math"
	"math/rand/v2"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

type P2C struct {
	balancers.BaseBalancer
}

func NewP2C(
	backends []*health.Backend,
) *P2C {

	return &P2C{
		BaseBalancer: balancers.NewBaseBalancer(
			backends,
		),
	}
}

func (lb *P2C) Next() *health.Backend {
	backends := lb.Backends()

	if len(backends) == 0 {
		return nil
	}

	if len(backends) == 1 {
		return backends[0]
	}

	b1 := rand.IntN(len(backends))

	b2 := rand.IntN(len(backends))
	for b2 == b1 {
		b2 = rand.IntN(len(backends))
	}

	var (
		selected  *health.Backend
		bestScore = math.MaxFloat64
	)

	for _, backend := range []*health.Backend{
		backends[b1],
		backends[b2],
	} {
		switch backend.StateValue() {
		case health.Unhealthy,
			health.Draining,
			health.Removed:
			continue
		}

		penalty := 1.0

		switch backend.StateValue() {
		case health.Recovering:
			penalty = 1.5

		case health.Suspect:
			penalty = 2
		}

		ttfb := int64(1)

		if v := backend.TTFBValue(); v != 0 {
			ttfb = v
		}

		latency := math.Max(
			backend.EWMA(),
			1,
		)

		active := backend.ActiveConnections.Load()

		weight := backend.WeightValue()

		score := (float64(latency) + 0.2*float64(ttfb)) * math.Sqrt(float64(active+1)) * penalty / float64(weight)

		if score < bestScore {
			bestScore = score
			selected = backend
		}
	}

	return selected
}
