package transport

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"time"

	"golang.org/x/sys/unix"
)

type connState int

const (
	StateInspecting connState = iota
	StateConnecting
	StateRouting
	StateSending
	StateProxy
)

type Conn struct {
	id uint64

	clientFD  int
	backendFD int

	state connState

	activeCounted bool

	inspector inspector.Inspector

	initialSent int

	backend *health.Backend

	c2b *Splicer
	b2c *Splicer

	clientClosedRead  bool
	backendClosedRead bool

	clientWantsWrite  bool
	backendWantsWrite bool

	clientAddr unix.Sockaddr

	readyToClose bool
	closed bool

	firstBackendByte bool

	start time.Time
}
