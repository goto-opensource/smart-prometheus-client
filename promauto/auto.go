// Modified version of the promauto package of the Prometheus golang client
//
//
// Original file
// - https://github.com/prometheus/client_golang/blob/main/prometheus/promauto/auto.go
//
// Original file Header:
//
// Copyright 2018 The Prometheus Authors
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package promauto provides alternative constructors for the fundamental
// Prometheus metric types and their …Vec ( …Func variants or not implemented)
//
// This is a modified version of the original [prometheus.promauto] package that adds supports for "smart metrics" of this library
package promauto

import (
	"time"

	"github.com/goto-opensource/smart-prometheus-client/metrics"
	"github.com/prometheus/client_golang/prometheus"
)

type SmartMetricOpts struct {
	// WarmUpDuration represents the time during which metrics are collected
	// with their initial value instead of their actual value, starting at the first collection.
	// The warmup period start at the first collection and ends after WarmUpDuration.
	WarmUpDuration time.Duration
	// ExpirationDelay is the maximum times a metrics keeps beeing collected when it not accessed/updated anymore.
	// It is only applicable to vector of metrics and zero value means infinite expiration time.
	ExpirationDelay time.Duration
}

// DefaultOptions are the default 'Smart metrics' options used by all the package level NewXXX functions
var DefaultOptions SmartMetricOpts

// NewCounter works like the function of the same name in the [metrics] package
// but it automatically registers the Counter with the
// prometheus.DefaultRegisterer. If the registration fails, NewCounter panics.
//
// It used the default options DefaultOptions
func NewCounter(opts prometheus.CounterOpts) prometheus.Counter {
	return With(prometheus.DefaultRegisterer).NewCounter(opts)
}

// NewCounterVec works like the function of the same name in the metrics
// package but it automatically registers the CounterVec with the
// prometheus.DefaultRegisterer. If the registration fails, NewCounterVec
// panics.
//
// It used the default options DefaultOptions
func NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *metrics.MetricVec[prometheus.Counter] {
	return With(prometheus.DefaultRegisterer).NewCounterVec(opts, labelNames)
}

// NewGauge works like the function of the same name in the prometheus package
// but it automatically registers the Gauge with the
// prometheus.DefaultRegisterer. If the registration fails, NewGauge panics.
func NewGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	return With(prometheus.DefaultRegisterer).NewGauge(opts)
}

// NewGaugeVec works like the function of the same name in the metrics
// package but it automatically registers the GaugeVec with the
// prometheus.DefaultRegisterer. If the registration fails, NewGaugeVec panics.
//
// It used the default options DefaultOptions
func NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *metrics.MetricVec[prometheus.Gauge] {
	return With(prometheus.DefaultRegisterer).NewGaugeVec(opts, labelNames)
}

// NewSummary works like the function of the same name in the metrics package
// but it automatically registers the Summary with the
// prometheus.DefaultRegisterer. If the registration fails, NewSummary panics.
//
// It used the default options DefaultOptions
func NewSummary(opts prometheus.SummaryOpts) prometheus.Summary {
	return With(prometheus.DefaultRegisterer).NewSummary(opts)
}

// NewSummaryVec works like the function of the same name in the metrics
// package but it automatically registers the SummaryVec with the
// prometheus.DefaultRegisterer. If the registration fails, NewSummaryVec
// panics.
//
// It used the default options DefaultOptions
func NewSummaryVec(opts prometheus.SummaryOpts, labelNames []string) *metrics.MetricVec[prometheus.Summary] {
	return With(prometheus.DefaultRegisterer).NewSummaryVec(opts, labelNames)
}

// NewHistogram works like the function of the same name in the metrics
// package but it automatically registers the Histogram with the
// prometheus.DefaultRegisterer. If the registration fails, NewHistogram panics.
//
// It used the default options DefaultOptions
func NewHistogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	return With(prometheus.DefaultRegisterer).NewHistogram(opts)
}

// NewHistogramVec works like the function of the same name in the metrics
// package but it automatically registers the HistogramVec with the
// prometheus.DefaultRegisterer. If the registration fails, NewHistogramVec
// panics.
//
// It used the default options DefaultOptions
func NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *metrics.MetricVec[prometheus.Histogram] {
	return With(prometheus.DefaultRegisterer).NewHistogramVec(opts, labelNames)
}

// Factory provides factory methods to create Collectors that are automatically
// registered with a Registerer. Create a Factory with the With function,
// providing a Registerer to auto-register created Collectors with. The zero
// value of a Factory creates Collectors that are not registered with any
// Registerer. All methods of the Factory panic if the registration fails.
//
// It uses the DefaultOptions as configuration for the created metrics collectors.
// Additionally you can use the WithOptions function to provides your own options instead
// of the default ones.
type Factory struct {
	r    prometheus.Registerer
	opts SmartMetricOpts
}

// With creates a Factory using the provided Registerer for registration of the
// created Collectors. If the provided Registerer is nil, the returned Factory
// creates Collectors that are not registered with any Registerer.
func With(r prometheus.Registerer) Factory { return Factory{r, DefaultOptions} }

// WithOptions creates a Factory using SmartMetricOpts to creates new metrics and that registers
// the created Collectors to the given Registerer.
//
// If the provided Registerer is nil, the returned Factory
// creates Collectors that are not registered with any Registerer.
//
// The returned Factory creates the new collector with the configurations provided by opts.
func WithOptions(r prometheus.Registerer, opts SmartMetricOpts) Factory { return Factory{r, opts} }

// NewCounter works like the function of the same name in the metrics package
// but it automatically registers the Counter with the Factory's Registerer.
func (f Factory) NewCounter(opts prometheus.CounterOpts) prometheus.Counter {
	c := metrics.NewCounter(metrics.CounterOpts{CounterOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay})
	if f.r != nil {
		f.r.MustRegister(c)
	}
	return c
}

// NewCounterVec works like the function of the same name in the metrics
// package but it automatically registers the CounterVec with the Factory's
// Registerer.
func (f Factory) NewCounterVec(opts prometheus.CounterOpts, labelNames []string) *metrics.MetricVec[prometheus.Counter] {
	c := metrics.NewCounterVec(metrics.CounterOpts{CounterOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay}, labelNames)
	if f.r != nil {
		f.r.MustRegister(c)
	}
	return c
}

// NewGauge works like the function of the same name in the prometheus package
// but it automatically registers the Gauge with the Factory's Registerer.
func (f Factory) NewGauge(opts prometheus.GaugeOpts) prometheus.Gauge {
	g := prometheus.NewGauge(opts)
	if f.r != nil {
		f.r.MustRegister(g)
	}
	return g
}

// NewGaugeVec works like the function of the same name in the metrics
// package but it automatically registers the GaugeVec with the Factory's
// Registerer.
func (f Factory) NewGaugeVec(opts prometheus.GaugeOpts, labelNames []string) *metrics.MetricVec[prometheus.Gauge] {
	g := metrics.NewGaugeVec(metrics.GaugeOpts{GaugeOpts: opts, ExpirationDelay: f.opts.ExpirationDelay}, labelNames)
	if f.r != nil {
		f.r.MustRegister(g)
	}
	return g
}

// NewSummary works like the function of the same name in the metrics package
// but it automatically registers the Summary with the Factory's Registerer.
func (f Factory) NewSummary(opts prometheus.SummaryOpts) prometheus.Summary {
	s := metrics.NewSummary(metrics.SummaryOpts{SummaryOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay})
	if f.r != nil {
		f.r.MustRegister(s)
	}
	return s
}

// NewSummaryVec works like the function of the same name in the metrics
// package but it automatically registers the SummaryVec with the Factory's
// Registerer.
func (f Factory) NewSummaryVec(opts prometheus.SummaryOpts, labelNames []string) *metrics.MetricVec[prometheus.Summary] {
	s := metrics.NewSummaryVec(metrics.SummaryOpts{SummaryOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay}, labelNames)
	if f.r != nil {
		f.r.MustRegister(s)
	}
	return s
}

// NewHistogram works like the function of the same name in the metrics
// package but it automatically registers the Histogram with the Factory's
// Registerer.
func (f Factory) NewHistogram(opts prometheus.HistogramOpts) prometheus.Histogram {
	h := metrics.NewHistogram(metrics.HistogramOpts{HistogramOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay})
	if f.r != nil {
		f.r.MustRegister(h)
	}
	return h
}

// NewHistogramVec works like the function of the same name in the metrics
// package but it automatically registers the HistogramVec with the Factory's
// Registerer.
func (f Factory) NewHistogramVec(opts prometheus.HistogramOpts, labelNames []string) *metrics.MetricVec[prometheus.Histogram] {
	h := metrics.NewHistogramVec(metrics.HistogramOpts{HistogramOpts: opts, WarmUpDuration: f.opts.WarmUpDuration, ExpirationDelay: f.opts.ExpirationDelay}, labelNames)
	if f.r != nil {
		f.r.MustRegister(h)
	}
	return h
}
