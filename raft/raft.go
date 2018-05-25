package raft

import (
    "fmt"
    "time"
    "encoding/json"
    "encoding/base64"
)

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry *chan messages.Message, timeouts *chan bool) {
	for {
		select {
		case msg := <-*timeouts:
			//timeouts message handle
			//Max
		default:
			select {
			case s.state == follower:
				select {
				//sets - J
				//gets - J
				//append - J
				//requestVote - M
				}
			case s.state == candidate:
				select {
				//requestVote - M
				//append -J
				}
			case s.state == leader:
				select {
				//gets -J
				//sets -J
				//append -J
				//requestVote -M
				//Propose-value?
				}
			}
		}
	}
}

// Encodes different Raft message RPC structs into strings for ZMQ type messages
func makePayload(payload interface{}) string {
    temp, err := json.Marshal(payload)
    if err != nil {
        fmt.Printf("RequestVote RPC ERROR in handle_timeout\n")
    } else {
        return base64.StdEncoding.EncodeToString(temp)
    }
}


func (s *Sailor) handle_timeout() {
    // ONLY BECOME CANDIDATE IF ALLOWED
    s.state = candidate
    s.currentTerm += 1
    s.numVotes = 1 // Votes for itself
    s.lastMessageTime = time.Now() // Triggers timer thread to restart timer

    // Fill RequestVotes RPC struct
    newmsg := make(requestVote)
    newmsg.term = s.currentTerm
    newmsg.candidateId = 0 // Current Node ID
    newmsg.lastLogIndex = len(s.log) - 1    // Index of last entry in log
    newmsg.lastLogTerm = s.log[newmsg.lastLogIndex].term    // The term of that entry index

    // SEND newmsg REQUESTVOTE RPC BROADCAST
    zmqMsg := messages.Message{}
    zmqMsg.Type = "requestVote"
    zmqMsg.Source = newmsg.client.NodeName
    zmqMsg.Value = makePayload(newmsg)
    s.client.Broadcast(zmqMsg)
}
