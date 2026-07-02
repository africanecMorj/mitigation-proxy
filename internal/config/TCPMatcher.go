package config

import (
    "github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)

type TCPMatcher struct {}

func (_ *TCPMatcher) Match (r *inspector.RouteInfo) bool {
	if r.Protocol != inspector.RawTCPProto {
		return false
	}

	return true
}

