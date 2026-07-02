package metrics

import (
	"time"
	"context"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

type Metrics interface {

	// Connections
	ConnectionAccepted()
	ConnectionClosed()

	ConnectionActiveInc()
	ConnectionActiveDec()

	BytesReceived(n uint64)
	BytesSent(n uint64)

	// Backend
	BackendSelected(cluster, backend string)

	BackendSuccess(cluster, backend string)

	BackendFailure(cluster, backend string)

	BackendConnectDuration(cluster, backend string, d time.Duration)

	BackendTTFB(cluster, backend string, d time.Duration)

	BackendState(cluster, backend string, state health.BackendState)

	// Runtime
	Reload()
	ReloadFailure()

	Drain(cluster, backend string)
}

type Server interface {
    Run() error
    Close(context.Context) error
}