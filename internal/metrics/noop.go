package metrics

import (
	"time"

	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
)

type Noop struct{}

func NewNoop() *Noop {
	return &Noop{}
}

func (*Noop) ConnectionAccepted() {}
func (*Noop) ConnectionClosed() {}

func (*Noop) ConnectionActiveInc() {}
func (*Noop) ConnectionActiveDec() {}

func (*Noop) BytesReceived(uint64) {}
func (*Noop) BytesSent(uint64) {}

func (*Noop) BackendSelected(string, *health.Backend) {}

func (*Noop) BackendSuccess(string, *health.Backend) {}

func (*Noop) BackendFailure(string, *health.Backend) {}

func (*Noop) BackendConnectDuration(string, *health.Backend, time.Duration) {}

func (*Noop) BackendTTFB(string, *health.Backend, time.Duration) {}

func (*Noop) BackendState(string, *health.Backend) {}

func (*Noop) Reload() {}

func (*Noop) ReloadFailure() {}

func (*Noop) Drain(string, *health.Backend) {}