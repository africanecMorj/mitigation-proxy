package prometheus

import (
	"github.com/africanecMorj/mitigation-proxy.git/internal/health"

	prom "github.com/prometheus/client_golang/prometheus"
)

type Runtime interface {
	Clusters(func(cluster string, b *health.Backend))
}

type Collector struct {
	rt Runtime

	accepted *prom.Desc
	active *prom.Desc
	closed *prom.Desc

	latency *prom.Desc
	ttfb    *prom.Desc

	success *prom.Desc
	failure *prom.Desc

	bytesIn  *prom.Desc
	bytesOut *prom.Desc

	weight *prom.Desc

	totalDrains *prom.Desc

	state *prom.Desc
}

func NewCollector(rt Runtime) *Collector {
	return &Collector{
		rt: rt,

		accepted: prom.NewDesc(
			"proxy_backend_accepted_connections",
			"Accepted backend connections",
			[]string{"cluster", "backend"},
			nil,
		),

		active: prom.NewDesc(
			"proxy_backend_active_connections",
			"Current active backend connections",
			[]string{"cluster", "backend"},
			nil,
		),

		closed: prom.NewDesc(
			"proxy_backend_closed_connections",
			"Closed backend connections",
			[]string{"cluster", "backend"},
			nil,
		),

		totalDrains: prom.NewDesc(
			"proxy_backend_total_drains",
			"Total backend drains counter",
			[]string{"cluster", "backend"},
			nil,
		),


		latency: prom.NewDesc(
			"proxy_backend_latency_seconds",
			"Backend EWMA latency",
			[]string{"cluster", "backend"},
			nil,
		),

		ttfb: prom.NewDesc(
			"proxy_backend_ttfb_seconds",
			"Backend TTFB",
			[]string{"cluster", "backend"},
			nil,
		),

		success: prom.NewDesc(
			"proxy_backend_success_total",
			"Backend successful requests",
			[]string{"cluster", "backend"},
			nil,
		),

		failure: prom.NewDesc(
			"proxy_backend_failure_total",
			"Backend failed requests",
			[]string{"cluster", "backend"},
			nil,
		),

		bytesIn: prom.NewDesc(
			"proxy_backend_bytes_received_total",
			"Backend received bytes",
			[]string{"cluster", "backend"},
			nil,
		),

		bytesOut: prom.NewDesc(
			"proxy_backend_bytes_sent_total",
			"Backend sent bytes",
			[]string{"cluster", "backend"},
			nil,
		),

		weight: prom.NewDesc(
			"proxy_backend_weight",
			"Backend weight",
			[]string{"cluster", "backend"},
			nil,
		),

		state: prom.NewDesc(
			"proxy_backend_state",
			"Backend state",
			[]string{"cluster", "backend", "state"},
			nil,
		),
	}
}

func (c *Collector) Describe(ch chan<- *prom.Desc) {
	ch <- c.active
	ch <- c.accepted
	ch <- c.closed
	ch <- c.latency
	ch <- c.ttfb
	ch <- c.success
	ch <- c.failure
	ch <- c.bytesIn
	ch <- c.bytesOut
	ch <- c.weight
	ch <- c.totalDrains
	ch <- c.state
}

func (c *Collector) Collect(ch chan<- prom.Metric) {

	c.rt.Clusters(func(cluster string, b *health.Backend) {

		ch <- prom.MustNewConstMetric(
			c.active,
			prom.GaugeValue,
			float64(b.ActiveConnections.Load()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.accepted,
			prom.GaugeValue,
			float64(b.Requests.Load()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.closed,
			prom.GaugeValue,
			float64(b.TotalClosed()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.totalDrains,
			prom.GaugeValue,
			float64(b.TotalDrains()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.latency,
			prom.GaugeValue,
			float64(b.Latency())/1e9,
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.ttfb,
			prom.GaugeValue,
			float64(b.TTFBValue())/1e9,
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.success,
			prom.CounterValue,
			float64(b.Successes()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.failure,
			prom.CounterValue,
			float64(b.Failures()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.bytesIn,
			prom.CounterValue,
			float64(b.BytesReceivedValue()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.bytesOut,
			prom.CounterValue,
			float64(b.BytesSentValue()),
			cluster,
			b.Address,
		)

		ch <- prom.MustNewConstMetric(
			c.weight,
			prom.GaugeValue,
			float64(b.WeightValue()),
			cluster,
			b.Address,
		)

		value := 0.0
		if b.StateValue() == health.Healthy {
			value = 1
		}

		ch <- prom.MustNewConstMetric(
			c.state,
			prom.GaugeValue,
			value,
			cluster,
			b.Address,
			b.StateValue().String(),
		)
	})
}