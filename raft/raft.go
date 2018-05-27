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
				case _ = <-requestVote:
					//Vote handle - Max
				}
			case candidate:
				select {
				case _ = <-appendEntry:
					//Append handle - Joseph
				case _ = <-requestVote:
					//VoteReply handle - Max
				}
			case leader:
				select {
				case msg := <-gets:
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
					//AppendReply handle - Joseph
				case _ = <-requestVote:
					//Vote/VoteReply handle - Max
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
	temp, err := json.Marshal(payload)
	if err != nil {
		fmt.Printf("RequestVote RPC ERROR in handle_timeout\n")
	} else {
		return base64.StdEncoding.EncodeToString(temp)
	}
	return "" //TODO: Make it return an error
}
