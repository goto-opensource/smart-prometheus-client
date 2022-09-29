package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type GaugeOpts struct {
	prometheus.GaugeOpts
	// ExpirationDelay is the maximum times a metrics keeps beeing collected when it not accessed/updated anymore.
	// It is only applicable to vector of metrics and zero value means infinite expiration time.
	ExpirationDelay time.Duration
}

func createGaugeMetricOpts(opts GaugeOpts) metricOpts {
	initialMetric := func(metric prometheus.Metric, labelValues []string) prometheus.Metric {
		// warmup mechanism is only useful for counter
		// for Gauge we disable it returning the metric itself as initial value
		return metric
	}
	return metricOpts{InitialMetric: initialMetric, ExpirationDelay: opts.ExpirationDelay}
}

// Note: for Gauge we don't need WarmUp so we do not provide constructor for single Metric
// (since it would not add any value to the existing prometheus.NewGauge(..))

// NewGaugeVec created a new vector of [prometheus.Gauge] metrics with expiration features.
func NewGaugeVec(opts GaugeOpts, labelNames []string) *MetricVec[prometheus.Gauge] {
	promVecFactory := func(labelNames []string) *prometheus.MetricVec {
		gaugeVec := prometheus.NewGaugeVec(opts.GaugeOpts, labelNames)
		return gaugeVec.MetricVec
	}
	return newMetricVec[prometheus.Gauge](promVecFactory, createGaugeMetricOpts(opts), labelNames)
}
