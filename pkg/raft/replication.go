package raft

import "sort"

type AppendEntriesArgs struct {
	Term         int
	LeaderID     int
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term          int
	Success       bool
	ConflictIndex int
	ConflictTerm  int
}

func (rn *RaftNode) broadcastAppendEntries() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	if rn.state != Leader {
		return
	}

	for peerID := range rn.peers {
		if peerID == rn.id {
			continue
		}

		go rn.replicateToPeer(peerID)
	}
}

func (rn *RaftNode) replicateToPeer(peerID int) {
	rn.mu.Lock()

	nextIndex := rn.nextIndex[peerID]
	prevLogIndex := nextIndex - 1
	prevLogTerm := 0

	if prevLogIndex >= 0 {
		prevLogTerm = rn.log[prevLogIndex].Term
	}

	entries := make([]LogEntry, 0)
	if nextIndex <= len(rn.log)-1 {
		entries = rn.log[nextIndex:]
	}

	args := AppendEntriesArgs{
		Term:         rn.currentTerm,
		LeaderID:     rn.id,
		PrevLogIndex: prevLogIndex,
		PrevLogTerm:  prevLogTerm,
		Entries:      entries,
		LeaderCommit: rn.commitIndex,
	}

	rn.mu.Unlock()

	var reply AppendEntriesReply
	err := rn.sendAppendEntries(peerID, &args, &reply)
	if err != nil {
		return
	}

	rn.mu.Lock()
	defer rn.mu.Unlock()

	if reply.Term > rn.currentTerm {
		rn.stepDown(reply.Term)
		return
	}

	if !reply.Success {
		if reply.ConflictTerm != 0 {
			lastIndex := rn.findLastIndexForTerm(reply.ConflictTerm)
			if lastIndex != -1 {
				rn.nextIndex[peerID] = lastIndex + 1
			} else {
				rn.nextIndex[peerID] = reply.ConflictIndex
			}
		} else {
			rn.nextIndex[peerID] = reply.ConflictIndex
		}
		return
	}

	rn.nextIndex[peerID] = nextIndex + len(entries)
	rn.matchIndex[peerID] = rn.nextIndex[peerID] - 1
	rn.updateCommitIndex()
}

func (rn *RaftNode) updateCommitIndex() {
	matchIndexes := make([]int, 0, len(rn.peers))
	for _, mi := range rn.matchIndex {
		matchIndexes = append(matchIndexes, mi)
	}
	matchIndexes = append(matchIndexes, len(rn.log)-1)

	sort.Ints(matchIndexes)
	newCommitIndex := matchIndexes[len(matchIndexes)/2]

	if newCommitIndex > rn.commitIndex &&
		rn.log[newCommitIndex].Term == rn.currentTerm {
		rn.commitIndex = newCommitIndex
		go rn.applyLogs()
	}
}

func (rn *RaftNode) applyLogs() {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	for rn.lastApplied < rn.commitIndex {
		rn.lastApplied++
		entry := rn.log[rn.lastApplied]
		rn.applyCh <- ApplyMsg{
			CommandValid: true,
			Command:      entry.Command,
			CommandIndex: rn.lastApplied,
		}
	}
}

func (rn *RaftNode) HandleAppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	rn.mu.Lock()
	defer rn.mu.Unlock()

	reply.Term = rn.currentTerm
	reply.Success = false

	if args.Term < rn.currentTerm {
		return
	}

	rn.resetElectionTimer()

	if args.Term > rn.currentTerm {
		rn.stepDown(args.Term)
	}

	if args.PrevLogIndex >= len(rn.log) ||
		(args.PrevLogIndex >= 0 && rn.log[args.PrevLogIndex].Term != args.PrevLogTerm) {

		reply.ConflictIndex = len(rn.log)
		if args.PrevLogIndex < len(rn.log) {
			reply.ConflictTerm = rn.log[args.PrevLogIndex].Term
			for i := args.PrevLogIndex - 1; i >= 0; i-- {
				if rn.log[i].Term != reply.ConflictTerm {
					reply.ConflictIndex = i + 1
					break
				}
			}
		}
		return
	}

	rn.log = append(rn.log[:args.PrevLogIndex+1], args.Entries...)

	if args.LeaderCommit > rn.commitIndex {
		rn.commitIndex = min(args.LeaderCommit, len(rn.log)-1)
		go rn.applyLogs()
	}

	reply.Success = true
}
