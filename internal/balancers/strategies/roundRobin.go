package strategies

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
)

type RoundRobin struct {
	balancers.BaseBalancer
}

func NewRoundRobin(
	backends []*health.Backend,
) *RoundRobin {

	return &RoundRobin{
		BaseBalancer: balancers.NewBaseBalancer(
			backends,
		),	
	}
}

func (rr *RoundRobin) Next() *health.Backend {
	backends := rr.Backends()

	var best *health.Backend

	total := int64(0)

	for _, b := range backends {

		var effectiveWeight int64
		weight := b.WeightValue()

		switch health.BackendState(
			b.State.Load(),
		) {

		case health.Unhealthy:
			continue

		case health.Draining:
			continue

		case health.Recovering:
			effectiveWeight = max(
				1,
				weight/4,
			)

		case health.Suspect:
			effectiveWeight = max(
				1,
				weight/2,
			)

		case health.Healthy:
			effectiveWeight = weight
		}

		total += effectiveWeight

		currentWeight := b.AddCurrentWeight(effectiveWeight) 

		if best == nil ||
			currentWeight >
				best.CurrentWeightValue() {

			best = b
		}
	}

	if best == nil {
		return nil
	}

	best.AddCurrentWeight(-total) 

	return best
}
