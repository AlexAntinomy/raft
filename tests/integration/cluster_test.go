package integration

import (
	"testing"
	"time"

	"raft-kv-store/pkg/kvstore"
	"raft-kv-store/pkg/raft"
)

func TestClusterFormation(t *testing.T) {
	// Создание кластера из 3 узлов
	nodes := make([]*raft.RaftNode, 3)
	stores := make([]*kvstore.Store, 3)

	for i := 0; i < 3; i++ {
		node := raft.NewRaftNode(i+1, []string{"node1", "node2", "node3"})
		store := kvstore.NewStore(node)
		nodes[i] = node
		stores[i] = store
		go node.Run()
	}

	// Ожидание выбора лидера
	time.Sleep(5 * time.Second)

	// Проверка наличия лидера
	leaderCount := 0
	for _, node := range nodes {
		if node.IsLeader() {
			leaderCount++
		}
	}
	if leaderCount != 1 {
		t.Fatalf("Expected 1 leader, got %d", leaderCount)
	}

	// Проверка репликации данных
	leaderStore := stores[0]
	if err := leaderStore.Put("key1", "value1"); err != nil {
		t.Fatalf("Failed to put value: %v", err)
	}

	time.Sleep(1 * time.Second)

	for _, store := range stores {
		value, err := store.Get("key1")
		if err != nil {
			t.Fatalf("Failed to get value: %v", err)
		}
		if value != "value1" {
			t.Fatalf("Expected value1, got %s", value)
		}
	}

	// Остановка кластера
	for _, node := range nodes {
		node.Stop()
	}
}
