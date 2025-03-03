package api

import (
	"context"
	"log"
	"net"

	"raft-kv-store/pkg/raft"
	pb "raft-kv-store/proto"

	"google.golang.org/grpc"
)

type RaftServiceServer struct {
	pb.UnimplementedRaftServiceServer
	raftNode *raft.RaftNode
}

func NewRaftService(raftNode *raft.RaftNode) *RaftServiceServer {
	return &RaftServiceServer{
		raftNode: raftNode,
	}
}

func (s *RaftServiceServer) RequestVote(ctx context.Context, req *pb.RequestVoteRequest) (*pb.RequestVoteResponse, error) {
	var reply pb.RequestVoteResponse
	s.raftNode.HandleRequestVote(req, &reply)
	return &reply, nil
}

func (s *RaftServiceServer) AppendEntries(ctx context.Context, req *pb.AppendEntriesRequest) (*pb.AppendEntriesResponse, error) {
	var reply pb.AppendEntriesResponse
	s.raftNode.HandleAppendEntries(req, &reply)
	return &reply, nil
}

func StartGRPCServer(raftNode *raft.RaftNode, addr string) {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterRaftServiceServer(grpcServer, NewRaftService(raftNode))

	log.Printf("gRPC server listening at %v", lis.Addr())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
