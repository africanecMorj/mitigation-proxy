package config

import (
    "github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)


type PostgresMatcher struct {
	Params map[string]string
}

func (p *PostgresMatcher) Match (r *inspector.RouteInfo) bool {
	if r.Protocol != inspector.PostgresProto {
		return false
	}


	for key, expected := range p.Params {


		actual, ok := r.Meta[key]

		if !ok {
			return false
		}


		if actual != expected {
			return false
		}
	}


	return true

}

