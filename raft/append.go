package raft

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func handleAppendEntries(s *Sailor, state *storage.State, am *appendMessage) (appendReply, error) {
	if (s.state != follower && am.Term == s.currentTerm) || am.Term > s.currentTerm {
		s.becomeFollower(am.Term)
	}
	//Converted to 1 indexed
	rep := appendReply{Term: s.currentTerm, Success: false}
	if s.currentTerm > am.Term {
		return rep, nil
	}

	if len(s.log) <= int(am.PrevLogIndex-1) || s.log[am.PrevLogIndex-1].term != am.PrevLogTerm {
		return rep, nil
	}

	rep.Success = true

	s.log = append(s.log[:am.PrevLogIndex], am.Entries...)
	if am.LeaderCommit > s.volatile.commitIndex {
		if int(am.LeaderCommit) <= len(s.log) {
			s.volatile.commitIndex = am.LeaderCommit
		} else {
			s.volatile.commitIndex = uint(len(s.log))
		}
		for s.volatile.lastApplied < s.volatile.commitIndex {
			s.volatile.lastApplied += 1
			state.ApplyTransaction(s.log[s.volatile.lastApplied-1].trans)
		}

	}
	rep.MatchIndex = uint(len(s.log))
	return rep, nil
}

func sendAppendEntries(s *Sailor, state *storage.State, peer string) error {
	//Converted to 1 indexed
	am := appendMessage{}
	am.Term = s.currentTerm
	am.LeaderId = s.client.NodeName
	am.PrevLogIndex = s.leader.nextIndex[peer] - 1
	am.PrevLogTerm = s.log[s.leader.nextIndex[peer]-2].term
	am.Entries = s.log[s.leader.nextIndex[peer]-1:]
	am.LeaderCommit = s.volatile.commitIndex
	ap := messages.Message{}
	ap.Type = "appendEntries"
	ap.ID = 0 //TODO(JM): Figure out what this should be?
	ap.Source = s.client.NodeName
	ap.Value = makePayload(am)
	return s.client.SendToPeer(ap, peer)
}

func sendHeartbeats(s *Sailor, state *storage.State) error {
	for _, peer := range s.client.Peers {
		err := sendAppendEntries(s, state, peer)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleAppendReply(s *Sailor, state *storage.State, ar *appendReply, source string) error {
	//TODO(JM): Check for moving commit index?

	//TODO(MAX): Commit code
	if ar.Success {
		s.leader.nextIndex[source] = ar.MatchIndex + 1
		s.leader.matchIndex[source] = ar.MatchIndex
	} else {
		if ar.Term != s.currentTerm {
			s.becomeFollower(ar.Term)
			return nil
		}
		s.leader.nextIndex[source] -= 1
		return sendAppendEntries(s, state, source)
	}
	return nil
}
