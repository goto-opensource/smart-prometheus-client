package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type HistogramOpts struct {
	prometheus.HistogramOpts
	// WarmUpDuration represents the time during which metrics are collected
	// with their initial value instead of their actual value, starting at the first collection.
	// The warmup period start at the first collection and ends after WarmUpDuration.
	WarmUpDuration time.Duration
	// ExpirationDelay is the maximum times a metrics keeps beeing collected when it not accessed/updated anymore.
	// It is only applicable to vector of metrics and zero value means infinite expiration time.
	ExpirationDelay time.Duration
}

func createHistogramMetricOpts(opts HistogramOpts) metricOpts {
	initialBuckets := make(map[float64]uint64, len(opts.Buckets))
	for _, val := range opts.Buckets {
		initialBuckets[val] = 0
	}
	initialMetric := func(metric prometheus.Metric, labelValues []string) prometheus.Metric {
		return prometheus.MustNewConstHistogram(metric.Desc(), 0, 0, initialBuckets, labelValues...)
	}
	return metricOpts{InitialMetric: initialMetric, WarmUpDuration: opts.WarmUpDuration, ExpirationDelay: opts.ExpirationDelay}
}

type histogram struct {
	prometheus.Histogram
	*singleCollector
}

// NewHistogram created a new [prometheus.Histogram] metric with the Warmup feature.
func NewHistogram(opts HistogramOpts) prometheus.Histogram {
	promHistogram := prometheus.NewHistogram(opts.HistogramOpts)
	collector := newSingleCollector(promHistogram, createHistogramMetricOpts(opts))
	return &histogram{promHistogram, collector}
}

// Collect implements [prometheus.Collector].
func (h *histogram) Collect(ch chan<- prometheus.Metric) {
	h.singleCollector.Collect(ch)
}

// NewHistogramVec created a new vector of [prometheus.Histogram] metrics with the Warmup and expiration features.
func NewHistogramVec(opts HistogramOpts, labelNames []string) *MetricVec[prometheus.Histogram] {
	promVecFactory := func(labelNames []string) *prometheus.MetricVec {
		histogramVec := prometheus.NewHistogramVec(opts.HistogramOpts, labelNames)
		return histogramVec.MetricVec
	}
	return newMetricVec[prometheus.Histogram](promVecFactory, createHistogramMetricOpts(opts), labelNames)
}
