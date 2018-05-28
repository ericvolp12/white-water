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
			err := s.handle_timeout()
			if err != nil {
				fmt.Printf("handle_timeout error: %v\n", err)
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
				case msg := <-appendEntry:
					//TODO(JM): Actually unpack message, waiting on function from Max
					val, err := handleAppendEntries(s, state, &appendMessage{})
					rep := makeReply(s, &msg, "appendReply")
					rep.Value = makePayload(val)
					rep.Error = err.Error()
					err = s.client.SendToPeers(rep, rep.Destination)
					if err != nil {
						//handle error
					}
					//Append handle - Joseph
				case msg := <-requestVote:
					s.resetTimer = true //restart timer
					if msg.Type == "requestVote" {
						err := s.handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Follower handle_requestVote Error: %v\n", err)
						}
					} // Don't respond to voteReply messages if follower
					//Vote handle - Max
				}
			case candidate:
				select {
				case _ = <-appendEntry:
					//Append handle - Joseph
				case msg := <-requestVote:
					if msg.Type == "requestVote" {
						err := s.handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Candidate handle_requestVote Error: %v\n", err)
						}
					} else { // Type == "voteReply"
						err := s.handle_voteReply(msg)
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
					rep.Error = err.Error()
					err = s.client.SendToBroker(rep)
					if err != nil {
						//handle error
					}
				case _ = <-sets:
					//Sets handle - Joseph
				case _ = <-appendEntry:
					//AppendReply handle - Joseph
				case msg := <-requestVote:
					//Vote
					if msg.Type == "requestVote" {
						err := s.handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Leader handle_requestVote Error:%v\n", err)
						}
					}
					//TODO Should votereplies w/ larger Term be considered?
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

// Decodes Zmq message.Value payload to whatever struct is passed into the func
func getPayload(value string, payload interface{}) error {
	temp, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		fmt.Printf("Decoding String Error:%v\n", err)
		return err
	}

	err2 := json.Unmarshal(temp, payload)
	if err2 != nil {
		fmt.Printf("Unmarshaling error: %v\n", err2)
		return err2
	}
	return nil
}

// Converts Sailor into follower state (Normally if msg.Term > s.currentTerm)
func (s *Sailor) becomeFollower(term uint) {
	s.currentTerm = term
	s.state = follower
	s.votedFor = ""
	s.numVotes = 0
	s.leader = nil
	s.resetTimer = true //always restart timer on failover
}
