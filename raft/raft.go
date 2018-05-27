package raft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry chan messages.Message, timeouts chan bool, state *storage.State) {
	for {
		select {
		case _ = <-timeouts:
			//timeouts message handle
			err := handle_timeout()
			if err != nil {
				// handle error?
				fmt.Printf("handle_timeout error\n")
			}
			//Max
		default:
			switch s.state {
			case follower:
				select {
				case msg := <-gets:
					//TODO(JM): Decide where to put leader-notify
					val, err := handleGetRequest(msg.Key, s, state)
					rep := makeReply(s, &msg, "getReply")
					rep.Key = msg.Key
					rep.Value = val
					rep.Error = err.Error()
					err = s.client.SendToBroker(rep)
					if err != nil {
						//handle error
					}
				case _ = <-sets:

					//Sets handle - Joseph
				case _ = <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					// TODO (MD) check when s.lastMessageTime should be set when voting
					if msg.Type == "requestVote" {
						err := handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Follower handle_requestVote Error: %v\n", err)
						}
					}
					//Vote handle - Max
				}
			case candidate:
				select {
				case _ = <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					if msg.Type == "requestVote" {
						err := handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Candidate handle_requestVote Error: %v\n", err)
						}
					} else { // Type == "voteReply"
						err := handle_voteReply(msg)
						if err != nil {
							fmt.Printf("Candidate handle_voteReply Error: %v\n", err)
						}
					}
					//VoteReply handle - Max
				}
			case leader:
				select {
				case msg := <-gets:
					val, err := handleGetRequest(msg.Key, s, state)
					rep := makeReply(s, &msg, "getReply")
					rep.Key = msg.Key
					rep.Value = val
					rep.Error = err
					err = s.client.sendMessage(rep)
					if err != nil {
						//handle error
					}
				case _ = <-sets:
					//Sets handle - Joseph
				case _ = <-appendEntry:
					//AppendReply handle - Joseph
				case _ = <-requestVote:
					//Vote/VoteReply handle - Max
					if msg.Type == "requestVote" {
						err := handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Leader handle_requestVote Error:%v\n", err)
						}
					}
					// Ignore vote replies if in leader state

					//Propose-value?
				}
			}
		}
	}
}

func makeReply(s *Sailor, msg *messages.Message, typestr string) messages.Message {
	rep := messages.Message{}
	rep.Type = typestr
	rep.ID = 0
	rep.Destination = []string{msg.Source}
	rep.Source = s.client.NodeName
	return rep
}

// Encodes different Raft message RPC structs into strings for ZMQ type messages
func makePayload(payload interface{}) string {
	temp, err := json.Marshal(payload) // Encodes to slice of bytes
	if err != nil {
		fmt.Printf("Marshaling error: %v\n", err)
		return ""
	} else {
		return base64.StdEncoding.EncodeToString(temp) //Encodes to string
	}
	return "" //TODO: Make it return an error
}

func getPayload(value string) interface{} {
	temp, err := base64.StdDecoding.DecodeString(value)
	if err != nil {
		fmt.Printf("Decoding String Error:%v\n", err)
		return nil
	}
	payload, err2 := json.Unmarshal(temp)
	if err2 != nil {
		fmt.Printf("Unmarshaling error: %v\n", err2)
		return nil
	}
	return payload
}

// Converts Sailor into follower state (Normally if msg.Term > s.currentTerm)
func (s *Sailor) becomeFollower(term int) {
	s.currentTerm = term
	s.state = follower
	s.votedFor = nil
	s.numVotes = 0
	s.leader = nil
}
