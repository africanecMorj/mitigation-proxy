package config

import (
	"testing"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/stretchr/testify/assert"
)

func TestHTTPMatcher(t *testing.T) {

	t.Run("exact host", func(t *testing.T) {
		m := &HTTPMatcher{Host: "example.com"}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "example.com",
		}))
	})

	t.Run("case insensitive", func(t *testing.T) {
		m := &HTTPMatcher{Host: "example.com"}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "Example.Com",
		}))
	})

	t.Run("wildcard subdomain", func(t *testing.T) {
		m := &HTTPMatcher{Host: "*.example.com"}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "api.example.com",
		}))
	})

	t.Run("wildcard nested subdomain", func(t *testing.T) {
		m := &HTTPMatcher{Host: "*.example.com"}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "v1.api.example.com",
		}))
	})

	t.Run("wildcard root domain", func(t *testing.T) {
		m := &HTTPMatcher{Host: "*.example.com"}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "example.com",
		}))
	})

	t.Run("different host", func(t *testing.T) {
		m := &HTTPMatcher{Host: "example.com"}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Host:     "google.com",
		}))
	})

	t.Run("wrong protocol", func(t *testing.T) {
		m := &HTTPMatcher{Host: "example.com"}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.PostgresProto,
			Host:     "example.com",
		}))
	})
}