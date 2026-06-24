package config

import (
	"fmt"
    "strings"
    "log"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport"
)

func NewPicker(
    listener Listener,
    clusters map[string]balancers.Balancer,
) (*transport.Picker, error) {

    p := &transport.Picker{
        ByALPN: make(map[string]balancers.Balancer),
        ExactHosts: make(map[string]balancers.Balancer),
    }

    for _, rule := range listener.Routing.Rules {
        b, ok := clusters[rule.Cluster]
        if !ok {
            return nil, fmt.Errorf(
                "unknown cluster %q", rule.Cluster,
            )
        }

        switch {
		case rule.Host == "*":
			p.DefaultBalancer = b

		case strings.HasPrefix(rule.Host, "*."):
			p.WildcardHosts = append(
				p.WildcardHosts,
				transport.WildcardRule{
					Suffix:   strings.TrimPrefix(rule.Host, "*"),
					Balancer: b,
				},
			)

		case rule.Host != "":
			p.ExactHosts[rule.Host] = b
		}

        for _, alpn := range rule.ALPN {
            p.ByALPN[alpn] = b
        }

        if rule.Default {
            p.DefaultBalancer = b
        }

    }

    if p.DefaultBalancer == nil && listener.Routing.DefaultCluster != "" {
        b, ok := clusters[listener.Routing.DefaultCluster]
        if !ok {
            return nil, fmt.Errorf(
                "unknown default cluster %q",
                listener.Routing.DefaultCluster,
            )
        }

        p.DefaultBalancer = b
    }

    log.Printf(
        "picker created default=%T cluster=%q",
        p.DefaultBalancer,
        listener.Routing.DefaultCluster,
    )

    return p, nil
}

