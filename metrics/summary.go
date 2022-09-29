package metrics

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type SummaryOpts struct {
	prometheus.SummaryOpts
	// WarmUpDuration represents the time during which metrics are collected
	// with their initial value instead of their actual value, starting at the first collection.
	// The warmup period start at the first collection and ends after WarmUpDuration.
	WarmUpDuration time.Duration
	// ExpirationDelay is the maximum times a metrics keeps beeing collected when it not accessed/updated anymore.
	// It is only applicable to vector of metrics and zero value means infinite expiration time.
	ExpirationDelay time.Duration
}

func createSummaryMetricOpts(opts SummaryOpts) metricOpts {
	initialQuantiles := make(map[float64]float64, len(opts.Objectives))
	for _, val := range opts.Objectives {
		initialQuantiles[val] = 0
	}
	initialMetric := func(metric prometheus.Metric, labelValues []string) prometheus.Metric {
		return prometheus.MustNewConstSummary(metric.Desc(), 0, 0, initialQuantiles, labelValues...)
	}
	return metricOpts{InitialMetric: initialMetric, WarmUpDuration: opts.WarmUpDuration, ExpirationDelay: opts.ExpirationDelay}
}

type summary struct {
	prometheus.Summary
	*singleCollector
}

// NewSummary created a new [prometheus.Summary] metric with the Warmup feature.
func NewSummary(opts SummaryOpts) prometheus.Summary {
	promSummary := prometheus.NewSummary(opts.SummaryOpts)
	collector := newSingleCollector(promSummary, createSummaryMetricOpts(opts))
	return &summary{promSummary, collector}
}

// Collect implements [prometheus.Collector].
func (s *summary) Collect(ch chan<- prometheus.Metric) {
	s.singleCollector.Collect(ch)
}

// NewSummaryVec created a new vector of [prometheus.Summary] metrics with the Warmup and expiration features.
func NewSummaryVec(opts SummaryOpts, labelNames []string) *MetricVec[prometheus.Summary] {
	promVecFactory := func(labelNames []string) *prometheus.MetricVec {
		summaryVec := prometheus.NewSummaryVec(opts.SummaryOpts, labelNames)
		return summaryVec.MetricVec
	}
	return newMetricVec[prometheus.Summary](promVecFactory, createSummaryMetricOpts(opts), labelNames)
}
