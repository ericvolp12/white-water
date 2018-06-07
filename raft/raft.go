package raft

// raft.go contains the very core of the raft logic, along with some helper functions

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// MsgHandler is the main function of our Raft nodes.
// It takes a sailor struct and a pointer to the state machine.
// It will loop over all incoming messages to determine where to process them
// and shunts them off appropriately.
// It also handles election timeout processing.
func (s *Sailor) MsgHandler(state *storage.State, peerCount int) {
	s.timer = time.NewTimer(time.Second * time.Duration(peerCount)) // Make the first delay longer so other nodes can start
	for {
		// Choose the timeout if it exists, otherwise try and get a message from the network
		select {
		case <-s.timer.C:
			s.handle_timeout()
		default:
			msg := s.client.ReceiveMessage()
			if msg != nil { // We got a message, send it to the correct handler
				switch s.state {
				case leader:
					s.handle_leader(*msg, state)
				case follower:
					s.handle_follower(*msg, state)
				case candidate:
					s.handle_candidate(*msg, state)
				}
			} else { // We didn't get a message
				time.Sleep(time.Millisecond * 5) // Sleep to reduce CPU load
			}
		}
	}
}

// handle_follower is a helper function that contains the logic
// of what to do with a message when we are a follower.
// It takes a sailor, the state machine, and a message.
func (s *Sailor) handle_follower(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "get": // It's a get messagae
		err := s.getAttempt(msg, state)
		if err != nil {
			fmt.Printf("Candidate getAttempt error: %v\n", err)
		}
	case "set": // It's a set message
		err := s.setReject(&msg)
		if err != nil {
			fmt.Printf("Follower Reject error: %v\n", err)
		}
	case "appendEntries": // It's an appendEntires RPC
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
	case "requestVote": // It's a requestVote RPC
		err := s.handle_requestVote(msg)
		if err != nil {
			fmt.Printf("Follower handle_requestVote Error: %v\n", err)
		}
	default: // Not a recognize message type.
		fmt.Printf("Handle_follower default case: Message is %s\n", msg.Type)
	}
}

// handle_leader is a helper function that contains the logic
// of what to do with a message when we are a leader.
// It takes a sailor, the state machine, and a message.
func (s *Sailor) handle_leader(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "get": // It's a get request
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
			//handle error, we've never seen one
		}
	case "set": // It's a set request
		s.handle_set(msg, state)
	case "appendEntries": // It's an appendEntries RPC
		am := appendMessage{}
		getPayload(msg.Value, &am)
		val, err := handleAppendEntries(s, state, &am, msg.Source)
		if err != nil {
			fmt.Printf("leader, append entries handle: %s", err)
		}
		rep := makeReply(s, &msg, "appendReply")
		rep.Value = makePayload(val)
		err = s.client.SendToPeers(rep, rep.Destination)
		if err != nil {
			fmt.Printf("leader, append entries: %s", err)
		}

	case "appendReply": // It's a reply to an appendEntries RPC
		ar := appendReply{}
		getPayload(msg.Value, &ar)
		err := handleAppendReply(s, state, &ar, msg.Source)
		if err != nil {
			fmt.Printf("leader, append reply handle: %s", err)
		}
	case "requestVote": // It's a requestVote RPC
		if msg.Type == "requestVote" {
			err := s.handle_requestVote(msg)
			if err != nil {
				fmt.Printf("Leader handle_requestVote Error:%v\n", err)
			}
		}
	case "voteReply": // Ignore vote replies if in leader state
	default:
		fmt.Printf("handle_leader default message is %s\n", msg.Type)
	}
}

// handle_candidate is a helper function that contains the logic
// of what to do with a message when we are a candidate.
// It takes a sailor, the state machine, and a message.
func (s *Sailor) handle_candidate(msg messages.Message, state *storage.State) {
	switch msg.Type {
	case "appendEntries": // It's an appendEntries RPC
		am := appendMessage{}
		getPayload(msg.Value, &am)
		val, err := handleAppendEntries(s, state, &am, msg.Source)
		rep := makeReply(s, &msg, "appendReply")
		rep.Value = makePayload(val)
		err = s.client.SendToPeers(rep, rep.Destination)
		if err != nil {
			fmt.Printf("candidate, append entries: %s", err)
		}
	case "requestVote": // It's a requestVote RPC
		err := s.handle_requestVote(msg)
		if err != nil {
			fmt.Printf("Candidate handle_requestVote Error: %v\n", err)
		}
	case "voteReply": // It's a requestVote RPC reply
		err := s.handle_voteReply(msg)
		if err != nil {
			fmt.Printf("Candidate handle_voteReply Error: %v\n", err)
		}
	case "set": // It's a set request
		err := s.setReject(&msg)
		if err != nil {
			fmt.Printf("candidate set: %v\n", err)
		}
	case "get": // It's a get rerquest
		err := s.getAttempt(msg, state)
		if err != nil {
			fmt.Printf("Candidate getAttempt error: %v\n", err)
		}
	default:
		fmt.Printf("Default Handle_candidate message is %s\n", msg.Type)
	}
}

// makeReply is a helper function that constructs a partially
// completed reply from the original request sent
func makeReply(s *Sailor, msg *messages.Message, typestr string) messages.Message {
	rep := messages.Message{}
	rep.Type = typestr
	rep.ID = msg.ID
	rep.Destination = []string{msg.Source}
	rep.Source = s.client.NodeName
	return rep
}

// makePayload encodes different Raft message RPC structs into strings for ZMQ type messages
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

// getPayload decodes Zmq message.Value payload to whatever struct is passed into the func
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

// becomeFollower c\onverts Sailor into follower state (Normally if msg.Term > s.currentTerm)
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
