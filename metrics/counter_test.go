package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

var defaultTime = time.Unix(1219204980, 0).UTC()

func TestCounter_FirstCollectReturnsZero(t *testing.T) {
	RestoreNowTime()

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
	}
	counter := NewCounter(opts)
	counter.Add(10)

	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// second collect should return the actual counter value
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count 10
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)
}

func TestCounter_ReturnsZeroDuringWarmup(t *testing.T) {

	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
		WarmUpDuration: 10 * time.Second,
	}
	counter := NewCounter(opts)
	counter.Add(10)

	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// second collect should still return the actual counter value
	counter.Inc()
	SetUpNowTime(t0.Add(5 * time.Second))

	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count 0
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// after warmup duration
	SetUpNowTime(t0.Add(11 * time.Second))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count 11
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)
}

func TestCounterVec_FirstCollectReturnsZero(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
	}
	counter := NewCounterVec(opts, []string{"name", "instance"})

	counter.WithLabelValues("toto", "10.0.0.1").Add(10)
	counter.With(prometheus.Labels{"name": "titi", "instance": "10.0.0.2"}).Inc()

	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 0
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="titi"} 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// second collect should return the actual counter values
	SetUpNowTime(t0.Add(1))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 10
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="titi"} 1
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)
}

func TestCounterVec_ReturnsZeroDuringWarmup(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
		WarmUpDuration: 10 * time.Second,
	}
	counter := NewCounterVec(opts, []string{"name", "instance"})

	// First Collect cycle, only one counter registered
	counter.WithLabelValues("toto", "10.0.0.1").Add(10)
	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Second Collect cycle, first counter in warmUp and one new counter registered
	counter.With(prometheus.Labels{"name": "titi", "instance": "10.0.0.2"}).Inc()
	SetUpNowTime(t0.Add(5 * time.Second))

	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 0
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="titi"} 0
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Third cycle first counter warmUp is over, second one still warmingUp
	SetUpNowTime(t0.Add(11 * time.Second))
	counter.WithLabelValues("toto", "10.0.0.1").Inc()
	counter.WithLabelValues("titi", "10.0.0.2").Inc()
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 11
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="titi"} 0
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Last cycle both timer warmUp is over
	SetUpNowTime(t0.Add(16 * time.Second))
	counter.WithLabelValues("toto", "10.0.0.1").Inc()
	counter.WithLabelValues("titi", "10.0.0.2").Inc()
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 12
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="titi"} 3
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)
}

func TestCounterVec_Delete(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
	}
	counter := NewCounterVec(opts, []string{"name", "instance"})

	counter.WithLabelValues("toto", "10.0.0.1").Add(10)
	counter.WithLabelValues("toto", "10.0.0.2").Add(12)
	counter.WithLabelValues("titi", "10.0.0.3").Add(15)

	ok := counter.DeleteLabelValues("toto", "10.0.0.2")
	assert.True(t, ok)
	ok = counter.DeleteLabelValues("toto", "10.0.0.2")
	assert.False(t, ok)

	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 0
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.3",name="titi"} 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	ok = counter.Delete(prometheus.Labels{"name": "titi", "instance": "10.0.0.3"})
	assert.True(t, ok)
	ok = counter.Delete(prometheus.Labels{"name": "titi", "instance": "10.0.0.3"})
	assert.False(t, ok)

	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 10
		`
	SetUpNowTime(t0.Add(1))
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Test the internal structure and check we successfully unregistered the metrics
	assert.Len(t, counter.metricAttrs, 1)
}

func TestCounterVec_Reset(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
	}
	counter := NewCounterVec(opts, []string{"name", "instance"})

	counter.WithLabelValues("toto", "10.0.0.1").Add(10)
	counter.WithLabelValues("toto", "10.0.0.2").Add(12)
	counter.WithLabelValues("titi", "10.0.0.3").Add(15)

	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.1",name="toto"} 0
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.2",name="toto"} 0
		namespace_something_count{_tag_="48ab9774",instance="10.0.0.3",name="titi"} 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	counter.Reset()
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Test the internal structure and check we successfully unregistered the metrics
	assert.Len(t, counter.metricAttrs, 0)
	assert.Len(t, counter.tags.index, 0)
}

func TestCounterVec_MetricsExpiration(t *testing.T) {

	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
		ExpirationDelay: 10 * time.Second,
	}
	counter := NewCounterVec(opts, []string{"label"})
	counter.WithLabelValues("toto").Add(10)
	counter.WithLabelValues("titi").Add(15)

	// Make sure the metrics warmedUp
	testutil.CollectAndCount(counter)
	SetUpNowTime(t0.Add(1)) // clock tick

	// Warm-up complete
	expect := `
	# HELP namespace_something_count Help message
	# TYPE namespace_something_count counter
	namespace_something_count{_tag_="48ab9774",label="toto"} 10
	namespace_something_count{_tag_="48ab9774",label="titi"} 15
	`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// update only one of the metric after 5s
	SetUpNowTime(t0.Add(5 * time.Second))
	counter.WithLabelValues("toto").Inc()
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",label="toto"} 11
		namespace_something_count{_tag_="48ab9774",label="titi"} 15
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// Only the other metric should expire
	SetUpNowTime(t0.Add(11 * time.Second))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",label="toto"} 11
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

}

func TestCounterVec_SingleMetricFullLifeCycle(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := CounterOpts{
		CounterOpts: prometheus.CounterOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "count",
			Help:      "Help message",
		},
		WarmUpDuration:  2 * time.Second,
		ExpirationDelay: 10 * time.Second,
	}
	counter := NewCounterVec(opts, []string{"label"})
	counter.WithLabelValues("toto").Add(10)

	// warm-up
	expect := `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",label="toto"} 0
		`
	err := testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// warmup-complete
	SetUpNowTime(t0.Add(3 * time.Second))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9774",label="toto"} 10
	`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// expiration
	SetUpNowTime(t0.Add(15 * time.Second))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
	`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// add again a metric with the same counter
	// the tag value should change, and the new metric should be in warm-up state
	counter.WithLabelValues("toto").Add(15)
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9783",label="toto"} 0
		`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)

	// end of warm-up state
	SetUpNowTime(t0.Add(20 * time.Second))
	expect = `
		# HELP namespace_something_count Help message
		# TYPE namespace_something_count counter
		namespace_something_count{_tag_="48ab9783",label="toto"} 15
	`
	err = testutil.CollectAndCompare(counter, strings.NewReader(expect), "namespace_something_count")
	assert.NoError(t, err)
}
