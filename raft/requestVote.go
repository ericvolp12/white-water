package raft

import (
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
)

func (s *Sailor) handle_timeout() error {
	if s.state == leader {
		return sendHeartbeats(s)
	}
	// ONLY BECOME CANDIDATE IF ALLOWED
	s.state = candidate
	fmt.Printf("Becoming candidate %s\n", s.client.NodeName)
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
		recent := uint(len(s.log) - 1)
		if s.log == nil || reqVoteRPC.LastLogTerm > s.log[recent].term || reqVoteRPC.LastLogIndex >= recent+1 {
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
		fmt.Printf("Becoming leader! %s", s.client.NodeName)
		s.state = leader
		s.leader = &leaderState{}
		s.leader.nextIndex = make(map[string]uint)
		for _, peer := range s.client.Peers {
			s.leader.nextIndex[peer] = uint(len(s.log) + 2)
		}
		s.leader.matchIndex = make(map[string]uint)
		for _, peer := range s.client.Peers {
			s.leader.matchIndex[peer] = 0
		}
		timeouts <- false // Triggers timer thread to restart timer as leader
		timeouts <- false // Triggers the leader timer to start

		err := sendHeartbeats(s)
		if err != nil {
			fmt.Printf("Error in voteReply: %+v", err)
		}

	}
	return nil
}
