package config

import (
	"testing"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/stretchr/testify/assert"
)

func TestTCPMatcher(t *testing.T) {
	m := &TCPMatcher{}

	t.Run("matches tcp", func(t *testing.T) {
		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.RawTCPProto,
		}))
	})

	t.Run("rejects http", func(t *testing.T) {
		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
		}))
	})

	t.Run("rejects postgres", func(t *testing.T) {
		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.PostgresProto,
		}))
	})
}