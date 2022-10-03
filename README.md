
# Smart Prometheus Client

This is a Go client library for Prometheus to instrument application code. 
It extends the [official Prometheus client library](https://github.com/prometheus/client_golang) by adding two main features that ease the implemention of Prometheus exporters:

- [Counters warm-up](#counters-warm-up)
- [Smart clean-up of metric vector](#clean-up-of-metrics-vectors)

You will find more detailed description of these two main features [below](#features).

## How to use it

This library reuses the different metrics interfaces from the [official Prometheus client library](https://github.com/prometheus/client_golang) (Counter, Histogram, Summary.. ), and proposes a compatible vector API. This should allow to use both libraries interchangeably in the most common cases.

### Installation

```
go get github.com/goto-opensource/smart-prometheus-client
```

### Basic Example: 

```golang
import ( 
  "time"
  "math/rand"

  "github.com/prometheus/client_golang/prometheus"
  "github.com/goto-opensource/smart-prometheus-client/metrics"
)

var myCounter prometheus.Counter
var myHistogram prometheus.Histogram

func init() {
  countOpts := metrics.HistogramOpts{
		HistogramOpts: prometheus.HistogramOpts{
			Namespace: "namespace",
			Subsystem: "myapp",
			Name:      "hist",
			Help:      "Help message",
			Buckets:   []float64{1.0, 2.0, 10.0},
		},
		WarmUpDuration: 10 * time.Second,
	}
  myCounter = metrics.NewHistogram(countOpts)

  histOpts := metrics.HistogramOpts{
		HistogramOpts: prometheus.HistogramOpts{
			Namespace: "namespace",
			Subsystem: "myapp",
			Name:      "hist",
			Help:      "Help message",
			Buckets:   []float64{1.0, 2.0, 10.0},
		},
		WarmUpDuration: 10 * time.Second,
	}
  myHistogram = metrics.NewHistogram(histOpts)
}


func myAppProcess() {
  myCounter.Inc()
  myHistogram.Observe(rand.Float64() * 20)
}

```

## Features

### Counters Warm-up

Warm-up mechanism automates the initialization of counter values to 0 without the need of pre-populating them in advance.
Sometimes it is indeed not possible to predict all the possible set of labels values.

This solves the problem of [missing metrics](https://prometheus.io/docs/practices/instrumentation/#avoid-missing-metrics) that may cause unexpected result at query time when using increase() function (reported [here](https://github.com/prometheus/prometheus/issues/1673)).

At first collection the counters will first returns 0 instead of their actual value (which is still kept in the background).
In case several Prometheus instances scrapes your exporter, you can also define a Warm-up duration (usually set to the scrape period) during which the counters will be collected with 0 value.

Two drawbacks of this solution are:
- The initial export of the counters is delayed.
- The exporter configuration becomes dependant of its consumers scrape period.


### Clean-up of Metrics Vectors

Sometimes it is useful to remove some set of label values from metrics vectors, for example when they are not updated anymore. This saves memory in the exporter but also in Prometheus since this avoid keeping exporting constant time series for ever.

However, deleting time series may cause inconsistency issues when the same set of labels is added again later-on.

This library solves this problem of label collision for deleted metrics in vector. Internally it adds and manages an additional label whose value changes when a new life cycle starts.

It is also possible to automatically removes idle metrics from Vector thanks to the `ExpirationDelay` option provided at vector creation. Still the removed set of label values can be safely added again due to the mechanism described earlier. Note that the WarmUp process triggers again in such case, which makes it safe for counters, histograms and summary.

## Contributing

Please see the [contribution guidelines](CONTRIBUTING.md).
