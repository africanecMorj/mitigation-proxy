package transport

import (
	"net"
	"time"

	"golang.org/x/sys/unix"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)

const baseEvents = uint32(
	unix.EPOLLIN |
		unix.EPOLLET |
		unix.EPOLLRDHUP |
		unix.EPOLLHUP |
		unix.EPOLLERR,
)

func (l *EventLoop) Inspector() inspector.Inspector {
    return l.inspector.Load().(inspector.Inspector)
}

func (l *EventLoop) Picker() BackendPicker {
    return l.picker.Load().(BackendPicker)
}

func (l *EventLoop) updateInterest(c *Conn) {
	clientEvents := uint32(
		unix.EPOLLET |
		unix.EPOLLRDHUP |
		unix.EPOLLHUP |
		unix.EPOLLERR,	
	)

	backendEvents := uint32(
		unix.EPOLLET |
		unix.EPOLLRDHUP |
		unix.EPOLLHUP |
		unix.EPOLLERR,
	)
	if c.clientFD >= 0 {
		if !c.clientClosedRead {
    		clientEvents |= unix.EPOLLIN
		}

		if c.clientWantsWrite {
			clientEvents |= unix.EPOLLOUT
		}

		if err := l.mod(c.clientFD, clientEvents); err != nil {
			l.logger.Error(
				"epoll modify client failed",
				map[string]interface{}{
					"conn_id": c.id,
					"fd": c.clientFD,
					"error": err.Error(),
				},
			)

			l.failAndClose(c)
			return
		}
	}

	if c.backendFD >= 0 {
		if !c.backendClosedRead {
    		backendEvents |= unix.EPOLLIN
		}


		if c.backendWantsWrite {
			backendEvents |= unix.EPOLLOUT
		}

		if err := l.mod(c.backendFD, backendEvents); err != nil {
			l.logger.Error(
				"epoll modify backend failed",
				map[string]interface{}{
					"conn_id": c.id,
					"fd": c.backendFD,
					"error": err.Error(),
				},
			)

			l.failAndClose(c)
			return
		}
	}
}

func (l *EventLoop) maybeClose(c *Conn) {
	if !c.clientClosedRead ||
	   !c.backendClosedRead {
		return
	}

	if c.c2b != nil && c.c2b.buf > 0 {
		return
	}

	if c.b2c != nil && c.b2c.buf > 0 {
		return
	}
	
	l.closeConn(c)
	
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

	
	if c.inspector != nil {
		c.inspector.Close()
	}

	releaseConn(c)
}

func (l *EventLoop) failAndClose(c *Conn) {
	if c.closed {
		return
	}
	c.closed = true

	l.logger.Error(
		"connection failed",
		map[string]interface{}{
			"conn_id": c.id,
			"client_fd": c.clientFD,
			"backend_fd": c.backendFD,
		},
	)


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
		}
		c.backend.MarkFailure()
	}

	if c.c2b != nil {
		releasePipe(c.c2b.pipe)
	}

	if c.b2c != nil {
		releasePipe(c.b2c.pipe)
	}

	if c.inspector != nil {
		c.inspector.Close()
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

	l.logger.Debug(
		"epoll add",
		map[string]interface{}{
			"fd": fd,
			"events": events,
			"error": err,
		},
	)

	return err
}

func (l *EventLoop) del(fd int) {
	_ = unix.EpollCtl(l.epfd, unix.EPOLL_CTL_DEL, fd, nil)
}

func isConnected(fd int) bool {
	soErr, _ := unix.GetsockoptInt(
		fd,
		unix.SOL_SOCKET,
		unix.SO_ERROR,
	)

	return soErr == 0
}

func ipString(sa unix.Sockaddr) string {
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
	c.clientFD = -1
    c.backendFD = -1

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

	c.inspector = nil
	
	c.closed = false
}


