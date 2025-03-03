package raft

import (
	"context"
	"math/rand"
	"time"
)

const (
	minElectionTimeout = 1500 * time.Millisecond
	maxElectionTimeout = 3000 * time.Millisecond
)

func (rn *RaftNode) runElectionTimer() {
	timeout := randomTimeout()
	for {
		select {
		case <-rn.stopCh:
			return
		case <-time.After(timeout):
			rn.mu.Lock()
			if rn.state == Leader {
				rn.mu.Unlock()
				return
			}

			rn.startNewElection()
			rn.mu.Unlock()
			return

		case <-rn.resetTimerCh:
			timeout = randomTimeout()
		}
	}
}

func (rn *RaftNode) startNewElection() {
	rn.currentTerm++
	rn.state = Candidate
	rn.votedFor = rn.id
	rn.votesReceived = 1

	lastLogIndex, lastLogTerm := rn.getLastLogInfo()

	args := RequestVoteArgs{
		Term:         rn.currentTerm,
		CandidateID:  rn.id,
		LastLogIndex: lastLogIndex,
		LastLogTerm:  lastLogTerm,
	}

	voteChan := make(chan bool, len(rn.peers))

	for _, peer := range rn.peers {
		go func(peer string) {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()

			var reply RequestVoteReply
			err := rn.sendRequestVoteRPC(ctx, peer, &args, &reply)
			if err != nil {
				voteChan <- false
				return
			}

			rn.mu.Lock()
			defer rn.mu.Unlock()

			if reply.Term > rn.currentTerm {
				rn.stepDown(reply.Term)
				voteChan <- false
				return
			}

			if reply.VoteGranted {
				voteChan <- true
			} else {
				voteChan <- false
			}
		}(peer)
	}

	go rn.countVotes(voteChan, len(rn.peers))
}

func (rn *RaftNode) countVotes(voteChan <-chan bool, peersCount int) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	quorum := (peersCount+1)/2 + 1
	votes := 1

	for i := 0; i < peersCount; i++ {
		if <-voteChan {
			votes++
			if votes >= quorum && rn.state == Candidate {
				rn.becomeLeader()
				return
			}
		}
	}
}

func (rn *RaftNode) becomeLeader() {
	rn.state = Leader
	rn.nextIndex = make(map[int]int)
	rn.matchIndex = make(map[int]int)

	for _, peer := range rn.peers {
		rn.nextIndex[peer.id] = rn.log.LastIndex() + 1
		rn.matchIndex[peer.id] = 0
	}

	go rn.sendHeartbeats()
}

func (rn *RaftNode) stepDown(term int) {
	rn.state = Follower
	rn.currentTerm = term
	rn.votedFor = -1
}

func (rn *RaftNode) sendHeartbeats() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rn.mu.Lock()
			if rn.state != Leader {
				rn.mu.Unlock()
				return
			}
			rn.mu.Unlock()

			rn.broadcastAppendEntries()
		case <-rn.stopCh:
			return
		}
	}
}

func randomTimeout() time.Duration {
	return minElectionTimeout +
		time.Duration(rand.Int63n(int64(maxElectionTimeout-minElectionTimeout)))
}
