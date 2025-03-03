package raft

import (
	"sync"
)

type RaftNode struct {
	mu          sync.Mutex
	id          int
	state       State
	currentTerm int
	votedFor    int
	log         []LogEntry
	commitIndex int
	lastApplied int
	peers       []string

	nextIndex  map[int]int
	matchIndex map[int]int

	applyCh chan ApplyMsg
	rpcCh   chan RPC
	stopCh  chan struct{}
}

type LogEntry struct {
	Term    int
	Command interface{}
}

func (rn *RaftNode) Run() {
	for {
		select {
		case <-rn.stopCh:
			return
		default:
			switch rn.getState() {
			case Follower:
				rn.runFollower()
			case Candidate:
				rn.runCandidate()
			case Leader:
				rn.runLeader()
			}
		}
	}
}

func (rn *RaftNode) runFollower() {
	timeout := randomTimeout()
	select {
	case rpc := <-rn.rpcCh:
		rn.handleRPC(rpc)
	case <-timeout:
		rn.setState(Candidate)
	}
}

func (rn *RaftNode) startElection() {
	rn.mu.Lock()
	rn.currentTerm++
	rn.votedFor = rn.id
	votes := 1

	args := RequestVoteArgs{
		Term:        rn.currentTerm,
		CandidateId: rn.id,
	}

	for _, peer := range rn.peers {
		go func(peer string) {
			var reply RequestVoteReply
			if err := rpcCall(peer, "Raft.RequestVote", &args, &reply); err == nil {
				rn.mu.Lock()
				if reply.Term > rn.currentTerm {
					rn.currentTerm = reply.Term
					rn.setState(Follower)
				} else if reply.VoteGranted {
					votes++
					if votes > len(rn.peers)/2 {
						rn.setState(Leader)
					}
				}
				rn.mu.Unlock()
			}
		}(peer)
	}
	rn.mu.Unlock()
}
