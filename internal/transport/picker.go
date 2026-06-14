package transport

import (
	"errors"
	"strings"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/ratelimit"
)

type Picker struct {
    ExactHosts      map[string]balancers.Balancer
    WildcardHosts   []WildcardRule
    ByALPN          map[string]balancers.Balancer
    DefaultBalancer balancers.Balancer
}

type WildcardRule struct {
    Suffix   string
    Balancer balancers.Balancer
}

func (p *Picker) Pick(
	host string,
	alpn []string,
	ip string,
) (*health.Backend, int, error) {

	if !ratelimit.GetLimiter(ip).Allow() {
		return nil, 0, errors.New("rate limited")
	}

	bl := p.SelectBackend(host, alpn)

	b := bl.Next()

	fd, err := b.Dial()
	if err != nil {
		b.MarkFailure()
		return nil, 0, err
	}

	return b, fd, nil
}

func (p *Picker) SelectBackend(host string, alpn []string) balancers.Balancer {
    host = strings.ToLower(host)

	if b, ok := p.ExactHosts[host]; ok {
        return b
    }

    var best balancers.Balancer
    longest := 0

    for _, w := range p.WildcardHosts {
        if host == w.Suffix ||
   			strings.HasSuffix(host, "."+w.Suffix) {
				
            if len(w.Suffix) > longest {
                longest = len(w.Suffix)
                best = w.Balancer
            }
        }
    }

    if best != nil {
        return best
    }

	for _, a := range alpn {
		if b, ok := p.ByALPN[a]; ok {
        	return b
    	}
	}
	

    return p.DefaultBalancer
}