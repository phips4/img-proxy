package prom

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CachedImages = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imgproxy_cached_images_total",
		Help: "The total number of cached images on this node",
	})
	CachedImageBytes = promauto.NewCounter(prometheus.CounterOpts{
		Name: "imgproxy_cached_images_bytes",
		Help: "The total size of all images stored on this node",
	})
)
