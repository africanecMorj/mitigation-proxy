package transport 

import (
	"time"
	"io"
	"log"
	"errors"
	"sync/atomic"

	"golang.org/x/sys/unix"

)

type EventLoop struct {
	epfd   int
	conns  map[int]*Conn
	picker atomic.Value
    inspector atomic.Value

}

func NewEventLoop(w *Wrapper) (*EventLoop, error) {
	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}

	e := EventLoop{
		epfd:   epfd,
		conns:  make(map[int]*Conn),
	}

	e.picker.Store(w.Picker)
	e.inspector.Store(w.Inspector)

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
			log.Printf(
				"EPOLL fd=%d events=%#x client=%d backend=%d state=%v",
				fd,
				events[i].Events,
				c.clientFD,
				c.backendFD,
				c.state,
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
			clientFD:   nfd,
			clientAddr: sa,
			start:     time.Now(),
			state:     StateInspecting,
		}

		l.conns[nfd] = c

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
	done, err := l.Inspector().Read(
		c.clientFD,
	)

	log.Printf("sniff result: done=%v, err=%v", done, err)

	if err != nil {
	
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

	info := l.Inspector().RouteKey()
	log.Printf("Routeinfo:%+v", info)

	backend, bfd, err := l.Picker().Pick(info.SNI, info.Host, info.ALPN, ip)
	if err != nil {
		l.failAndClose(c)
		return
	}
	log.Printf("backend selected address: %+v", backend.Address)
	
	c.start = time.Now()

	c.backend = backend
	c.backendFD = bfd

	c.c2b = NewSplicer(acquirePipe())
	c.b2c = NewSplicer(acquirePipe())

	l.conns[bfd] = c

	if err := l.add(bfd, unix.EPOLLOUT); err != nil {
    	unix.Close(bfd)
    	return 
	}
	log.Printf("added backend fd=%d to epoll", bfd)

	log.Printf("routing sni=%s", info.SNI)
	c.state = StateConnecting
}

func (l *EventLoop) handleConnecting(c *Conn, fd int) {
	log.Printf("ENTER CONNECTING fd=%d", fd)
	log.Printf(
    "CONNECTING event_fd=%d backend_fd=%d client_fd=%d",
    fd,
    c.backendFD,
    c.clientFD,
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

	log.Printf(
		"PROXY fd=%d clientFD=%d backendFD=%d",
		fd,
		c.clientFD,
		c.backendFD,
	)

	if fd == c.clientFD {
		log.Printf("CLIENT SIDE")
	}

	if fd == c.backendFD {
		log.Printf("BACKEND SIDE")
	}

	if fd == c.clientFD {
		log.Println("Succesfully proxied (client)")
		res = c.c2b.Transfer(c.clientFD, c.backendFD)

		if res.Err == io.EOF {
			c.clientClosedRead = true
			unix.Shutdown(c.backendFD, unix.SHUT_WR)
		} else {
			c.backendWantsWrite = res.NeedWrite
		}

	} else {
		res = c.b2c.Transfer(c.backendFD, c.clientFD)
		if !c.firstBackendByte && res.Bytes > 0 {
			c.firstBackendByte = true

			ttfb := time.Since(c.start)

			if c.backend != nil {
				c.backend.SetTTFB(int64(ttfb))
			}

			log.Printf(
				"backend=%s ttfb=%s",
				c.backend.Address,
				ttfb,
			)
		}

		log.Println("Succesfully proxied (backend)")
		if res.Err == io.EOF {
			c.backendClosedRead = true
			unix.Shutdown(c.clientFD, unix.SHUT_WR)
		} else {
			c.clientWantsWrite = res.NeedWrite
		}
	}

	if res.Err != nil && res.Err != io.EOF {
		
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

	data := l.Inspector().Data()


	for c.initialSent < len(data) {

		n, err := unix.Write(
			c.backendFD,
			data[c.initialSent:],
		)

		if err != nil {

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

	l.Inspector().Close()

	c.backendWantsWrite = false

	c.state = StateProxy

	l.updateInterest(c)

	return nil
}