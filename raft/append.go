package raft

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func handleAppendEntries(s *Sailor, state *storage.State, am *appendMessage) (appendReply, error) {
	//TODO(JM): Handle check if new leader

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
	rep.MatchIndex = uint(len(s.log) - 1)
	return rep, nil
}

func sendAppendEntries(s *Sailor, state *storage.State, peer string) error {
	am := appendMessage{}
	am.Term = s.currentTerm
	am.LeaderId = s.client.NodeName
	am.PrevLogIndex = s.leader.nextIndex[peer] - 1
	am.PrevLogTerm = s.log[s.leader.nextIndex[peer]-1].term
	am.Entries = s.log[s.leader.nextIndex[peer]:]
	am.LeaderCommit = s.volatile.commitIndex
	ap := messages.Message{}
	ap.Type = "appendEntries"
	ap.ID = 0 //TODO(JM): Figure out what this should be?
	ap.Source = s.client.NodeName
	ap.Value = makePayload(am)
	s.client.SendToPeer(ap, peer)
	return nil
}

func handleAppendReply(s *Sailor, state *storage.State, ar *appendReply, source string) error {
	//TODO(JM): Check to update our term
	if ar.Success {
		s.leader.nextIndex[source] = ar.MatchIndex + 1
		s.leader.matchIndex[source] = ar.MatchIndex
	} else {
		s.leader.nextIndex[source] -= 1
		return sendAppendEntries(s, state, source)
	}
	return nil
}
