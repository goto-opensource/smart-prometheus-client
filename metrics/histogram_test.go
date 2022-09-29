package metrics

import (
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestHistogram_ReturnsZeroDuringWarmup(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := HistogramOpts{
		HistogramOpts: prometheus.HistogramOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "hist",
			Help:      "Help message",
			Buckets:   []float64{1.0, 2.0, 10.0},
		},
		WarmUpDuration: 10 * time.Second,
	}
	hist := NewHistogram(opts)
	hist.Observe(1.4)
	hist.Observe(0.5)

	expect := `
		# HELP namespace_something_hist Help message
		# TYPE namespace_something_hist histogram
		namespace_something_hist_bucket{le="1"} 0
		namespace_something_hist_bucket{le="2"} 0
		namespace_something_hist_bucket{le="10"} 0
		namespace_something_hist_bucket{le="+Inf"} 0
		namespace_something_hist_sum 0
		namespace_something_hist_count 0
		`
	err := testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)

	// second collect
	SetUpNowTime(t0.Add(11 * time.Second))
	expect = `
		# HELP namespace_something_hist Help message
		# TYPE namespace_something_hist histogram
		namespace_something_hist_bucket{le="1"} 1
		namespace_something_hist_bucket{le="2"} 2
		namespace_something_hist_bucket{le="10"} 2
		namespace_something_hist_bucket{le="+Inf"} 2
		namespace_something_hist_sum 1.9
		namespace_something_hist_count 2
		`
	err = testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)
}

func TestHistogramVec_ReturnsZeroDuringWarmup(t *testing.T) {
	t0 := defaultTime
	SetUpNowTime(t0)

	opts := HistogramOpts{
		HistogramOpts: prometheus.HistogramOpts{
			Namespace: "namespace",
			Subsystem: "something",
			Name:      "hist",
			Help:      "Help message",
			Buckets:   []float64{1.0, 2.0, 10.0},
		},
		WarmUpDuration:  10 * time.Second,
		ExpirationDelay: 60 * time.Minute,
	}
	hist := NewHistogramVec(opts, []string{"instance", "user"})

	// Add a new histogram and check warm-up state
	hist.With(prometheus.Labels{"instance": "host1", "user": "alex"}).Observe(1.4)
	hist.WithLabelValues("host1", "alex").Observe(3.0)

	expect := `
		# HELP namespace_something_hist Help message
		# TYPE namespace_something_hist histogram
		namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="1"} 0
		namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="2"} 0
		namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="10"} 0
		namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="+Inf"} 0
		namespace_something_hist_sum{_tag_="48ab9774",instance="host1",user="alex"} 0
		namespace_something_hist_count{_tag_="48ab9774",instance="host1",user="alex"} 0
		`
	err := testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)

	// Add another histogram and advance the time to the end of warm-up state of the first one
	hist.With(prometheus.Labels{"instance": "host1", "user": "alex"}).Observe(2)
	hist.WithLabelValues("host1", "alex").Observe(3.1)
	hist.With(prometheus.Labels{"instance": "host2", "user": "toto"}).Observe(15)
	hist.WithLabelValues("host2", "toto").Observe(0.0)

	SetUpNowTime(t0.Add(11 * time.Second))
	expect = `
	# HELP namespace_something_hist Help message
	# TYPE namespace_something_hist histogram
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="1"} 0
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="2"} 2
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="10"} 4
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="+Inf"} 4
	namespace_something_hist_sum{_tag_="48ab9774",instance="host1",user="alex"} 9.5
	namespace_something_hist_count{_tag_="48ab9774",instance="host1",user="alex"} 4
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="1"} 0
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="2"} 0
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="10"} 0
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="+Inf"} 0
	namespace_something_hist_sum{_tag_="48ab9774",instance="host2",user="toto"} 0
	namespace_something_hist_count{_tag_="48ab9774",instance="host2",user="toto"} 0
	`
	err = testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)

	// Advance the time to make sure both histogram has completed their warm-up state
	SetUpNowTime(t0.Add(22 * time.Second))
	expect = `
	# HELP namespace_something_hist Help message
	# TYPE namespace_something_hist histogram
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="1"} 0
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="2"} 2
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="10"} 4
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host1",user="alex",le="+Inf"} 4
	namespace_something_hist_sum{_tag_="48ab9774",instance="host1",user="alex"} 9.5
	namespace_something_hist_count{_tag_="48ab9774",instance="host1",user="alex"} 4
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="1"} 1
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="2"} 1
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="10"} 1
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="+Inf"} 2
	namespace_something_hist_sum{_tag_="48ab9774",instance="host2",user="toto"} 15
	namespace_something_hist_count{_tag_="48ab9774",instance="host2",user="toto"} 2
	`
	err = testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)

	// Touch one of the histogram at t0 + 30min, then advance the time to t0 + ExpirationDelay,
	// Check that only one histogram remains
	SetUpNowTime(t0.Add(30 * time.Minute))
	hist.WithLabelValues("host2", "toto").Observe(9.2)
	SetUpNowTime(t0.Add(61 * time.Minute))
	expect = `
	# HELP namespace_something_hist Help message
	# TYPE namespace_something_hist histogram
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="1"} 1
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="2"} 1
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="10"} 2
	namespace_something_hist_bucket{_tag_="48ab9774",instance="host2",user="toto",le="+Inf"} 3
	namespace_something_hist_sum{_tag_="48ab9774",instance="host2",user="toto"} 24.2
	namespace_something_hist_count{_tag_="48ab9774",instance="host2",user="toto"} 3
	`
	err = testutil.CollectAndCompare(hist, strings.NewReader(expect), "namespace_something_hist")
	assert.NoError(t, err)
}
