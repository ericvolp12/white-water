package raft

import (
	//"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func handleAppendEntries(s *Sailor, state *storage.State, am *appendMessage, leaderId string) (appendReply, error) {
	if (s.state != follower && am.Term == s.currentTerm) || am.Term > s.currentTerm {
		s.becomeFollower(am.Term)
	}
	//Converted to 1 indexed
	rep := appendReply{Term: s.currentTerm, Success: false}
	if s.currentTerm > am.Term { //TODO(JM): Update term
		return rep, nil
	}
	s.timer.Reset(new_time())
	s.leaderId = leaderId
	//	fmt.Printf("prevLogIndex %d", am.PrevLogIndex)
	if am.PrevLogIndex != 0 && (len(s.log) <= int(am.PrevLogIndex-1) || (len(s.log) > 0 && s.log[am.PrevLogIndex-1].Term != am.PrevLogTerm)) {
		return rep, nil
	}

	rep.Success = true

	rep.PrepLower = am.PrevLogIndex + 1
	rep.ComLower = s.volatile.commitIndex
	//fmt.Printf("log len= %d, prevlogindex= %d\n", len(s.log), am.PrevLogIndex)
	s.log = append(s.log[:am.PrevLogIndex], am.Entries...)
	if am.LeaderCommit > s.volatile.commitIndex {
		//fmt.Printf("%s, %+v, %+v, %+v\n", s.client.NodeName, am, s.log, s)
		if int(am.LeaderCommit) <= len(s.log) {
			s.volatile.commitIndex = am.LeaderCommit
		} else {
			s.volatile.commitIndex = uint(len(s.log))
		}
		for s.volatile.lastApplied < s.volatile.commitIndex {
			s.volatile.lastApplied += 1
			//TODO(JM): Update log entry term to leader log entry
			state.ApplyTransaction(s.log[s.volatile.lastApplied-1].Trans)
		}

	}
	rep.PrepUpper = uint(len(s.log))
	rep.ComUpper = s.volatile.commitIndex
	return rep, nil
}

func sendAppendEntries(s *Sailor, peer string) error {
	//Converted to 1 indexed
	am := appendMessage{}
	am.Term = s.currentTerm
	am.LeaderId = s.client.NodeName
	am.PrevLogIndex = s.leader.nextIndex[peer] - 1
	//fmt.Printf("Log: %+v, s.leader.nextIndex[peer]-2: %d", s.log, s.leader.nextIndex[peer]-2)
	if len(s.log) == 0 {
		am.PrevLogTerm = 0
		am.Entries = nil
	} else {
		//fmt.Printf("nextIndex of peer: %d, peer: %s\n", s.leader.nextIndex[peer], peer)
		if len(s.log) == 1 || int(s.leader.nextIndex[peer])-2 <= 0 {
			am.PrevLogTerm = 0
		} else {
			am.PrevLogTerm = s.log[s.leader.nextIndex[peer]-2].Term
		}
		//fmt.Printf("s.leader.nextIndex[%s]: %d, log: %+v\n", peer, s.leader.nextIndex[peer], s.log)
		if s.leader.nextIndex[peer] >= uint(len(s.log)) {
			am.Entries = []entry{}
		} else {
			am.Entries = s.log[s.leader.nextIndex[peer]-1:]
		}
	}

	am.LeaderCommit = s.volatile.commitIndex
	ap := messages.Message{}
	ap.Type = "appendEntries"
	ap.ID = 0 //TODO(JM): Figure out what this should be?
	ap.Source = s.client.NodeName
	ap.Value = makePayload(am)
	return s.client.SendToPeer(ap, peer)
}

func sendHeartbeats(s *Sailor) error {
	for _, peer := range s.client.Peers {
		err := sendAppendEntries(s, peer)
		if err != nil {
			return err
		}
	}
	return nil
}

func handleAppendReply(s *Sailor, state *storage.State, ar *appendReply, source string) error {
	if ar.Success {

		_ = s.handle_prepare(ar.PrepLower, ar.PrepUpper)
		_ = s.handle_commit(ar.ComLower, ar.ComUpper, state)

		s.leader.nextIndex[source] = ar.PrepUpper + 1
		s.leader.matchIndex[source] = ar.PrepUpper
		//s.handle_commit(ar.MatchIndex) // TODO MAKE SUER THIS WORKS
	} else {
		if ar.Term != s.currentTerm {
			s.becomeFollower(ar.Term)
			return nil
		}
		s.leader.nextIndex[source] -= 1
		return sendAppendEntries(s, source)
	}
	return nil
}
