package raft

import (
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// Generats a set transaction and applies it to the raft log
func (s *Sailor) handle_set(msg messages.Message, state *storage.State) {
	trans := storage.GenerateTransaction(storage.SetOp, msg.Key, msg.Value)
	newEntry := entry{term: s.currentTerm, trans: trans, votes: 1}
	s.log = append(s.log, newEntry)
}

// Checks for all commited transactions in an appendReply, majority commited get a broker reply
func (s *Sailor) handle_commit(lowCommit uint, upperCommit uint, state *storage.State) error {
	majority := uint((len(s.client.Peers) + 1) / 2)
	for i := lowCommit - 1; i <= upperCommit-1; i++ {
		s.log[i].votes += 1 // Increments the number of commits
		if s.log[i].votes == majority {
			_, err := state.ApplyTransaction(s.log[i].trans)
			if err != nil {
				fmt.Printf("Handle Commit ApplyTrans error: %v\n", err)
				return err
			}
			zmqMsg := messages.Message{} //TODO confirm type string
			zmqMsg.Type = "setResponse"
			zmqMsg.Source = s.client.NodeName
			zmqMsg.Key = s.log[i].trans.Key
			zmqMsg.Value = s.log[i].trans.Value + " Write Successful"
			err = s.client.SendToBroker(zmqMsg)
			if err != nil {
				fmt.Printf("Handle commit SendToBroker error:%v\n", err)
				return err
			}
		}
	}
	return nil
}

func (s *Sailor) handle_prepare(PrepLower uint, PrepUpper uint) error {

	for i := PrepLower; i <= PrepUpper; i++ {
		if i <= s.volatile.commitIndex {
			continue
		}
		s.log[i-1].votes += 1
		if s.log[i-1].votes > uint((len(s.client.Peers)+1)/2) && s.log[i-1].term == s.currentTerm {
			for j := s.volatile.commitIndex + 1; j <= i; j++ {
				s.log[j-1].votes = 1
			}
			s.volatile.commitIndex = i
		}
	}
	return nil
}
