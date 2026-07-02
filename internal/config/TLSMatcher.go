package config

import (
	"strings"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
)

type TLSMatcher struct {
	SNI  string
	ALPN []string
}

func (t *TLSMatcher) Match(r *inspector.RouteInfo) bool {
	if r.Protocol != inspector.TLSProto {
		return false
	}

	name := strings.ToLower(r.Host)
	sni := strings.ToLower(t.SNI)

	// SNI check
	if sni != "" {
		if strings.HasPrefix(sni, "*.") {
			suffix := strings.TrimPrefix(sni, "*.")

			if !(name == suffix || strings.HasSuffix(name, "."+suffix)) {
				return false
			}
		} else if name != sni {
			return false
		}
	}

	// ALPN check
	if len(t.ALPN) > 0 {
		match := false

		for _, expected := range t.ALPN {
			for _, actual := range r.ALPN {
				if expected == actual {
					match = true
					break
				}
			}

			if match {
				break
			}
		}

		if !match {
			return false
		}
	}

	return true
}
