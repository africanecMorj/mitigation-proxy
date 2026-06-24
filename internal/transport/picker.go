package transport

import (
	"errors"
	"strings"
	"log"
	"fmt"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"
	"github.com/africanecMorj/mitigation-proxy.git/internal/ratelimit"
)

type Picker struct {
    ExactHosts      map[string]balancers.Balancer
    WildcardHosts   []WildcardRule
    ByALPN          map[string]balancers.Balancer
    DefaultBalancer balancers.Balancer
    Meta map[string]string
}


type WildcardRule struct {
    Suffix   string
    Balancer balancers.Balancer
}

func (p *Picker) Pick(
    sni string,
    host string,
    alpn []string,
    ip string,
) (*health.Backend, int, error) {

    log.Printf(
        "Pick sni=%q host=%q alpn=%v default=%T",
        sni,
        host,
        alpn,
        p.DefaultBalancer,
    )

    if !ratelimit.GetLimiter(ip).Allow() {
        return nil, 0, errors.New("rate limited")
    }

    bl := p.SelectBackend(sni, host, alpn)
    if bl == nil {
        return nil, 0, fmt.Errorf(
            "no balancer found for sni=%q host=%q alpn=%v",
            sni,
            host,
            alpn,
        )
    }

    b := bl.Next()
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


func (p *Picker) SelectBackend(
    sni string,
    host string,
    alpn []string,
    meta map[string]string,
) balancers.Balancer {

    if b := p.selectByName(strings.ToLower(sni)); b != nil {
        return b
    }

     if b := p.selectByName(strings.ToLower(host)); b != nil {
        return b
    }

    for _, a := range alpn {
        if b, ok := p.ByALPN[a]; ok {
            return b
        }
    }

    matched := true

     for k,v := range p.Meta {

        if meta[k] != v {
            matched=false
            break
        }
    }


    if matched {
        return rule.Balancer
    }
    

    return p.DefaultBalancer
}

func (p *Picker) selectByName(name string) balancers.Balancer {
    if name == "" {
        return nil
    }

    if b, ok := p.ExactHosts[name]; ok {
        return b
    }

    var best balancers.Balancer
    longest := 0

    for _, w := range p.WildcardHosts {
        suffix := strings.TrimPrefix(w.Suffix, ".")

        if name == suffix ||
            strings.HasSuffix(name, "."+suffix) {

            if len(suffix) > longest {
                longest = len(suffix)
                best = w.Balancer
            }
        }
    }

    return best
}