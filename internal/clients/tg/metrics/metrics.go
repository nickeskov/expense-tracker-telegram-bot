package metrics

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const (
	endpointLabelName    = "endpoint"
	errorStatusLabelName = "error_status"
)

var (
	inFlightRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: "tg",
		Subsystem: "bot",
		Name:      "in_flight_requests_total",
	}, []string{endpointLabelName, errorStatusLabelName})
	inFlightRequestsDuration = promauto.NewSummaryVec(prometheus.SummaryOpts{
		Namespace: "tg",
		Subsystem: "bot",
		Name:      "summary_response_time_seconds",
		Objectives: map[float64]float64{
			0.5:  0.1,
			0.9:  0.01,
			0.99: 0.001,
		},
	}, []string{endpointLabelName, errorStatusLabelName})
)

func IncInFlightRequests(stringEndpoint string, errStatus bool) {
	inFlightRequests.WithLabelValues(stringEndpoint, strconv.FormatBool(errStatus)).Inc()
}

func ObserveInFlightRequestsDuration(stringEndpoint string, errStatus bool, duration time.Duration) {
	inFlightRequestsDuration.WithLabelValues(stringEndpoint, strconv.FormatBool(errStatus)).Observe(duration.Seconds())
}
