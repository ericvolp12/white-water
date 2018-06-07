package raft

// get.go
// Functions in this file are used to handle get request from a client
// for any Raft state a node may be in

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// Given a key from a get request message, this function calls the storage
// module function to apply a get transaction. This call to the storage will
// either return the Value associated with that key in the node's storage
// or it returns an error message if no such key exists
// This function is used only by a leader to provide the most-up-to-date
// information to the client.
// Used in raft.go: handle_leader()
func handleGetRequest(key string, s *Sailor, st *storage.State) (string, error) {
	gt := storage.GenerateTransaction(storage.GetOp, key, "")
	return st.ApplyTransaction(gt)
}

// Given a message of type "get", this function attempts to apply a get
// transaction to the storage module for the indicated Key given in the message.
// No matter what the storage module returns (an error if the key is not in
// storage or a value associated with that key), this function places that
// info, along with the current leader's Id string, int a response message's
// error field. This function is only called when a non-leader recieves a
// get request, in order to indicate that it is not the leader (and therefore
// is not guaranteed to have the most updated value for a given key), but may
// still give the most recent value the node has stored.
// This function sends said message back to the client, and returns an error
// if an error was encountered while sending the message.
// Used in raft.go:handle_follower() and handle_candidate()
func (s *Sailor) getAttempt(msg messages.Message, st *storage.State) error {
	gt := storage.GenerateTransaction(storage.GetOp, msg.Key, "")
	item, err := st.ApplyTransaction(gt)

	reply := makeReply(s, &msg, "getResponse")
	if err != nil {
		if s.state == candidate {
			reply.Error = err.Error() + " : In Election Cycle"
		} else {
			reply.Error = "|Src: " + s.client.NodeName + "|" + err.Error() + " : Current Leader is " + s.leaderId
		}
	} else {
		reply.Error = msg.Key + "= " + item + " : Current Leader is " + s.leaderId
	}
	return s.client.SendToBroker(reply)
}
