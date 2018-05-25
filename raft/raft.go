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
            handle_timeout()
			//Max
		default:
			select {
			case s.state == follower:
				select {
				case msg := <-gets:
					//TODO(JM): Decide where to put leader-notify
					val, err := handleGetRequest(msg.key, s, state)
					rep := makeReply(s, msg, "getReply")
					rep.Key = msg.key
					rep.Value = val
					rep.Error = err
					err := s.client.sendMessage(rep)
					if err != nil {
						//handle error
					}
				case msg := <-sets:

					//Sets handle - Joseph
				case msg := <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					//Vote handle - Max
                    // If msg.Type == "voteReply" -> 
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
					val, err := handleGetRequest(msg.key, s, state)
					rep := makeReply(s, msg, "getReply")
					rep.Key = msg.key
					rep.Value = val
					rep.Error = err
					err := s.client.sendMessage(rep)
					if err != nil {
						//handle error
					}
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

func (s *Sailor) becomeFollower(term int) {
    s.currentTerm = term
    s.state = follower
    s.votedFor = nil
}
