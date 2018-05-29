package raft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry chan messages.Message, timereset, timeouts chan bool, state *storage.State) {
	for {
		select {
		case msg := <-timeouts:
			//timeouts message handle
			fmt.Printf("%s timeout handler got %s", s.client.NodeName, msg)
			err := s.handle_timeout()
			timereset <- false // Triggers timer thread to restart timer
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
					rep := makeReply(s, &msg, "getResponse")
					rep.Key = msg.Key
					if err != nil {
						rep.Error = err.Error()
					} else {
						rep.Value = val
					}
					err = s.client.SendToBroker(rep)
					if err != nil {
						//handle error
					}
				case _ = <-sets:

					//Sets handle - Joseph
				case msg := <-appendEntry:
					timereset <- false // Triggers timer thread to restart timer
					if msg.Type == "appendEntries" {
						am := appendMessage{}
						getPayload(msg.Value, &am)
						val, err := handleAppendEntries(s, state, &am)
						rep := makeReply(s, &msg, "appendReply")
						rep.Value = makePayload(val)
						//rep.Error = err.Error()
						err = s.client.SendToPeers(rep, rep.Destination)
						if err != nil {
							fmt.Printf("follower, append entries: %s", err)
						}
					}
				case msg := <-requestVote:
					timereset <- false //restart timer
					if msg.Type == "requestVote" {
						err := s.handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Follower handle_requestVote Error: %v\n", err)
						}
					} // Don't respond to voteReply messages if follower
				//Vote handle - Max
				default:
				}

			case candidate:
				select {
				case msg := <-appendEntry:
					timereset <- false // Triggers timer thread to restart timer
					if msg.Type == "appendEntries" {
						am := appendMessage{}
						getPayload(msg.Value, &am)
						val, err := handleAppendEntries(s, state, &am)
						rep := makeReply(s, &msg, "appendReply")
						rep.Value = makePayload(val)
						//rep.Error = err.Error()
						err = s.client.SendToPeers(rep, rep.Destination)
						if err != nil {
							fmt.Printf("candidate, append entries: %s", err)
						}
					}
				case msg := <-requestVote:
					if msg.Type == "requestVote" {
						err := s.handle_requestVote(msg)
						if err != nil {
							fmt.Printf("Candidate handle_requestVote Error: %v\n", err)
						}
					} else { // Type == "voteReply"
						err := s.handle_voteReply(msg, timereset)
						if err != nil {
							fmt.Printf("Candidate handle_voteReply Error: %v\n", err)
						}
					}
				default:

					//VoteReply handle - Max
				}
			case leader:
				select {
				case msg := <-gets:
					val, err := handleGetRequest(msg.Key, s, state)
					rep := makeReply(s, &msg, "getResponse")
					rep.Key = msg.Key
					if err != nil {
						rep.Error = err.Error()
					} else {
						rep.Value = val
					}
					err = s.client.SendToBroker(rep)
					if err != nil {
						//handle error
					}
				case msg := <-sets:
					s.handle_set(msg, state)
					//Sets handle - Joseph
				case msg := <-appendEntry:
					if msg.Type == "appendEntries" {
						am := appendMessage{}
						getPayload(msg.Value, &am)
						val, err := handleAppendEntries(s, state, &am)
						if err != nil {
							fmt.Printf("leader, append entries handle: %s", err)
						}
						rep := makeReply(s, &msg, "appendReply")
						rep.Value = makePayload(val)
						//rep.Error = err.Error()
						err = s.client.SendToPeers(rep, rep.Destination)
						if err != nil {
							fmt.Printf("leader, append entries: %s", err)
						}
					} else if msg.Type == "appendReply" {
						ar := appendReply{}
						getPayload(msg.Value, &ar)
						err := handleAppendReply(s, state, &ar, msg.Source)
						if err != nil {
							fmt.Printf("leader, append reply handle: %s", err)
						}
					}
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
				default:

				}
			default:
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
	fmt.Printf("Becoming follower! %s\n", s.client.NodeName)
	s.currentTerm = term
	s.state = follower
	s.votedFor = ""
	s.numVotes = 0
	s.leader = nil
}
