package config

import (
	"testing"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/stretchr/testify/assert"
)

func TestTLSMatcher(t *testing.T) {

	t.Run("exact sni", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "example.com",
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "example.com",
		}))
	})

	t.Run("case insensitive sni", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "Example.COM",
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "EXAMPLE.com",
		}))
	})

	t.Run("wildcard sni", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "*.example.com",
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "api.example.com",
		}))
	})

	t.Run("wildcard nested subdomain", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "*.example.com",
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "v1.api.example.com",
		}))
	})

	t.Run("wildcard root domain", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "*.example.com",
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "example.com",
		}))
	})

	t.Run("alpn only", func(t *testing.T) {
		m := &TLSMatcher{
			ALPN: []string{"h2"},
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			ALPN:     []string{"http/1.1", "h2"},
		}))
	})

	t.Run("sni and alpn", func(t *testing.T) {
		m := &TLSMatcher{
			SNI:  "example.com",
			ALPN: []string{"h2"},
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "example.com",
			ALPN:     []string{"h2"},
		}))
	})

	t.Run("sni matches but alpn does not", func(t *testing.T) {
		m := &TLSMatcher{
			SNI:  "example.com",
			ALPN: []string{"h2"},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "example.com",
			ALPN:     []string{"http/1.1"},
		}))
	})

	t.Run("alpn matches but sni does not", func(t *testing.T) {
		m := &TLSMatcher{
			SNI:  "example.com",
			ALPN: []string{"h2"},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "google.com",
			ALPN:     []string{"h2"},
		}))
	})

	t.Run("wrong protocol", func(t *testing.T) {
		m := &TLSMatcher{
			SNI: "example.com",
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "example.com",
		}))
	})

	t.Run("no match", func(t *testing.T) {
		m := &TLSMatcher{
			SNI:  "example.com",
			ALPN: []string{"h2"},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.TLSProto,
			Host:     "google.com",
			ALPN:     []string{"http/1.1"},
		}))
	})
}