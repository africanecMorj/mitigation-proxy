package runtime

import (
	"fmt"
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	
)

func (rt *Runtime) Drain(
	cluster string,
	address string,
) error {

	snap := rt.clusters.Load()

	bl, ok := snap.clusters[cluster]
	if !ok {
		return fmt.Errorf(
			"cluster %s not found",
			cluster,
		)
	}

	timeout := time.Duration(rt.environment.DrainTimeout.Load())

	for _, b := range bl.Backends() {
		if address == "*" {
			b.StartDrain(timeout, health.Healthy)
		}

		if b.Address == address {
			b.StartDrain(timeout, health.Healthy)
			return nil
		}
	}

	if address != "*" {
		return fmt.Errorf(
			"backend %s not found",
			address,
		)
	}

	return nil

}

func (rt *Runtime) Undrain(
	cluster string,
	address string,
) error {

	clusters := rt.clusters.Load()

	bl, ok := clusters.clusters[cluster]
	if !ok {
		return fmt.Errorf(
			"cluster %s not found",
			cluster,
		)
	}

	for _, b := range bl.Backends() {

		if b.Address != address && b.Address != "*" {
			continue
		}

		if b.StateValue() == health.Draining {
			b.SetState(health.Healthy)
		}

		b.DrainStartedAt.Store(0)
		b.SetDraining(false)

		return nil
	}

	return fmt.Errorf(
		"backend %s not found",
		address,
	)
}
