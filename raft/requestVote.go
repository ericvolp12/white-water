package raft

// requestVote.go
// This file contains all functions associated with the Raft algorithm's
// election process.

import (
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// This function is called when a timeout has occured. If the node is currently
// the leader, the function initiates a round of heartbeat messages to be sent
// out to the other nodes in the network. If the node is currently a follower
// or a candidate, it initiates a new election cycle by becoming a candidate,
// updating the current term, and broadcasting requestVote messages to all other
// nodes on the network. It returns an error if one was encountered while
// attempting to send messages to another node.
// Used in raft.go:MsgHandler()
func (s *Sailor) handle_timeout() error {
	if s.state == leader {
		s.timer.Reset(leaderReset())
		return sendHeartbeats(s)
	}

	fmt.Printf("Becoming candidate! %s\n", s.client.NodeName)
	s.timer.Reset(new_time())
	s.state = candidate
	s.currentTerm += 1
	s.votedFor = s.client.NodeName // Voting for itself
	s.numVotes = 1
	s.leaderId = ""

	// Fill RequestVotes RPC struct
	newmsg := requestVote{}
	newmsg.Term = s.currentTerm
	newmsg.CandidateId = s.client.NodeName // Current Node ID
	last := uint(len(s.log))
	newmsg.LastLogIndex = last // Index of last entry in log
	if last == 0 {
		newmsg.LastLogTerm = 0 // No prior log term if there are no prior entries
	} else {
		newmsg.LastLogTerm = s.log[last-1].Term // The term of the last entry
	}
	zmqMsg := messages.Message{}
	zmqMsg.Type = "requestVote"
	zmqMsg.Source = s.client.NodeName
	zmqMsg.Value = makePayload(newmsg)
	return s.client.Broadcast(zmqMsg)
}

// Given a requestVote message type, this function will process the information
// in the message and reply to the original message's source node (the candidate)
// with a confirmation of a vote or a rejection if appropriate. The function
// will also make appropriate changes to a nodes state where appropriate (
// such as failover to a follower state), and returns an error if the node
// failed to send a reply back to the candidate
// Used in raft.go: handle_follower(), handle_leader(), and handle_candidate()
func (s *Sailor) handle_requestVote(original_msg messages.Message) error {
	reqVoteRPC := requestVote{}
	err := getPayload(original_msg.Value, &reqVoteRPC) //Cast payload to requestVote
	if err != nil {
		fmt.Printf("getPayload error: %v\n", err)
	}
	reply_payload := reply{}
	if reqVoteRPC.Term > s.currentTerm { // Become a follower for any new election cycle
		s.becomeFollower(reqVoteRPC.Term)
	}

	reply_payload.Term = s.currentTerm
	//if the node just became a follower or already voted for the candidate previously
	if s.votedFor == "" || s.votedFor == reqVoteRPC.CandidateId {
		recent := uint(len(s.log) - 1)
		// if the candidate indicates it has a log at least as updated as the node (or more)
		if s.log == nil || reqVoteRPC.LastLogTerm > s.log[recent].Term ||
			(reqVoteRPC.LastLogTerm == s.log[recent].Term && reqVoteRPC.LastLogIndex >= recent+1) {
			reply_payload.VoteGranted = true
			s.votedFor = reqVoteRPC.CandidateId
			s.timer.Reset(new_time())
		} else {
			reply_payload.VoteGranted = false
		}
	} else {
		reply_payload.VoteGranted = false
	}

	// Extra check to ensure the requestVote is not old or node votes for itself
	if s.state == candidate || reqVoteRPC.Term < s.currentTerm {
		reply_payload.VoteGranted = false
	}

	zmq_msg := makeReply(s, &original_msg, "voteReply")
	zmq_msg.Value = makePayload(reply_payload)
	return s.client.SendToPeer(zmq_msg, original_msg.Source)
}

// This function is used to process vote replies from other nodes on the
// network. It will get the payload of a voteReply message, and process it
// accordingly. If this candidate has recieved enough votes from other nodes
// in the network (a majority), then it becomes the new leader and sends an
// empty heartbeat to the other nodes to indicate the state change. The
// function returns an error if one was encountered while sending the first
// heartbeat. Only used by a candidate- if the node is already a leader, it
// does not need this reply.
// Used in raft.go:handle_candidate()
func (s *Sailor) handle_voteReply(original_msg messages.Message) error {
	reply := reply{}
	err := getPayload(original_msg.Value, &reply) //Put payload into reply form
	if err != nil {
		fmt.Printf("getPayload Error: %v\n", err)
	}
	if reply.Term > s.currentTerm {
		s.becomeFollower(reply.Term)
		return nil
	}
	if reply.Term < s.currentTerm { //Ignore old votes
		return nil
	}

	if reply.VoteGranted == true {
		s.numVotes += 1
	}
	// If candidate has a majority of votes
	if s.numVotes > ((len(s.client.Peers) + 1) / 2) {
		fmt.Printf("Becoming leader! %s, %+v\n", s.client.NodeName, s.log)
		s.timer.Reset(leaderReset()) // starting the heartbeat timer
		s.state = leader
		s.leader = &leaderState{} // Reseting the s.leader structure
		s.leader.nextIndex = make(map[string]uint)
		for _, peer := range s.client.Peers {
			s.leader.nextIndex[peer] = uint(len(s.log) + 1) //2
		}
		s.leader.matchIndex = make(map[string]uint)
		for _, peer := range s.client.Peers {
			s.leader.matchIndex[peer] = 0
		}

		trans := storage.GenerateTransaction(storage.GetOp, "no-op", "")
		newEntry := entry{Term: s.currentTerm, Trans: trans, votes: 1, Id: -1}
		s.log = append(s.log, newEntry)

		err := sendHeartbeats(s)
		if err != nil {
			return err
		}
	}
	return nil
}
