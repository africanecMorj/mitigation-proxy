package config

import (
    "github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
    "strings"
)

type HTTPMatcher struct {
    Host string
}

func (h *HTTPMatcher) Match(
    r *inspector.RouteInfo,
) bool {

    if r.Protocol != inspector.HTTPProto {
		return false
	}


    host := strings.ToLower(h.Host)
	name := strings.ToLower(r.Host)
	
	if strings.HasPrefix(host, "*.") {
		suffix := strings.TrimPrefix(host, "*")
		suffix = strings.TrimPrefix(suffix, ".")

        if name == suffix ||
            strings.HasSuffix(name, "."+suffix) {
			return true
        }
	
	}
   
	
	if host != "" && host == name {
		return true
	} 

    return false
}