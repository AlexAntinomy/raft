package monitoring

import (
	"raft-kv-store/pkg/raft"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Метрики состояния
	raftTerm = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "raft_term",
		Help: "Current Raft term",
	}, []string{"node_id"})

	raftState = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "raft_state",
		Help: "Current Raft state (0=Follower, 1=Candidate, 2=Leader)",
	}, []string{"node_id"})

	// Метрики производительности
	raftLogSize = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "raft_log_size",
		Help: "Size of the Raft log",
	}, []string{"node_id"})

	raftCommitIndex = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "raft_commit_index",
		Help: "Current commit index",
	}, []string{"node_id"})

	// Метрики сети
	raftRpcRequests = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "raft_rpc_requests_total",
		Help: "Total number of RPC requests",
	}, []string{"node_id", "type"})

	raftRpcErrors = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "raft_rpc_errors_total",
		Help: "Total number of RPC errors",
	}, []string{"node_id", "type"})
)

func RegisterMetrics(node *raft.RaftNode) {
	go func() {
		for {
			select {
			case <-node.StopCh():
				return
			case <-time.After(10 * time.Second):
				updateMetrics(node)
			}
		}
	}()
}

func updateMetrics(node *raft.RaftNode) {
	nodeID := strconv.Itoa(node.ID())

	// Состояние
	raftTerm.WithLabelValues(nodeID).Set(float64(node.CurrentTerm()))
	raftState.WithLabelValues(nodeID).Set(float64(node.State()))

	// Производительность
	raftLogSize.WithLabelValues(nodeID).Set(float64(node.LogSize()))
	raftCommitIndex.WithLabelValues(nodeID).Set(float64(node.CommitIndex()))

	// Сеть
	stats := node.RPCStats()
	for typ, count := range stats.Requests {
		raftRpcRequests.WithLabelValues(nodeID, typ).Add(float64(count))
	}
	for typ, count := range stats.Errors {
		raftRpcErrors.WithLabelValues(nodeID, typ).Add(float64(count))
	}
}
