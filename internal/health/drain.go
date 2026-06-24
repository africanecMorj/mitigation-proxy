package health

import (
	"log"
	"time"
)

func (b *Backend) StartDrain(timeout time.Duration, state BackendState) {
	if !b.draining.CompareAndSwap(false, true) {
		log.Printf("monitor drain failed")
		return
	}

	b.SetState(Draining)

	b.DrainStartedAt.Store(
		time.Now().UnixNano(),
	)

	log.Printf("monitor drain started")
	go b.monitorDrain(timeout, state)
}

func (b *Backend) monitorDrain(timeout time.Duration, state BackendState) {
	log.Printf("monitor drain started %s", b.Address)
	defer b.draining.Store(false)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {
		log.Printf(
			"backend=%s active=%d",
			b.Address,
			b.ActiveConnections.Load(),
		)

		if b.ActiveConnections.Load() == 0 {
			log.Printf(
				"backend %s drain complete",
				b.Address,
			)

		
			
			b.SetState(state)
			if state == Removed {
				b.Close()
			}
			return
		}

		started := time.Unix(
			0,
			b.DrainStartedAt.Load(),
		)

		if time.Since(started) >= timeout {

			log.Printf(
				"backend %s drain timeout exceeded",
				b.Address,
			)


			if state == Healthy {
				b.SetState(Unhealthy)
				return
			}	
			
			if state == Removed {
				b.Close()
			}
			b.SetState(state)
			return
		}
	}
}

