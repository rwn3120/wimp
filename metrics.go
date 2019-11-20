package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	Total = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wimp_request_total",
		Help: "The total number of received requests",
	})
	Accepted = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wimp_request_accepted",
		Help: "The total number of accepted requests",
	})
	Refused = promauto.NewCounter(prometheus.CounterOpts{
		Name: "wimp_request_refused",
		Help: "The total number of refused requests",
	})
	Status = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "wimp_status",
		Help: "Current status of wimp",
	})
)
