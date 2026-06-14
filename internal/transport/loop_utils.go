package transport

import (
	"log"
	"net"
	"time"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)

type EventLoop struct {
	epfd   int
	conns  map[int]*Conn
	picker BackendPicker
	inspector inspector.Inspector
}

func NewEventLoop(w *Wrapper) (*EventLoop, error) {
	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	return &EventLoop{
		epfd:   epfd,
		conns:  make(map[int]*Conn),
		picker: w.picker,
		inspector: w.inspector,
	}, nil
}

const baseEvents = uint32(
	unix.EPOLLIN |
		unix.EPOLLET |
		unix.EPOLLRDHUP |
		unix.EPOLLHUP |
		unix.EPOLLERR,
)

func (l *EventLoop) updateInterest(c *Conn) {
	clientEvents := baseEvents
	backendEvents := baseEvents

	if c.clientFD > 0 {

		if c.clientWantsWrite {
			clientEvents |= unix.EPOLLOUT
		}

		if err := l.mod(c.clientFD, clientEvents); err != nil {
			log.Printf("epoll mod client fd=%d err=%v", c.clientFD, err)
			l.failAndClose(c)
			return
		}
	}

	if c.backendFD > 0 {

		if c.backendWantsWrite {
			backendEvents |= unix.EPOLLOUT
		}

		if err := l.mod(c.backendFD, backendEvents); err != nil {
			log.Printf("epoll mod client fd=%d err=%v", c.clientFD, err)
			l.failAndClose(c)
			return
		}
	}
}

func (l *EventLoop) registerConn(c *Conn) {
	l.conns[c.clientFD] = c
	l.conns[c.backendFD] = c

	l.add(c.clientFD, unix.EPOLLIN)
	l.add(c.backendFD, unix.EPOLLIN)

	c.backend.ActiveConnections.Add(1)
}

func (l *EventLoop) maybeClose(c *Conn) {
	if c.clientClosedRead && c.backendClosedRead {
		l.closeConn(c)
	}
}

func (l *EventLoop) closeConn(c *Conn) {
	if c.closed {
		return
	}
	c.closed = true

	if c.clientFD >= 0 {
		l.del(c.clientFD)
		unix.Close(c.clientFD)
		delete(l.conns, c.clientFD)
	}

	if c.backendFD >= 0 {
		l.del(c.backendFD)
		unix.Close(c.backendFD)
		delete(l.conns, c.backendFD)
	}

	if c.backend != nil && c.activeCounted {
		c.backend.ActiveConnections.Add(-1)
	}

	latency := time.Since(c.start)

	if c.backend != nil {
		c.backend.MarkSuccess(latency)
	}

	if c.c2b != nil {
		releasePipe(c.c2b.pipe)
	}

	if c.b2c != nil {
		releasePipe(c.b2c.pipe)
	}

	releaseConn(c)
}

func (l *EventLoop) failAndClose(c *Conn) {
	if c.closed {
		return
	}
	c.closed = true

	if c.clientFD >= 0 {
		l.del(c.clientFD)
		unix.Close(c.clientFD)
		delete(l.conns, c.clientFD)
	}

	if c.backendFD >= 0 {
		l.del(c.backendFD)
		unix.Close(c.backendFD)
		delete(l.conns, c.backendFD)
	}

	if c.backend != nil {
		if c.activeCounted {
			c.backend.ActiveConnections.Add(-1)

			c.backend.MarkFailure()
		}
	}

	if c.c2b != nil {
		releasePipe(c.c2b.pipe)
	}

	if c.b2c != nil {
		releasePipe(c.b2c.pipe)
	}

	releaseConn(c)
}

func setNonblock(fd int) error {
	return unix.SetNonblock(fd, true)
}

func (l *EventLoop) mod(fd int, events uint32) error {
	return unix.EpollCtl(
		l.epfd,
		unix.EPOLL_CTL_MOD,
		fd,
		&unix.EpollEvent{
			Events: events,
			Fd:     int32(fd),
		},
	)
}

func (l *EventLoop) add(fd int, events uint32) error {
	err := unix.EpollCtl(
		l.epfd,
		unix.EPOLL_CTL_ADD,
		fd,
		&unix.EpollEvent{
			Events: events |
				unix.EPOLLET |
				unix.EPOLLRDHUP |
				unix.EPOLLHUP |
				unix.EPOLLERR,
			Fd: int32(fd),
		},
	)

	log.Printf(
		"EPOLL ADD fd=%d events=%#x err=%v",
		fd,
		events,
		err,
	)

	return err
}

func (l *EventLoop) del(fd int) {
	_ = unix.EpollCtl(l.epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func isConnected(fd int) bool {
	errno, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
	if err != nil {
		return false
	}
	return errno == 0
}

func iPString(sa unix.Sockaddr) string {
	switch addr := sa.(type) {

	case *unix.SockaddrInet4:
		return net.IP(addr.Addr[:]).String()

	case *unix.SockaddrInet6:
		return net.IP(addr.Addr[:]).String()

	default:
		return ""
	}
}

func resetConn(c *Conn) {
	c.clientFD = 0
	c.backendFD = 0

	c.start = time.Time{}
	c.firstBackendByte = false

	c.clientClosedRead = false
	c.backendClosedRead = false

	c.clientWantsWrite = false
	c.backendWantsWrite = false

	c.activeCounted = false

	c.clientAddr = nil

	c.backend = nil

	c.c2b = nil
	c.b2c = nil

	c.closed = false
}
