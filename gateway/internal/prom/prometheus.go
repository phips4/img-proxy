package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ImageHandlerHits = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imgproxy_image_handler_hits_total",
		Help: "The total number of processed requests from the image handler",
	})
	ImageHandlerErrors = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imgproxy_image_handler_errors_total",
		Help: "The total number of errors which occurred in the image handler",
	})
)
