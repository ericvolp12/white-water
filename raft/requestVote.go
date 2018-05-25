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
	newmsg.CandidateId = 0                               // Current Node ID
	newmsg.LastLogIndex = len(s.log) - 1// Index of last entry in log
	newmsg.LastLogTerm = s.log[newmsg.LastLogIndex].term // The term of that entry index

	// SEND newmsg REQUESTVOTE RPC BROADCAST
	zmqMsg := messages.Message{}
	zmqMsg.Type = "requestVote"
	zmqMsg.Source = newmsg.client.NodeName
	zmqMsg.Value = makePayload(newmsg)
	return s.client.Broadcast(zmqMsg)
}

//TODO: NEED ZMQ message DECODER to get proper dst
func (s *Sailor) handle_requestVote(msg *requestVote) {
    payload := make(reply)
    if msg.Term > s.currentTerm {
        s.becomeFollower(msg.Term)
    }

    payload.Term = s.currentTerm
    if s.state == candidate || msg.Term < s.currentTerm {
        payload.VoteGranted = false
    }

    if s.votedFor == nil || s.votedFor == msg.candidateId {
        recent := len(s.log) - 1
        if msg.LastLogTerm > s.log[recent].term || msg.LastLogIndex >= recent {
            payload.VoteGranted = true
            s.votedFor = msg.candidateId
        } else {
            payload.VoteGranted = false
        }
    } else {
        payload.VoteGranted = false
    }
    // NEED TO BUILD NEW MESSAGE
    // zmqMsg.Type = "voteReply"
}
