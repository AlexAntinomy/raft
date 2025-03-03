package main

import (
	"context"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"

	"raft-kv-store/pkg/api"
	"raft-kv-store/pkg/kvstore"
	"raft-kv-store/pkg/raft"
)

var (
	nodeID   = flag.Int("id", 1, "Unique node ID")
	httpPort = flag.String("http", "8080", "HTTP server port")
	grpcPort = flag.String("grpc", "9090", "gRPC server port")
	peers    = flag.String("peers", "", "Comma-separated list of peer addresses")
)

func main() {
	flag.Parse()

	raftNode := raft.NewRaftNode(*nodeID, parsePeers(*peers))

	store := kvstore.NewStore(raftNode)

	go raftNode.Run()

	router := mux.NewRouter()
	apiServer := api.NewHTTPServer(store, raftNode)

	router.HandleFunc("/key/{key}", apiServer.HandleGetKey).Methods("GET")
	router.HandleFunc("/key/{key}", apiServer.HandlePutKey).Methods("PUT")
	router.HandleFunc("/cluster/status", apiServer.HandleClusterStatus).Methods("GET")

	httpServer := &http.Server{
		Addr:         ":" + *httpPort,
		Handler:      router,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	grpcServer := grpc.NewServer()
	raft.RegisterRaftServer(grpcServer, api.NewRaftService(raftNode))

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		log.Printf("Starting HTTP server on port %s", *httpPort)
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	go func() {
		defer wg.Done()
		lis, err := net.Listen("tcp", ":"+*grpcPort)
		if err != nil {
			log.Fatalf("Failed to listen: %v", err)
		}
		log.Printf("Starting gRPC server on port %s", *grpcPort)
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("gRPC server failed: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	<-stop
	log.Println("Shutting down servers...")

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	grpcServer.GracefulStop()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Printf("HTTP shutdown error: %v", err)
	}

	raftNode.Stop()
	wg.Wait()
	log.Println("Server stopped gracefully")
}

func parsePeers(peersStr string) []string {
	if peersStr == "" {
		return []string{}
	}
	return strings.Split(peersStr, ",")
}
