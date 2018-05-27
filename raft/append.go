package raft

import storage "github.com/ericvolp12/white-water/storage"

func handleAppendEntries(s *Sailor, state *storage.State, am *appendMessage) (appendReply, error) {
	rep := appendReply{Term: s.currentTerm, Success: false}
	if s.currentTerm > am.Term {
		return rep, nil
	}

	if len(s.log) <= int(am.PrevLogIndex) || s.log[am.PrevLogIndex].term != am.PrevLogTerm {
		return rep, nil
	}

	rep.Success = true

	s.log = append(s.log[:am.PrevLogIndex+1], am.Entries...)
	if am.LeaderCommit > s.volatile.commitIndex {
		if int(am.LeaderCommit) <= len(s.log)-1 {
			s.volatile.commitIndex = am.LeaderCommit
		} else {
			s.volatile.commitIndex = uint(len(s.log) - 1)
		}
		for s.volatile.lastApplied < s.volatile.commitIndex {
			s.volatile.lastApplied += 1
			state.ApplyTransaction(s.log[s.volatile.lastApplied].trans)
		}

	}

	return rep, nil
}
