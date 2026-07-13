package transport

import (
	"errors"
	
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/ratelimit"
	"github.com/africanecMorj/mitigation-proxy.git/internal/config"
)

type Picker struct {
    *config.Selector
}

func (p *Picker) Pick(
    info *inspector.RouteInfo,
    ip string,
) (*health.Backend, int, error) {

    if !ratelimit.GetLimiter(ip).Allow() {
        return nil, 0, errors.New("rate limited")
    }

    bl, err := p.SelectBackend(info)
    if bl == nil {
        return nil, 0 , err
    }

    b := bl.Next(ip)
    if b == nil {
        return nil, 0, errors.New("backend is nil")
    }

    fd, err := b.Dial()
    if err != nil {
        b.MarkFailure()
        return nil, 0, err
    }

    return b, fd, nil
}

