package config

import "github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector" 

type Matcher interface {
	Match(*inspector.RouteInfo) bool
}