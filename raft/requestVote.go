package raft

func (s *Sailor) handle_timeout() error {
	// ONLY BECOME CANDIDATE IF ALLOWED
	s.state = candidate
	s.currentTerm += 1
	s.votedFor = s.client.NodeName
	s.numVotes = 1                 // Votes for itself
	s.lastMessageTime = time.Now() // Triggers timer thread to restart timer

	// Fill RequestVotes RPC struct
	newmsg := make(requestVote)
	newmsg.Term = s.currentTerm
	newmsg.CandidateId = s.client.NodeName               // Current Node ID
	newmsg.LastLogIndex = len(s.log) - 1                 // Index of last entry in log
	newmsg.LastLogTerm = s.log[newmsg.LastLogIndex].term // The term of that entry index

	// SEND newmsg REQUESTVOTE RPC BROADCAST
	// TODO (MD) this can be makeReply w/ nil
	zmqMsg := messages.Message{}
	zmqMsg.Type = "requestVote"
	zmqMsg.Source = newmsg.client.NodeName
	zmqMsg.Value = makePayload(newmsg)
	return s.client.Broadcast(zmqMsg)
}

//TODO: NEED ZMQ message DECODER to get proper dst
func (s *Sailor) handle_requestVote(original_msg messages.Message) error {
	reqVoteRPC := getPayload(original_msg).(requestVote) //Cast payload to requestVote
	reply_payload := make(reply)
	if reqVoteRPC.Term > s.currentTerm {
		s.becomeFollower(msg.Term)
	}

	reply_payload.Term = s.currentTerm
	if s.state == candidate || reqVoteRPC.Term < s.currentTerm {
		reply_payload.VoteGranted = false
	}

	if s.votedFor == nil || s.votedFor == reqVoteRPC.CandidateId {
		recent := len(s.log) - 1
		if reqVoteRPC.LastLogTerm > s.log[recent].term || reqVoteRPC.LastLogIndex >= recent {
			reply_payload.VoteGranted = true
			s.votedFor = reqVoteRPC.CandidateId
		} else {
			reply_payload.VoteGranted = false
		}
	} else {
		reply_payload.VoteGranted = false
	}
	zmq_msg := makeReply(s, msg, "voteReply")
	zmq_msg.Value = makePayload(reply_payload)
	return s.client.SendToPeer(zmq_msg, original_msg.Source)
}

func (s *Sailor) handle_voteReply(original_msg messages.Message) error {
	reply := getPayload(original_msg)
	if reply.Term > s.currentTerm {
		s.BecomeFollower()
		return nil
	}
	if reply.Term < s.currentTerm { //Ignore old votes
		return nil
	}
	// TODO Check term stuff? Maybe convert to follower
	if reply.VoteGranted == true {
		s.numVotes += 1
	}
	if s.numVotes > (len(s.client.Peers)+1)/2 { // become leader, send empty heartbeat
		s.state = leader
		newmsg := make(appendMessage)
		newmsg.Term = s.currentTerm
		newmsg.LeaderId = s.client.NodeName

		recent = len(s.log) - 1
		if recent == -1 {
			newmsg.PrevLogTerm = -1
		} else {
			newmsg.PrevLogTerm = s.log[recent].term
		}
		newmsg.PrevLogIndex = recent
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
