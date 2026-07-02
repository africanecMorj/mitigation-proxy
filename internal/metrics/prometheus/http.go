package prometheus

import (
	"net/http"
	"context"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	prom "github.com/prometheus/client_golang/prometheus"
)

type Server struct {
	registry *prom.Registry
	server   *http.Server
}

func NewProm(addr string, rt Runtime) *Server {
	registry := prom.NewRegistry()

	registry.MustRegister(
        prom.NewProcessCollector(prom.ProcessCollectorOpts{}),
        prom.NewGoCollector(),
        NewCollector(rt),
    )

	return &Server{
		registry: registry,
		server: &http.Server{
			Addr: addr,
			Handler: promhttp.HandlerFor(
				registry,
				promhttp.HandlerOpts{},
			),
		},
	}
}

func (s *Server) Run() error {
	return s.server.ListenAndServe()
}

func (s *Server) Close(ctx context.Context) error {
	return s.server.Shutdown(ctx)
}

func Start(addr string, rt Runtime) (*Server, error) {
	s := NewProm(addr, rt)

	go func() {
		_ = s.Run()
	}()

	return s, nil
}