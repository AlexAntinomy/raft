package raft

import (
	"context"
	"log"
	"net"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type RaftServiceServer struct {
	raftNode *RaftNode
}

func (s *RaftServiceServer) RequestVote(ctx context.Context, req *RequestVoteRequest) (*RequestVoteResponse, error) {
	reply := &RequestVoteResponse{}
	s.raftNode.HandleRequestVote(req, reply)
	return reply, nil
}

func (s *RaftServiceServer) AppendEntries(ctx context.Context, req *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	reply := &AppendEntriesResponse{}
	s.raftNode.HandleAppendEntries(req, reply)
	return reply, nil
}

func (rn *RaftNode) startRPCServer() {
	lis, err := net.Listen("tcp", rn.rpcAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	RegisterRaftServiceServer(grpcServer, &RaftServiceServer{raftNode: rn})

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()
}

type RaftClient struct {
	conn   *grpc.ClientConn
	client RaftServiceClient
}

var clientCache = struct {
	sync.RWMutex
	clients map[string]*RaftClient
}{clients: make(map[string]*RaftClient)}

func getClient(addr string) (*RaftClient, error) {
	clientCache.RLock()
	if client, ok := clientCache.clients[addr]; ok {
		clientCache.RUnlock()
		return client, nil
	}
	clientCache.RUnlock()

	clientCache.Lock()
	defer clientCache.Unlock()

	// Double check
	if client, ok := clientCache.clients[addr]; ok {
		return client, nil
	}

	conn, err := grpc.Dial(addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithTimeout(2*time.Second))
	if err != nil {
		return nil, err
	}

	client := &RaftClient{
		conn:   conn,
		client: NewRaftServiceClient(conn),
	}

	clientCache.clients[addr] = client
	return client, nil
}

func (rn *RaftNode) sendRequestVote(peer string, args *RequestVoteRequest) (*RequestVoteResponse, error) {
	client, err := getClient(peer)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return client.client.RequestVote(ctx, args)
}

func (rn *RaftNode) sendAppendEntries(peer string, args *AppendEntriesRequest) (*AppendEntriesResponse, error) {
	client, err := getClient(peer)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	return client.client.AppendEntries(ctx, args)
}

func (rn *RaftNode) stopRPCServer() {
	clientCache.Lock()
	defer clientCache.Unlock()

	for addr, client := range clientCache.clients {
		client.conn.Close()
		delete(clientCache.clients, addr)
	}
}
