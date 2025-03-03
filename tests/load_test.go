package tests

import (
	"log"
	"math/rand"
	"sync"
	"testing"
	"time"

	"raft-kv-store/pkg/kvstore"
	"raft-kv-store/pkg/raft"
)

func TestLoad(t *testing.T) {
	// Создание кластера
	node := raft.NewRaftNode(1, []string{})
	store := kvstore.NewStore(node)
	go node.Run()

	// Параметры теста
	const (
		numWorkers = 10
		numOps     = 1000
		keySize    = 10
		valueSize  = 100
	)

	var wg sync.WaitGroup
	start := time.Now()

	// Запуск воркеров
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < numOps; j++ {
				key := randString(keySize)
				value := randString(valueSize)
				if err := store.Put(key, value); err != nil {
					t.Errorf("Put failed: %v", err)
				}
				if _, err := store.Get(key); err != nil {
					t.Errorf("Get failed: %v", err)
				}
			}
		}()
	}

	wg.Wait()
	duration := time.Since(start)

	log.Printf("Load test completed in %v", duration)
	log.Printf("Throughput: %.2f ops/sec",
		float64(numWorkers*numOps)/duration.Seconds())
}

func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
