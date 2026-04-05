package engine

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	RateLimitRequests = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "gopher_guard_requests_total",
			Help: "Total number of rate limit requests processed by Gopher-Guard",
		},
		[]string{"status"},
	)

	TotalGRPCRequest = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total no of grpc requests achieved",
		},
		[]string{"method"},
	)

	RaftCurrentTerm = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "raft_current_term",
			Help: "The current raft term(increments on elections)",
		},
	)

	RaftIsLeader = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "raft_is_leader",
			Help: "Returns 1 if the node is the leader otherwise 0",
		},
	)
)
