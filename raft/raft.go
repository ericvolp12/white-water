package raft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry chan messages.Message, timeouts chan bool, state *State) {
	for {
		select {
		case msg := <-timeouts:
			//timeouts message handle
			//Max
		default:
			select {
			case s.state == follower:
				select {
				case msg := <-gets:
					val, err := handleGetRequest(msg.key, s, state)
					rep := makeReply(s, msg, "getReply")
					rep.Key = msg.key
					rep.Value = val
					rep.Error = err
					err := s.client.sendMessage(rep)
					if err != nil {
						//handle error
					}
					//Gets handle - Joseph
				case msg := <-sets:
					//Sets handle - Joseph
				case msg := <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					//Vote handle - Max
				}
			case s.state == candidate:
				select {
				case msg := <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					//VoteReply handle - Max
				}
			case s.state == leader:
				select {
				case msg := <-gets:
					//Gets handle - Joseph
				case msg := <-sets:
					//Sets handle - Joseph
				case msg := <-appendEntry:
					//AppendReply handle - Joseph
				case msg := <-requestVote:
					//Vote/VoteReply handle - Max

					//Propose-value?
				}
			}
		}
	}
}

func makeReply(s *Sailor, msg *messages.Message, typestr string) *messages.Message {
	rep := messages.Message{}
	rep.Type = typestr
	rep.ID = 0
	rep.Destination = []string{msg.Source}
	rep.Source = s.client.NodeName
	return rep
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
	s.numVotes = 1                 // Votes for itself
	s.lastMessageTime = time.Now() // Triggers timer thread to restart timer

	// Fill RequestVotes RPC struct
	newmsg := make(requestVote)
	newmsg.Term = s.currentTerm
	newmsg.CandidateId = 0                               // Current Node ID
	newmsg.LastLogIndex = len(s.log) - 1                 // Index of last entry in log
	newmsg.LastLogTerm = s.log[newmsg.lastLogIndex].term // The term of that entry index

	// SEND newmsg REQUESTVOTE RPC BROADCAST
	zmqMsg := messages.Message{}
	zmqMsg.Type = "requestVote"
	zmqMsg.Source = newmsg.client.NodeName
	zmqMsg.Value = makePayload(newmsg)
	s.client.Broadcast(zmqMsg)
}
