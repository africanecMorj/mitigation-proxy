package config

import (
	"testing"

	"github.com/africanecMorj/mitigation-proxy.git/internal/transport/inspector"
	"github.com/stretchr/testify/assert"
)

func TestPostgresMatcher(t *testing.T) {

	t.Run("all metadata matches", func(t *testing.T) {
		m := &PostgresMatcher{
			Params: map[string]string{
				"database": "app",
				"user":     "postgres",
			},
		}

		assert.True(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.PostgresProto,
			Meta: map[string]string{
				"database": "app",
				"user":     "postgres",
			},
		}))
	})

	t.Run("missing metadata", func(t *testing.T) {
		m := &PostgresMatcher{
			Params: map[string]string{
				"user": "postgres",
			},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.PostgresProto,
			Meta:     map[string]string{},
		}))
	})

	t.Run("wrong metadata", func(t *testing.T) {
		m := &PostgresMatcher{
			Params: map[string]string{
				"user": "postgres",
			},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.PostgresProto,
			Meta: map[string]string{
				"user": "root",
			},
		}))
	})

	t.Run("wrong protocol", func(t *testing.T) {
		m := &PostgresMatcher{
			Params: map[string]string{
				"user": "postgres",
			},
		}

		assert.False(t, m.Match(&inspector.RouteInfo{
			Protocol: inspector.HTTPProto,
			Meta: map[string]string{
				"user": "postgres",
			},
		}))
	})
}