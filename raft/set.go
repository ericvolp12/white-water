package raft

import (
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// Generats a set transaction and applies it to the raft log
func (s *Sailor) handle_set(msg messages.Message, state *storage.State) {
	trans := storage.GenerateTransaction(storage.SetOp, msg.Key, msg.Value)
	newEntry := entry{Term: s.currentTerm, Trans: trans, votes: 1, Id: msg.ID}
	s.log = append(s.log, newEntry)
	fmt.Printf("		*************** LEADER SET%+v\n", newEntry)
}

func (s *Sailor) setReject(msg *messages.Message) error {
	rej := makeReply(s, msg, "setResponse")
	if s.state != candidate {
		rej.Error = "|Src:" + s.client.NodeName + " | Current Leader is " + s.leaderId
	} else {
		rej.Error = "|Src:" + s.client.NodeName + " | Election in progress"
	}
	rej.Key = msg.Key
	return s.client.SendToBroker(rej)
}

// Checks for all commited transactions in an appendReply, majority commited get a broker reply
func (s *Sailor) handle_commit(lowCommit uint, upperCommit uint, state *storage.State) error {
	majority := uint((len(s.client.Peers) + 1) / 2)
	for i := int(lowCommit); i <= int(upperCommit)-1; i++ {
		s.log[i].votes += 1 // Increments the number of commits
		if s.log[i].votes == majority && s.log[i].Id != -1 {
			_, err := state.ApplyTransaction(s.log[i].Trans)
			if err != nil {
				fmt.Printf("Handle Commit ApplyTrans error: %v\n", err)
				return err
			}
			zmqMsg := messages.Message{}
			zmqMsg.Type = "setResponse"
			zmqMsg.Source = s.client.NodeName
			zmqMsg.Key = s.log[i].Trans.Key
			zmqMsg.ID = s.log[i].Id
			zmqMsg.Value = s.log[i].Trans.Value
			fmt.Printf("HANDLED COMMIT:		")
			err = s.client.SendToBroker(zmqMsg)
			if err != nil {
				fmt.Printf("Handle commit SendToBroker error:%v\n", err)
				return err
			}
		}
	}
	return nil
}

// Checks all prepared indices in an appendReply, majority prepared moves up the commit index
func (s *Sailor) handle_prepare(PrepLower uint, PrepUpper uint) error {

	for i := PrepLower; i <= PrepUpper; i++ {
		if i <= s.volatile.commitIndex {
			continue
		}
		if s.log[i-1].Term == s.currentTerm {
			s.log[i-1].votes += 1
		}
		if s.log[i-1].votes > uint((len(s.client.Peers)+1)/2) && s.log[i-1].Term == s.currentTerm {
			for j := s.volatile.commitIndex + 1; j <= i; j++ {
				s.log[j-1].votes = 1
			}
			s.volatile.commitIndex = i
		}
	}
	return nil
}
