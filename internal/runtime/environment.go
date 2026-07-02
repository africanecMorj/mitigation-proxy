package runtime

import (
	"time"
	"fmt"

	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
)

const (
	DefaultMaxConnections    = 1000000
	DefaultBurst             = 100
	DefaultRPS               = 1000

	DefaultDrainTimeout      = 5 * time.Minute
	DefaultShutdownTimeout   = 5 * time.Minute
	DefaultHealthcheckTimeout = 10 * time.Second
)

func (rt *Runtime) applyGlobals(cfg *config.Config) error {
	maxConnections := cfg.Global.Limits.MaxConnections
	if maxConnections <= 0 {
		maxConnections = DefaultMaxConnections
	}

	burst := cfg.Global.Limits.Burst
	if burst <= 0 {
		burst = DefaultBurst
	}

	rps := cfg.Global.Limits.RPS
	if rps <= 0 {
		rps = DefaultRPS
	}

	rt.environment.MaxConnections.Store(maxConnections)
	rt.environment.Burst.Store(burst)
	rt.environment.RPS.Store(rps)
	rt.environment.Dev.Store(cfg.Global.Dev)

	drainTimeout := DefaultDrainTimeout
	if cfg.Global.Timeouts.DrainTimeout != "" {
		d, err := time.ParseDuration(cfg.Global.Timeouts.DrainTimeout)
		if err != nil {
			return fmt.Errorf("invalid drain timeout: %w", err)
		}
		drainTimeout = d
	}

	shutdownTimeout := DefaultShutdownTimeout
	if cfg.Global.Timeouts.ShutdownTimeout != "" {
		d, err := time.ParseDuration(cfg.Global.Timeouts.ShutdownTimeout)
		if err != nil {
			return fmt.Errorf("invalid shutdown timeout: %w", err)
		}
		shutdownTimeout = d
	}

	healthcheckTimeout := DefaultHealthcheckTimeout
	if cfg.Global.Timeouts.HealthCheckTimeout != "" {
		d, err := time.ParseDuration(cfg.Global.Timeouts.HealthCheckTimeout)
		if err != nil {
			return fmt.Errorf("invalid healthcheck timeout: %w", err)
		}
		healthcheckTimeout = d
	}

	rt.environment.DrainTimeout.Store(int64(drainTimeout))
	rt.environment.ShutdownTimeout.Store(int64(shutdownTimeout))
	rt.environment.HealthCheckTimeout.Store(int64(healthcheckTimeout))

	return nil
}