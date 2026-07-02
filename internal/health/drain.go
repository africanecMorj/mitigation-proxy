package health

import (
	"time"

    "github.com/africanecMorj/mitigation-proxy.git/internal/logger"
)

func (b *Backend) StartDrain(timeout time.Duration, state BackendState) {
	if !b.draining.CompareAndSwap(false, true) {
		var log = logger.NewPretty(false)
		
		log.Error("drain failed",
			map[string]interface{}{
				"backend":b.Address,
			},
		)
		return
	}

	b.SetState(Draining)
	b.totalDrains.Add(1)

	b.DrainStartedAt.Store(
		time.Now().UnixNano(),
	)

	go b.monitorDrain(timeout, state)
}

func (b *Backend) monitorDrain(timeout time.Duration, state BackendState) {
	log := logger.NewPretty(false)

	log.Info("monitor drain started", map[string]interface{}{
		"backend": b.Address,
	})

	defer b.draining.Store(false)

	ticker := time.NewTicker(100 * time.Millisecond) 
	defer ticker.Stop()

	startedNs := b.DrainStartedAt.Load()
	if startedNs == 0 {
		log.Error("drain started timestamp missing", map[string]interface{}{
			"backend": b.Address,
		})
		return
	}

	started := time.Unix(0, startedNs)

	for {
		elapsed := time.Since(started)

		if b.ActiveConnections.Load() == 0 {
			log.Info("drain complete", map[string]interface{}{
				"backend": b.Address,
			})

			b.SetState(state)

			if state == Shutdown {
				b.Close()
			}

			return
		}

		if elapsed >= timeout {
			log.Error("drain timeout exceeded", map[string]interface{}{
				"backend": b.Address,
			})

			if state == Healthy {
				b.SetState(Unhealthy)
			} else {
				b.SetState(state)

				if state == Shutdown {
					b.Close()
				}
			}

			return
		}

		log.Info("drain progress", map[string]interface{}{
			"backend": b.Address,
			"active":  b.ActiveConnections.Load(),
			"elapsed": elapsed,
			"timeout": timeout,
		})

		<-ticker.C
	}
}