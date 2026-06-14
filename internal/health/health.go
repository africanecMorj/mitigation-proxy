package health

import (
	"log"
	"time"
	"math"
	"errors"

	"golang.org/x/sys/unix"
)

func (b *Backend) ObserveLatency(sample time.Duration) {
	b.ewmaMu.Lock()
	defer b.ewmaMu.Unlock()

	now := time.Now()

	if b.lastEWMAUpdate.IsZero() {
		b.ewma = float64(sample)
		b.lastEWMAUpdate = now
		return
	}

	dt := now.Sub(b.lastEWMAUpdate).Seconds()

	alpha := math.Exp(-dt / b.Tau)

	b.ewma =
		b.ewma*alpha +
			float64(sample)*(1-alpha)

	b.lastEWMAUpdate = now
}

func (b *Backend) EWMA() float64 {
	b.ewmaMu.Lock()
	defer b.ewmaMu.Unlock()

	return b.ewma
}


func (b *Backend) SetState(state BackendState) {
	b.State.Store(int32(state))
	b.LastStateChange.Store(time.Now().UnixNano())
}

func (b *Backend) TTFBValue() int64 {
    return int64(b.ttfb.Load())
}

func (b *Backend) SetTTFB(t int64) {
    b.ttfb.Store(t)
}

func (b *Backend) WeightValue() int64 {
    return int64(b.Weight.Load())
}

func (b *Backend) SetWeight(weight int64) {
    b.Weight.Store(weight)
}

func (b *Backend) CurrentWeightValue() int64 {
    return b.currentWeight.Load()
}

func (b *Backend) AddCurrentWeight(delta int64) int64 {
    return b.currentWeight.Add(delta)
}

func (b *Backend) StartDrain(timeout time.Duration) {
	if !b.draining.CompareAndSwap(false, true) {
		return
	}

	b.SetState(Draining)

	b.DrainStartedAt.Store(
		time.Now().UnixNano(),
	)

	go b.monitorDrain(timeout)
}

func (b *Backend) monitorDrain(timeout time.Duration) {
	defer b.draining.Store(false)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for range ticker.C {

		if b.ActiveConnections.Load() == 0 {
			log.Printf(
				"backend %s drain complete",
				b.Address,
			)

			b.SetState(Unhealthy)
			return
		}

		started := time.Unix(
			0,
			b.DrainStartedAt.Load(),
		)

		if time.Since(started) >= timeout {

			log.Printf(
				"backend %s drain timeout exceeded",
				b.Address,
			)

			b.SetState(Unhealthy)
			return
		}
	}
}

func (b *Backend) HealthCheck() {
	fd, err := unix.Socket(
		b.Family,
		unix.SOCK_STREAM,
		0,
	)
	if err != nil {
		return
	}

	defer unix.Close(fd)

	if err := unix.SetNonblock(fd, true); err != nil {
		return
	}

	start := time.Now()

	err = unix.Connect(fd, b.SockAddr)

	if err != nil &&
		err != unix.EINPROGRESS {

		b.MarkFailure()
		return
	}

	pollfds := []unix.PollFd{
		{
			Fd:     int32(fd),
			Events: unix.POLLOUT,
		},
	}

	n, err := unix.Poll(pollfds, 2000)

	if err != nil || n == 0 {
		b.MarkFailure()
		return
	}

	revents := pollfds[0].Revents

	if revents&(unix.POLLERR|unix.POLLHUP) != 0 {
		b.MarkFailure()
		return
	}

	soerr, err := unix.GetsockoptInt(
		fd,
		unix.SOL_SOCKET,
		unix.SO_ERROR,
	)

	if err != nil || soerr != 0 {
		b.MarkFailure()
		return
	}

	b.MarkSuccess(time.Since(start))
}

func (b *Backend) MarkFailure() {
	failures := b.ConsecutiveFailures.Add(1)

	b.ConsecutiveSuccesses.Store(0)

	state := BackendState(
		b.State.Load(),
	)

	switch {

	case failures >= 4:
		b.SetState(Unhealthy)

	case failures >= 2 &&
		state == Healthy:

		b.SetState(Suspect)
	}
}

func (b *Backend) MarkSuccess(
	latency time.Duration,
) {
	b.ConsecutiveFailures.Store(0)

	successes := b.ConsecutiveSuccesses.Add(1)

	b.ObserveLatency(latency)

	b.AvgLatency.Store(
		latency.Milliseconds(),
	)

	state := BackendState(
		b.State.Load(),
	)

	switch state {

	case Suspect:
		if successes >= 2 {
			b.SetState(Healthy)
		}

	case Unhealthy:
		if successes >= 2 {
			b.SetState(Recovering)
		}

	case Recovering:
		if successes >= 5 {
			b.SetState(Healthy)
		}
	}
}

func isConnected(fd int) bool {
	errno, err := unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_ERROR)
	if err != nil {
		return false
	}
	return errno == 0
}

func (b *Backend) Dial() (int, error) {
	fd, err := unix.Socket(
		b.Family,
		unix.SOCK_STREAM|unix.SOCK_NONBLOCK,
		0,
	)
	if err != nil {
		return 0, err
	}

	err = unix.Connect(fd, b.SockAddr)
	log.Printf(
		"CONNECT fd=%d addr=%s err=%v",
		fd,
		b.Address,
		err,
	)

	if err != nil && err != unix.EINPROGRESS {
		unix.Close(fd)
		return 0, err
	}

	return fd, nil
}

//TODO: pool connection system

func (b *Backend) InitPool(maxIdle int) {
	b.pool = NewBackendPool(b, maxIdle)
}

func NewBackendPool(b *Backend, maxIdle int) *BackendPool {
	return &BackendPool{
		backend: b,
		idle:    make(chan int, maxIdle),
	}
}

func (p *BackendPool) Acquire() (int, error) {
	for {
		select {
		case fd := <-p.idle:
			if isConnected(fd) {
				p.backend.ActiveConnections.Add(1)
				return fd, nil
			}
			unix.Close(fd)

		default:
			state := BackendState(p.backend.State.Load())

			if state == Unhealthy || state == Draining {
				return 0, errors.New("backend unavailable")
			}

			fd, err := p.dial()
			if err != nil {
				return 0, err
			}

			p.backend.ActiveConnections.Add(1)
			return fd, nil
		}
	}
}

func (p *BackendPool) Release(fd int) {
	state := BackendState(p.backend.State.Load())

	if state == Draining || state == Unhealthy {
		unix.Close(fd)
		return
	}

	if !isConnected(fd) {
		unix.Close(fd)
		return
	}

	select {
	case p.idle <- fd:
	default:
		unix.Close(fd)
	}
}

func (p *BackendPool) dial() (int, error) {
	fd, err := unix.Socket(p.backend.Family, unix.SOCK_STREAM, 0)
	if err != nil {
		return 0, err
	}

	if err := unix.SetNonblock(fd, true); err != nil {
		unix.Close(fd)
		return 0, err
	}

	err = unix.Connect(fd, p.backend.SockAddr)
	if err != nil && err != unix.EINPROGRESS {
		unix.Close(fd)
		p.backend.MarkFailure()
		return 0, err
	}

	return fd, nil
}