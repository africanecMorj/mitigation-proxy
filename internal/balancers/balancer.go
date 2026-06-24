package balancers

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	
	"time"
)


type Balancer interface {
	Next() *health.Backend
	AddBackend(*health.Backend)
	RemoveBackend(string)
	Backends() []*health.Backend
	StartHealthChecks(interval time.Duration)
}
