package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type CounterOpts struct {
	prometheus.CounterOpts
	// WarmUpDuration represents the time during which metrics are collected
	// with their initial value instead of their actual value, starting at the first collection.
	// The warmup period start at the first collection and ends after WarmUpDuration.
	WarmUpDuration time.Duration
	// ExpirationDelay is the maximum times a metrics keeps beeing collected when it not accessed/updated anymore.
	// It is only applicable to vector of metrics and zero value means infinite expiration time.
	ExpirationDelay time.Duration
}

func createCounterMetricOpts(opts CounterOpts) metricOpts {
	initialMetric := func(metric prometheus.Metric, labelValues []string) prometheus.Metric {
		return prometheus.MustNewConstMetric(metric.Desc(), prometheus.CounterValue, 0, labelValues...)
	}
	return metricOpts{InitialMetric: initialMetric, WarmUpDuration: opts.WarmUpDuration, ExpirationDelay: opts.ExpirationDelay}
}

type counter struct {
	prometheus.Counter
	*singleCollector
}

// NewCounter created a new [prometheus.Counter] metric with the Warmup feature.
func NewCounter(opts CounterOpts) prometheus.Counter {
	promCounter := prometheus.NewCounter(opts.CounterOpts)
	collector := newSingleCollector(promCounter, createCounterMetricOpts(opts))
	return &counter{promCounter, collector}
}

// Collect implements [prometheus.Collector].
func (c *counter) Collect(ch chan<- prometheus.Metric) {
	c.singleCollector.Collect(ch)
}

// NewCounterVec created a new vector of [prometheus.Counter] metrics with the Warmup and expiration features.
func NewCounterVec(opts CounterOpts, labelNames []string) *MetricVec[prometheus.Counter] {
	promVecFactory := func(labelNames []string) *prometheus.MetricVec {
		counterVec := prometheus.NewCounterVec(opts.CounterOpts, labelNames)
		return counterVec.MetricVec
	}
	return newMetricVec[prometheus.Counter](promVecFactory, createCounterMetricOpts(opts), labelNames)
}
