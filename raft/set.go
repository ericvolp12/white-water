package raft

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func (s *Sailor) handle_set(msg messages.Message, state *storage.State) {
	trans := storage.GenerateTransaction(storage.SetOp, msg.Key, msg.Value)
	newEntry := entry{term: s.currentTerm, trans: trans, votes: 1}
	s.log = append(s.log, newEntry)
}

/*
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
	/*
		i := s.incrementCommit(MatchIndex)
		if i == -1 {
			fmt.Printf("ERROR: Commit reply had non-existant MatchIndex\n")
			return i
		}
		majority := uint((len(s.client.Peers) + 1) / 2) //TODO fix simple majority
		if s.leader.queue[i].commit_count >= majority {
			s.leader.queue = delete(s.leader.queue, uint(i))
			// TODO Send to broker
		}
	return 1
}
*/

func (s *Sailor) handle_prepare(PrepLower uint, PrepUpper uint) error {

	for i := PrepLower; i <= PrepUpper; i++ {
		if i <= s.volatile.commitIndex {
			continue
		}
		s.log[i-1].votes += 1
		if s.log[i-1].votes > uint((len(s.client.Peers)+1)/2) && s.log[i-1].term == s.currentTerm {
			for j := s.volatile.commitIndex + 1; j <= i; j++ {
				s.log[j-1].votes = 1
			}
			s.volatile.commitIndex = i
		}
	}
	return nil
}
