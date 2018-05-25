package raft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry chan messages.Message, timeouts chan bool) {
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
					//Gets handle - Joseph
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
