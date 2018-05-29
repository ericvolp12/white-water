package raft

import (
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
)

func (s *Sailor) handle_timeout() error {
	fmt.Printf("in handle_timeout%s\n", s.client.NodeName)
	if s.state == leader {
		return sendHeartbeats(s)
	}
	// ONLY BECOME CANDIDATE IF ALLOWED
	s.state = candidate
	s.currentTerm += 1
	s.votedFor = s.client.NodeName
	s.numVotes = 1 // Votes for itself

	// Fill RequestVotes RPC struct
	newmsg := requestVote{}
	newmsg.Term = s.currentTerm
	newmsg.CandidateId = s.client.NodeName // Current Node ID
	last := uint(len(s.log))
	newmsg.LastLogIndex = last // Index of last entry in log
	if last == 0 {
		newmsg.LastLogTerm = 0
	} else {
		newmsg.LastLogTerm = s.log[last-1].term // The term of that entry index
	}
	// SEND newmsg REQUESTVOTE RPC BROADCAST
	// TODO (MD) this can be makeReply w/ nil
	zmqMsg := messages.Message{}
	zmqMsg.Type = "requestVote"
	zmqMsg.Source = s.client.NodeName
	zmqMsg.Value = makePayload(newmsg)
	fmt.Printf("Broadcasting\n")
	return s.client.Broadcast(zmqMsg)
}

func (s *Sailor) handle_requestVote(original_msg messages.Message) error {
	reqVoteRPC := requestVote{}
	err := getPayload(original_msg.Value, &reqVoteRPC) //Cast payload to requestVote
	if err != nil {
		fmt.Printf("getPayload error: %v\n", err)
	}
	reply_payload := reply{}
	if reqVoteRPC.Term > s.currentTerm {
		s.becomeFollower(reqVoteRPC.Term)
	}

	reply_payload.Term = s.currentTerm
	if s.state == candidate || reqVoteRPC.Term < s.currentTerm {
		reply_payload.VoteGranted = false
	}

	if s.votedFor == "" || s.votedFor == reqVoteRPC.CandidateId {
		recent := uint(len(s.log))
		if reqVoteRPC.LastLogTerm > s.log[recent].term || reqVoteRPC.LastLogIndex >= recent {
			reply_payload.VoteGranted = true
			s.votedFor = reqVoteRPC.CandidateId
		} else {
			reply_payload.VoteGranted = false
		}
	} else {
		reply_payload.VoteGranted = false
	}
	zmq_msg := makeReply(s, &original_msg, "voteReply")
	zmq_msg.Value = makePayload(reply_payload)
	return s.client.SendToPeer(zmq_msg, original_msg.Source)
}

func (s *Sailor) handle_voteReply(original_msg messages.Message, timeouts chan bool) error {
	reply := reply{}
	err := getPayload(original_msg.Value, &reply)
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
	// TODO Check term stuff? Maybe convert to follower
	if reply.VoteGranted == true {
		s.numVotes += 1
	}
	if s.numVotes > ((len(s.client.Peers) + 1) / 2) { // become leader, send empty heartbeat
		s.state = leader
		fmt.Printf("LEADER ELECTED\n")
		timeouts <- false // Triggers timer thread to restart timer as leader
		newmsg := appendMessage{}
		newmsg.Term = s.currentTerm
		newmsg.LeaderId = s.client.NodeName

		last := uint(len(s.log))
		if last == 0 {
			newmsg.PrevLogTerm = 0
		} else {
			newmsg.PrevLogTerm = s.log[last-1].term
		}
		newmsg.PrevLogIndex = last
		newmsg.Entries = nil
		newmsg.LeaderCommit = s.volatile.commitIndex

		// TODO (MD) makeReply w/ nil dest
		zmqmsg := messages.Message{}
		zmqmsg.Type = "appendEntries"
		zmqmsg.Source = s.client.NodeName
		zmqmsg.Value = makePayload(newmsg)
		return s.client.Broadcast(zmqmsg)
	}
	return nil
}
