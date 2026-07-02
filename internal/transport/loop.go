package transport 

import (
	"time"
	"io"
	"errors"
	"sync/atomic"

	"github.com/africanecMorj/mitigation-proxy.git/internal/logger"

	"golang.org/x/sys/unix"

)

type EventLoop struct {
	epfd int
	conns map[int]*Conn

	picker atomic.Value
	inspector atomic.Value

	logger logger.Logger
}

func NewEventLoop(
	w *Wrapper,
	l logger.Logger,
) (*EventLoop, error) {

	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	e := EventLoop{
		epfd:   epfd,
		conns: make(map[int]*Conn),
		logger: l,
	}

	e.picker.Store(w.Picker)
	e.inspector.Store(w.Inspector)

	l.Info(
		"event loop created",
		map[string]interface{}{
			"epfd": epfd,
		},
	)

	return &e, nil
}



func (l *EventLoop) Run(listenerFD int) error {
	if err := setNonblock(listenerFD); err != nil {
    	unix.Close(listenerFD)
    	return err
	}

	unix.EpollCtl(l.epfd, unix.EPOLL_CTL_ADD, listenerFD, &unix.EpollEvent{
		Events: unix.EPOLLIN | unix.EPOLLET,
		Fd:     int32(listenerFD),
	})

	defer unix.Close(l.epfd)

	events := make([]unix.EpollEvent, 1024)

	for {
		n, err := unix.EpollWait(l.epfd, events, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return err
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)

			if fd == listenerFD {
				l.acceptLoop(listenerFD)
				continue
			}

			c := l.conns[fd]
			if c == nil {
				continue
			}
			l.logger.Debug(
				"epoll event",
				map[string]interface{}{
					"conn_id": c.id,
					"fd": fd,
					"events": events[i].Events,
					"client_fd": c.clientFD,
					"backend_fd": c.backendFD,
					"state": c.state,
				},
			)

			if events[i].Events&(unix.EPOLLERR) != 0 {
				l.failAndClose(c)
				continue
			}

			l.handleIO(c, fd)
		}
	}
}

func (l *EventLoop) acceptLoop(listenerFD int) {
	for {
		nfd, sa, err := unix.Accept4(
			listenerFD,
			unix.SOCK_NONBLOCK|unix.SOCK_CLOEXEC,
		)
		if err != nil {
			if err == unix.EAGAIN || err == unix.EBADF {
				return
			}
			return
		}

		c := acquireConn()

		*c = Conn{
			id: logger.NextConnectionID(),
			clientFD:   nfd,
			clientAddr: sa,
			start:     time.Now(),
			inspector: l.inspector.Load().(InspectorFactory)(),
			state:     StateInspecting,
		}

		l.conns[nfd] = c

		l.logger.Info(
			"connection accepted",
			map[string]interface{}{
				"conn_id": c.id,
				"client_fd": nfd,
				"address": sa,
			},
		)

		if err := l.add(nfd, unix.EPOLLIN); err != nil {
    		unix.Close(nfd)
    		return 
		}
	}
}

func (l *EventLoop) handleIO(c *Conn, fd int) {
	for {
		prev := c.state

		switch c.state {
		case StateInspecting:
			l.handleInspecting(c)

		case StateRouting:
			l.handleRouting(c)

		case StateConnecting:
			l.handleConnecting(c, fd)
		
		case StateSending:
			l.handleSending(c)

		case StateProxy:
			l.handleProxy(c, fd)
		}

		if c.state == prev {
			return
		}
	}
}


func (l* EventLoop) handleInspecting(c *Conn) {
	done, err := c.inspector.Read(
		c.clientFD,
	)

	l.logger.Info(
		"inspection result",
		map[string]interface{}{
			"conn_id": c.id,
			"done": done,
		},
	)

	if err != nil {
		l.logger.Error(
			"inspection failed",
			map[string]interface{}{
				"conn_id": c.id,
				"error": err,
			},
		)
		if errors.Is(err, io.EOF) {
    		l.closeConn(c)
    		return
		}

		l.failAndClose(c)
		return 
	}

	if !done {
		return 
	}

	

	c.state = StateRouting
}

func (l *EventLoop) handleRouting(c *Conn) {
	ip := ipString(c.clientAddr)

	info := c.inspector.RouteKey()
	l.logger.Info(
		"route info",
		map[string]interface{}{
			"conn_id": c.id,
			"host": info.Host,
		},
	)

	backend, bfd, err := l.Picker().Pick(&info, ip)
	if err != nil {
		l.logger.Error(
			"routing failed",
			map[string]interface{}{
				"conn_id": c.id,
				"error": err,
			},
		)
		l.failAndClose(c)
		return
	}
	
	l.logger.Info(
		"backend selected",
		map[string]interface{}{
			"conn_id": c.id,
			"backend": backend.Address,
		},
	)
	
	c.start = time.Now()

	c.backend = backend
	c.backendFD = bfd

	c.c2b = NewSplicer(acquirePipe())
	c.b2c = NewSplicer(acquirePipe())

	l.conns[bfd] = c

	if err := l.add(bfd, unix.EPOLLOUT); err != nil {
		l.logger.Error(
			"routing failed",
			map[string]interface{}{
				"conn_id": c.id,
				"error": err,
			},
		)
		delete(l.conns, bfd)
		unix.Close(bfd)
    	return 
	}

	l.logger.Debug(
		"backend added to epoll",
		map[string]interface{}{
			"conn_id": c.id,
			"backend_fd": bfd,
		},
	)

	c.state = StateConnecting
}

func (l *EventLoop) handleConnecting(c *Conn, fd int) {
	l.logger.Info(
		"backend connecting",
		map[string]interface{}{
			"conn_id": c.id,
			"event_fd": fd,
			"backend_fd": c.backendFD,
			"client_fd": c.clientFD,
		},
	)

	if fd != c.backendFD {
		return
	}


	if !isConnected(fd) {
		if c.backend != nil {
			c.backend.MarkFailure()
		}
		l.failAndClose(c)
		return
	}

	
	latency := time.Since(c.start)
	c.backend.MarkSuccess(latency)

	if !c.activeCounted {
		c.backend.ActiveConnections.Add(1)
		c.activeCounted = true
	}
	  
	c.state = StateSending

	l.updateInterest(c)
}

func (l *EventLoop) handleProxy(c *Conn, fd int) {
	var res SpliceResult

	l.logger.Info(
		"proxy transfer",
		map[string]interface{}{
			"conn_id": c.id,
			"fd": fd,
			"client_fd": c.clientFD,
			"backend_fd": c.backendFD,
		},
	)

	if fd == c.clientFD {

		res = c.c2b.Transfer(c.clientFD, c.backendFD)
		l.logger.Debug(
			"splice result",
			map[string]interface{}{
				"conn_id": c.id,
				"fd": fd,
				"bytes": res.Bytes,
				"need_read": res.NeedRead,
				"need_write": res.NeedWrite,
				"err": res.Err,
			},
		)

		if res.Err == io.EOF {
			c.clientClosedRead = true
			c.backend.SetBytesSent(res.Bytes)
			unix.Shutdown(c.backendFD, unix.SHUT_WR)
		} else {

			c.backendWantsWrite = res.NeedWrite
		}

	} else {
		
		res = c.b2c.Transfer(c.backendFD, c.clientFD)
		if !c.firstBackendByte && res.Bytes > 0 {
			c.firstBackendByte = true
			l.logger.Debug(
				"splice result",
				map[string]interface{}{
					"conn_id": c.id,
					"fd": fd,
					"bytes": res.Bytes,
					"need_read": res.NeedRead,
					"need_write": res.NeedWrite,
					"err": res.Err,
				},
			)

			ttfb := time.Since(c.start)

			if c.backend != nil {
				c.backend.SetTTFB(ttfb)
			}

			l.logger.Debug(
				"backend first byte",
				map[string]interface{}{
					"conn_id": c.id,
					"backend": c.backend.Address,
					"ttfb": ttfb,
				},
			)
		}

		if res.Err == io.EOF {
			c.backend.SetBytesReceived(res.Bytes)
			c.backendClosedRead = true
			unix.Shutdown(c.clientFD, unix.SHUT_WR)
		} else {
			c.clientWantsWrite = res.NeedWrite
		}
	}

	if res.Err != nil && res.Err != io.EOF {
		l.logger.Error(
			"proxy error",
			map[string]interface{}{
				"conn_id": c.id,
				"fd": fd,
				"err": res.Err,
				"bytes": res.Bytes,
				"need_read": res.NeedRead,
				"need_write": res.NeedWrite,
			},
		)	
	
		if c.backend != nil {
			c.backend.MarkFailure()
		}

		l.failAndClose(c)
		return
	}

	l.maybeClose(c)
	l.updateInterest(c)
}

func (l *EventLoop) handleSending(c *Conn) error {
	data := c.inspector.Data()
	l.logger.Info(
		"State sending",
		map[string]interface{}{
			"conn_id": c.id,
			"sent": c.initialSent,
			"len":len(data),
			"record_len": 5 + (int(data[3])<<8 | int(data[4])),
		},
	)


	for c.initialSent < len(data) {

		n, err := unix.Write(
			c.backendFD,
			data[c.initialSent:],
		)

		l.logger.Debug(
			"startup write",
			map[string]interface{}{
				"conn_id": c.id,
				"written": n,
			},
		)

		if err != nil {

			l.logger.Error(
			"state sending failed",
				map[string]interface{}{
					"conn_id": c.id,
					"error": err,
				},
			)
			if err == unix.EINTR {
				continue
			}

			if err == unix.EAGAIN {

				c.backendWantsWrite = true
				l.updateInterest(c)

				return nil
			}

			return err
		}

		c.initialSent += n
	}

	c.backend.SetBytesSent(int64(c.initialSent))

	c.inspector.Close()
	c.inspector = nil

	c.backendWantsWrite = false

	c.state = StateProxy

	l.updateInterest(c)

	return nil
}

