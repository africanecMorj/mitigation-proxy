package config

import (
	"fmt"

	"github.com/africanecMorj/mitigation-proxy.git/internal/balancers"
	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
    
)

type matchInput struct {
	host string
	alpn []string
	meta map[string]string
}

type MatchRule struct {
    Matcher Matcher 
    Balancer balancers.Balancer
}

type Selector struct {
    Rules []MatchRule
    DefaultBalancer balancers.Balancer
}

func NewPicker(
	listener Listener,
	clusters map[string]balancers.Balancer,
) (*Selector ,error) {


	p := &Selector{}


	proto := parseProto(
		listener.Routing.Type,
	)


	for _, cfgRule := range listener.Routing.Rules {


		b,ok := clusters[cfgRule.Cluster]


		if !ok {
			return nil,fmt.Errorf(
				"unknown cluster %q",
				cfgRule.Cluster,
			)
		}



		input := matchInput{
			host: cfgRule.Host,
			alpn: cfgRule.ALPN,
			meta: cfgRule.Metadata,
		}



		p.Rules = append(
			p.Rules,
			MatchRule{

				Matcher:buildMatcher(
					proto,
					input,
				),

				Balancer:b,
			},
		)

        if listener.Routing.DefaultCluster != "" {
            b, ok := clusters[listener.Routing.DefaultCluster]
            if !ok {
                return nil, fmt.Errorf(
                    "unknown default cluster %q",
                    listener.Routing.DefaultCluster,
                )
            }
            p.DefaultBalancer = b
        }

	}



	return p, nil
}

func (p *Selector) SelectBackend(info *inspector.RouteInfo) (balancers.Balancer, error) {
	for _, r := range p.Rules {
		ok := r.Matcher.Match(info)

		if ok {
			return r.Balancer, nil
		}
	}

	if p.DefaultBalancer != nil {
		return p.DefaultBalancer, nil
	}

	return nil, fmt.Errorf("no backend matched")
}

