package metrics

import (
	"strconv"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

const labelLifeCycleTag = "_tag_"

type metricOpts struct {
	InitialMetric   func(metric prometheus.Metric, labelValues []string) prometheus.Metric
	WarmUpDuration  time.Duration
	ExpirationDelay time.Duration
}

type metricState uint32

const (
	stateWarmUpPending metricState = iota
	stateWarmUpOngoing
	stateWarmUpComplete
	stateExpired
)

type metricAttr struct {
	tag            string
	labelValues    []string
	state          metricState
	stateMutex     sync.Mutex
	warmUpDeadLine time.Time
	activeDeadLine time.Time
}

func (a *metricAttr) hasExpired() bool {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	return a.state == stateExpired
}

func (a *metricAttr) onCollect(warmUpDuration time.Duration) metricState {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	nowTime := nowFunc()
	if a.state == stateWarmUpPending {
		a.warmUpDeadLine = nowTime.Add(warmUpDuration)
		a.state = stateWarmUpOngoing
	} else if a.state == stateWarmUpOngoing && nowTime.After(a.warmUpDeadLine) {
		a.state = stateWarmUpComplete
	} else if a.state == stateWarmUpComplete && !a.activeDeadLine.IsZero() && nowTime.After(a.activeDeadLine) {
		a.state = stateExpired
	}
	return a.state
}

func (a *metricAttr) onAccess(expirationDelay time.Duration) {
	a.stateMutex.Lock()
	defer a.stateMutex.Unlock()
	if expirationDelay > 0 {
		a.activeDeadLine = nowFunc().Add(expirationDelay)
	}
}

type singleCollector struct {
	metric prometheus.Metric
	attr   *metricAttr
	opts   metricOpts
}

func newSingleCollector(metric prometheus.Metric, opts metricOpts) *singleCollector {
	return &singleCollector{
		metric: metric,
		attr:   &metricAttr{},
		opts:   opts,
	}
}

// Collect implements the collection process of the prometheus Collector interface.
// It handles the metrics warm-up and returns the initial value instead of the actual metric value
// till the warm-up delay has passed.
func (c *singleCollector) Collect(ch chan<- prometheus.Metric) {
	state := c.attr.onCollect(c.opts.WarmUpDuration)
	if state == stateWarmUpOngoing {
		ch <- c.opts.InitialMetric(c.metric, c.attr.labelValues)
	} else {
		ch <- c.metric
	}
}

// generateLifeCycleTag generates a value for the internal label lifeCycleTag that is added as time series labels.
// It makes sure we produce different value over the time.
func generateLifeCycleTag() string {
	return strconv.FormatInt(nowFunc().Unix(), 16)
}

// MetricVec is a generic implementation of a Vector of metrics, to bundle metrics of the same name that differ in
// their label values. It is an extension of [prometheus.MetricVec] that adds two functionalities to the vanilla prometheus.MetricVec: the metric 'warm-up'
// and automatic delete (expiration delay).
//
// The available operations are the same as [prometheus.MetricVec] expect for the Curry operation that is not implemented.
//
// You should not instantiate directly this struct
type MetricVec[M prometheus.Metric] struct {
	metricVec   *prometheus.MetricVec
	labelNames  []string
	opts        metricOpts
	metricAttrs map[prometheus.Metric]*metricAttr
	tags        *tagMap
	mutex       sync.RWMutex
}

func newMetricVec[M prometheus.Metric](vecFactory func(labelNames []string) *prometheus.MetricVec, opts metricOpts, labelNames []string) *MetricVec[M] {
	// Add the internal label labelLifeCycleTag at the end of the list of labels
	// This is an extra label managed internally to avoid label collision with expired metrics.
	allLabelNames := make([]string, len(labelNames)+1)
	copy(allLabelNames, labelNames)
	allLabelNames[len(labelNames)] = labelLifeCycleTag
	vec := vecFactory(allLabelNames)

	return &MetricVec[M]{
		metricVec:  vec,
		labelNames: allLabelNames[:len(labelNames)],
		opts:       opts,
		// using prometheus.Metric as key will only work when the underlying implementation use pointer receiver on struct
		// (the interface must be comparable). Fortunately this is the case for all basic metric types of prometheus library.
		metricAttrs: make(map[prometheus.Metric]*metricAttr),
		tags:        newTagMap(),
	}
}

// must be called holding mv.mutex.RLock or mv.mutex.Lock
func (mv *MetricVec[M]) getMetric(labelValues ...string) (prometheus.Metric, error) {
	tag, present := mv.tags.Get(labelValues)
	if !present {
		return nil, nil
	}
	metric, err := mv.metricVec.GetMetricWithLabelValues(append(labelValues, tag)...)
	if err != nil {
		return metric, err
	}
	attr := mv.metricAttrs[metric]
	if attr == nil || attr.hasExpired() {
		return nil, nil
	}
	// Schedule the expiration time
	attr.onAccess(mv.opts.ExpirationDelay)
	return metric, nil
}

// must be called holding mv.mutex.Lock
func (mv *MetricVec[M]) addMetric(labelValues ...string) (prometheus.Metric, error) {
	// When adding a new metric in the vector we generate a new tag.
	// This tag will be the value of the internal label labelLifeCycleTag till the expiration of the metric.
	tag := generateLifeCycleTag()
	metric, err := mv.metricVec.GetMetricWithLabelValues(append(labelValues, tag)...)
	if err != nil {
		return metric, err
	}
	mv.tags.Add(labelValues, tag)

	attr := &metricAttr{tag: tag, labelValues: labelValues}
	// Schedule the expiration time
	attr.onAccess(mv.opts.ExpirationDelay)
	mv.metricAttrs[metric] = attr
	return metric, nil
}

// must be called holding mv.mutex.Lock
func (mv *MetricVec[M]) deleteMetric(labelValues ...string) bool {
	tag, present := mv.tags.Get(labelValues)
	if !present {
		return false
	}
	metric, _ := mv.metricVec.GetMetricWithLabelValues(append(labelValues, tag)...)
	mv.tags.Delete(labelValues)
	delete(mv.metricAttrs, metric)
	mv.metricVec.DeleteLabelValues(append(labelValues, tag)...)
	return true
}

// must be called holding mv.mutex.Lock
func (mv *MetricVec[M]) deleteMetricByInstance(metric prometheus.Metric) bool {
	attr := mv.metricAttrs[metric]
	if attr == nil {
		return false
	}
	mv.metricVec.DeleteLabelValues(append(attr.labelValues, attr.tag)...)
	delete(mv.metricAttrs, metric)
	tag, present := mv.tags.Get(attr.labelValues)
	// if the deleted metric has expired it is possible that attr.tag differs from
	// the current tag (if a metric with the same label was added againg)
	if present && tag == attr.tag {
		mv.tags.Delete(attr.labelValues)
	}
	return true
}

// GetMetricWithLabelValues returns the Metric for the given slice of label
// values (same order as the variable labels in Desc). If that combination of
// label values is accessed for the first time, a new Metric is created.
//
// If an expiration delay was set in the options, the expiration time of the returned metrics is set
// to Now+ExpirationDelay.
//
// IMPORTANT: you should not keep the returned metric for later usage if an ExpirationDelay was defined, for
// the reason explained above. Since the expiration time will never be reset, the metrics would automatically
// expires after ExpirationDelay, even if its value is updated.
//
// This function mimics the function of [prometheus.MetricVec] with the same name.
func (mv *MetricVec[M]) GetMetricWithLabelValues(labelValues ...string) (M, error) {

	// First try to get an existing metric with Read lock only
	mv.mutex.RLock()
	metric, err := mv.getMetric(labelValues...)
	mv.mutex.RUnlock()

	if metric == nil && err == nil {
		// The metrics was not found, take a write lock to create a new one
		mv.mutex.Lock()
		metric, err = mv.getMetric(labelValues...) // a metric may still have been created between the two locks
		if metric == nil && err == nil {
			metric, err = mv.addMetric(labelValues...)
		}
		mv.mutex.Unlock()
	}
	return metric.(M), err
}

// WithLabelValues works as GetMetricWithLabelValues, but panics where
// GetMetricWithLabelValues would have returned an error.
func (mv *MetricVec[M]) WithLabelValues(labelValues ...string) M {
	metric, err := mv.GetMetricWithLabelValues(labelValues...)
	if err != nil {
		panic(err)
	}
	return metric
}

// GetMetricWith returns the Metric for the given Labels map (the label names
// must match those of the variable labels in Desc). If that label map is accessed for the first time,
// a new Metric is created.
//
// If an expiration delay was set in the options, the expiration time of the returned metrics is set
// to Now+ExpirationDelay.
//
// IMPORTANT: you should not keep the returned metric for later usage if an ExpirationDelay was defined, for
// the reason explained above. Since the expiration time will never be reset, the metrics would automatically
// expires after ExpirationDelay, even if its value is updated.
//
// This function mimics the function of [prometheus.MetricVec] with the same name.
func (mv *MetricVec[M]) GetMetricWith(labels prometheus.Labels) (M, error) {
	labelValues := make([]string, len(labels))
	for i, name := range mv.labelNames {
		labelValues[i] = labels[name]
	}
	return mv.GetMetricWithLabelValues(labelValues...)
}

// With works as GetMetricWith, but panics where GetMetricWithLabels would have
// returned an error.
func (mv *MetricVec[M]) With(labels prometheus.Labels) M {
	metric, err := mv.GetMetricWith(labels)
	if err != nil {
		panic(err)
	}
	return metric
}

// DeleteLabelValues removes the metrics associated to the given slice of label
// values (same order as the variable labels in Desc). It returns true if a metric was deleted.
func (mv *MetricVec[M]) DeleteLabelValues(labelValues ...string) bool {
	mv.mutex.Lock()
	defer mv.mutex.Unlock()
	return mv.deleteMetric(labelValues...)
}

// DeleteLabelValues removes the metrics associated to the given label map
// (should match the variable labels in Desc)). It returns true if a metric was deleted.
func (mv *MetricVec[M]) Delete(labels prometheus.Labels) bool {
	labelValues := make([]string, len(labels))
	for i, name := range mv.labelNames {
		labelValues[i] = labels[name]
	}
	return mv.DeleteLabelValues(labelValues...)
}

// Reset delete all the metrics of this vector.
func (mv *MetricVec[M]) Reset() {
	mv.mutex.Lock()
	defer mv.mutex.Unlock()
	mv.metricAttrs = make(map[prometheus.Metric]*metricAttr)
	mv.metricVec.Reset()
	mv.tags = newTagMap()
}

// Describe implements [prometheus.Collector].
func (mv *MetricVec[M]) Describe(ch chan<- *prometheus.Desc) {
	mv.metricVec.Describe(ch)
}

// Collect implements [prometheus.Collector].
//
// Recently added metrics are collected with their initial value till the end of their WarmUp duration.
//
// Expired metrics are ignored and removed from this vector.
func (mv *MetricVec[M]) Collect(ch chan<- prometheus.Metric) {

	expiredMetrics := make([]prometheus.Metric, 16)

	mv.mutex.RLock()
	defer mv.mutex.RUnlock()
	for metric, attr := range mv.metricAttrs {
		state := attr.onCollect(mv.opts.WarmUpDuration)
		if state == stateExpired {
			expiredMetrics = append(expiredMetrics, metric)
		} else if state == stateWarmUpOngoing {
			ch <- mv.opts.InitialMetric(metric, append(attr.labelValues, attr.tag))
		} else {
			ch <- metric
		}
	}

	// Clean-up the expired metrics asynchronously
	if len(expiredMetrics) >= 0 {
		go func() {
			mv.mutex.Lock()
			defer mv.mutex.Unlock()
			for _, metric := range expiredMetrics {
				mv.deleteMetricByInstance(metric)
			}
		}()
	}
}
