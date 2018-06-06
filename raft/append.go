package raft

// append.go is a file that handles all the logic of the
// processing appendEntries RPC calls, as described in the raft paper

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// handleAppendEntries gets called when a follower recieves an appendEntries RPC.
// It is responsible for processing it according to the Raft paper.
// This involves checking the state to transition to follower,
// rejecting the RPC, updating the log, and commiting to the state machine.
// s is the Sailor struct that holds the Raft info, state is the state machine,
// am is the actual RPC struct, leaderId is the ID of the node who sent the RPC.
func handleAppendEntries(s *Sailor, state *storage.State, am *appendMessage, leaderId string) (appendReply, error) {
	// Check if the node needs to convert to the follower state
	if (s.state != follower && am.Term == s.currentTerm) || am.Term > s.currentTerm {
		s.becomeFollower(am.Term)
	}

	rep := appendReply{Term: s.currentTerm, Success: false}
	// Reject the RPC if it has the wrong term
	if s.currentTerm > am.Term {
		return rep, nil
	}
	s.timer.Reset(new_time()) // The message is from the leader, reset our election-start clock
	s.leaderId = leaderId
	// Reject the RPC if the position or the term is wrong
	if am.PrevLogIndex != 0 && (len(s.log) <= int(am.PrevLogIndex-1) ||
		(len(s.log) > 0 && s.log[am.PrevLogIndex-1].Term != am.PrevLogTerm)) {
		return rep, nil
	}

	rep.Success = true // The RPC is valid and the call will succeed

	rep.PrepLower = am.PrevLogIndex + 1
	rep.ComLower = s.volatile.commitIndex

	// Loop over the values we're dropping from our log. If we were the leader when
	// the values were proposed, then we reply to the client with a set failed message.
	// We know we were the leader if we were counting votes for that log entry.
	for i := am.PrevLogIndex; i < uint(len(s.log)); i++ {
		if s.log[i].votes != 0 {
			fail := messages.Message{}
			fail.Source = s.client.NodeName
			fail.Type = "setResponse"
			fail.Error = "SET FAILED"
			fail.ID = s.log[i].Id
			fail.Key = s.log[i].Trans.Key
			s.client.SendToBroker(fail)
		}
	}
	// Drops the extra entries and adds the new ones
	s.log = append(s.log[:am.PrevLogIndex], am.Entries...)
	// If we have new values to commit
	if am.LeaderCommit > s.volatile.commitIndex {
		// Choose the min of our log size and the leader's commit index
		if int(am.LeaderCommit) <= len(s.log) {
			s.volatile.commitIndex = am.LeaderCommit
		} else {
			s.volatile.commitIndex = uint(len(s.log))
		}
		// Actually commit and apply the transactions to the state machine
		for s.volatile.lastApplied < s.volatile.commitIndex {
			s.volatile.lastApplied += 1
			state.ApplyTransaction(s.log[s.volatile.lastApplied-1].Trans)
		}

	}
	rep.PrepUpper = uint(len(s.log))
	rep.ComUpper = s.volatile.commitIndex
	return rep, nil
}

// sendAppendEntries takes a sailor and a node-name.
// It then sends the appropriate appendEntries RPC to that node.
func sendAppendEntries(s *Sailor, peer string) error {
	am := appendMessage{}
	am.Term = s.currentTerm
	am.LeaderId = s.client.NodeName
	am.PrevLogIndex = s.leader.nextIndex[peer] - 1
	// This is just some fancy logic to check for the bounds on the log
	// e.g. our log has 0 entries, so the prevEntryTerm cannot be pulled from the log
	if len(s.log) == 0 {
		am.PrevLogTerm = 0
		am.Entries = nil
	} else {
		// If our log is too short to have prevTerm, use 0
		if int(s.leader.nextIndex[peer])-2 < 0 {
			am.PrevLogTerm = 0
		} else {
			am.PrevLogTerm = s.log[s.leader.nextIndex[peer]-2].Term
		}
		// If our nextIndex is a value we don't have yet, send nothing
		if s.leader.nextIndex[peer] > uint(len(s.log)) {
			am.Entries = []entry{}
		} else {
			am.Entries = s.log[s.leader.nextIndex[peer]-1:]
		}
	}

	am.LeaderCommit = s.volatile.commitIndex
	ap := messages.Message{}
	ap.Type = "appendEntries"
	ap.ID = 0
	ap.Source = s.client.NodeName
	ap.Value = makePayload(am)
	return s.client.SendToPeer(ap, peer)
}

// sendHeartbeats takes a pointer to a sailor struct and uses it
// to send the heartbeat RPCs to every node on the network.
func sendHeartbeats(s *Sailor) error {
	for _, peer := range s.client.Peers {
		err := sendAppendEntries(s, peer)
		if err != nil {
			return err
		}
	}
	return nil
}

// handleAppendReply takes a sailor, the state, a reply to an appendEntries RPC,
// and a node-name as the source of the reply. It determines what to do with the result,
// as defined in the Raft paper.
func handleAppendReply(s *Sailor, state *storage.State, ar *appendReply, source string) error {
	if ar.Success {

		_ = s.handle_prepare(ar.PrepLower, ar.PrepUpper)
		_ = s.handle_commit(ar.ComLower, ar.ComUpper, state)

		s.leader.nextIndex[source] = ar.PrepUpper + 1
		s.leader.matchIndex[source] = ar.PrepUpper
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
