package transport

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
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
	clientFD  int
	backendFD int

	state connState

	activeCounted bool

	initialSent int

	backend *health.Backend

	c2b *Splicer
	b2c *Splicer

	clientClosedRead  bool
	backendClosedRead bool

	clientWantsWrite  bool
	backendWantsWrite bool

	clientAddr unix.Sockaddr

	closed bool

	firstBackendByte bool

	start time.Time
}
