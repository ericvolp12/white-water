package raft

import (
	"fmt"
	messages "github.com/ericvolp12/white-water/messages"
)

func (s *Sailor) handle_set(msg messages.Message) {

	s.leader.queue = append(s.leader.queue, commit_queue{index: 0, commit_count: 0}) //TODO GET INDEX
}

// Delete an item in the leader's queue of pending commits
func delete(q []commit_queue, i uint) []commit_queue {
	q = append(q[:i], q[i+1:]...)
	return q
}

// Find the queue entry for MatchIndex and increments the number of commits
func (s *Sailor) incrementCommit(MatchIndex uint) int {
	for i, _ := range s.leader.queue {
		if s.leader.queue[i].index == MatchIndex {
			s.leader.queue[i].commit_count += 1
			return i
		}
	}
	return -1
}

// Upon appendReply w/ Sucess, increments the appropriate pending commit count
// and sends a message to the broker if a majority have now commited
func (s *Sailor) handle_commit(MatchIndex uint) int {
	i := s.incrementCommit(MatchIndex)
	if i == -1 {
		fmt.Printf("ERROR: Commit reply had non-existant MatchIndex\n")
		return i
	}
	majority := uint((len(s.client.Peers) + 1) / 2)
	if s.leader.queue[i].commit_count > majority {
		s.leader.queue = delete(s.leader.queue, uint(i))
		// TODO Send to broker
	}
	return i
}
