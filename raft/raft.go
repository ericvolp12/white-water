package raft

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func (s *Sailor) MsgHandler(state *storage.State) {
	s.timer = time.NewTimer(time.Second * 2)
	for {
		select {
		case <-s.timer.C:
			s.handle_timeout()
		default:
			msg := s.client.ReceiveMessage()
			if msg != nil {
				switch s.state {
				case leader:
					s.handle_leader(*msg, state)
				case follower:
					s.handle_follower(*msg, state)
				case candidate:
					s.handle_candidate(*msg, state)
				}
			} else {
				time.Sleep(time.Millisecond * 5)
			}
		}
	}
}

func (s *Sailor) handle_follower(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "get":
		err := s.getAttempt(msg, state)
		if err != nil {
			fmt.Printf("Candidate getAttempt error: %v\n", err)
		}
	case "set":
		err := s.setReject(&msg)
		if err != nil {
			fmt.Printf("Follower Reject error: %v\n", err)
		}
		//Sets handle - Max
	case "appendEntries":
		//timereset <- false // Triggers timer thread to restart timer
		am := appendMessage{}
		getPayload(msg.Value, &am)
		val, err := handleAppendEntries(s, state, &am, msg.Source)
		rep := makeReply(s, &msg, "appendReply")
		rep.Value = makePayload(val)
		//rep.Error = err.Error()
		err = s.client.SendToPeers(rep, rep.Destination)
		if err != nil {
			fmt.Printf("follower, append entries: %s", err)
		}
	case "requestVote":
		//restart timer
		err := s.handle_requestVote(msg)
		if err != nil {
			fmt.Printf("Follower handle_requestVote Error: %v\n", err)
		} // Don't respond to voteReply messages if follower
	//Vote handle - Max
	default:
		fmt.Printf("Handle_follower default case: Message is %s\n", msg.Type)
	}
}

func (s *Sailor) handle_leader(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "get":
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
	case "set":
		s.handle_set(msg, state)
		//Sets handle - Max
	case "appendEntries":
		am := appendMessage{}
		getPayload(msg.Value, &am)
		val, err := handleAppendEntries(s, state, &am, msg.Source)
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

	case "appendReply":
		ar := appendReply{}
		getPayload(msg.Value, &ar)
		err := handleAppendReply(s, state, &ar, msg.Source)
		if err != nil {
			fmt.Printf("leader, append reply handle: %s", err)
		}
		//AppendReply handle - Joseph
	case "requestVote":
		//Vote
		if msg.Type == "requestVote" {
			err := s.handle_requestVote(msg)
			if err != nil {
				fmt.Printf("Leader handle_requestVote Error:%v\n", err)
			}
		}
		// Ignore vote replies if in leader state
	case "votReply":
	default:
		fmt.Printf("handle_leader default message is %s\n", msg.Type)
	}
}

func (s *Sailor) handle_candidate(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "appendEntries":
		//timereset <- false // Triggers timer thread to restart timer
		am := appendMessage{}
		getPayload(msg.Value, &am)
		val, err := handleAppendEntries(s, state, &am, msg.Source)
		rep := makeReply(s, &msg, "appendReply")
		rep.Value = makePayload(val)
		//rep.Error = err.Error()
		err = s.client.SendToPeers(rep, rep.Destination)
		if err != nil {
			fmt.Printf("candidate, append entries: %s", err)
		}
	case "requestVote":
		err := s.handle_requestVote(msg)
		if err != nil {
			fmt.Printf("Candidate handle_requestVote Error: %v\n", err)
		}
	case "voteReply":
		err := s.handle_voteReply(msg)
		if err != nil {
			fmt.Printf("Candidate handle_voteReply Error: %v\n", err)
		}
	case "set":
		err := s.setReject(&msg)
		if err != nil {
			fmt.Printf("candidate set: %v\n", err)
		}
	case "get":
		err := s.getAttempt(msg, state)
		if err != nil {
			fmt.Printf("Candidate getAttempt error: %v\n", err)
		}
	default:
		fmt.Printf("Default Handle_candidate message is %s\n", msg.Type)

		//VoteReply handle - Max
	}
}

func makeReply(s *Sailor, msg *messages.Message, typestr string) messages.Message {
	rep := messages.Message{}
	rep.Type = typestr
	rep.ID = msg.ID
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
	return ""
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
	s.timer.Reset(new_time())
	s.leaderId = ""
}
