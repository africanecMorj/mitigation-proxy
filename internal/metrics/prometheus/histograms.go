package prometheus

import (
	"time"

	prom "github.com/prometheus/client_golang/prometheus"
)

type Histograms struct {
	Connect *prom.HistogramVec
	TTFB    *prom.HistogramVec
}

func NewHistograms(reg prom.Registerer) *Histograms {
	h := &Histograms{
		Connect: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "proxy_backend_connect_duration_seconds",
				Help:    "Backend TCP connect duration",
				Buckets: prom.DefBuckets,
			},
			[]string{"cluster", "backend"},
		),

		TTFB: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "proxy_backend_ttfb_seconds",
				Help:    "Backend time to first byte",
				Buckets: prom.DefBuckets,
			},
			[]string{"cluster", "backend"},
		),
	}

	reg.MustRegister(h.Connect)
	reg.MustRegister(h.TTFB)

	return h
}

func (h *Histograms) ObserveConnect(cluster, backend string, d time.Duration) {
	h.Connect.WithLabelValues(cluster, backend).Observe(d.Seconds())
}

func (h *Histograms) ObserveTTFB(cluster, backend string, d time.Duration) {
	h.TTFB.WithLabelValues(cluster, backend).Observe(d.Seconds())
}