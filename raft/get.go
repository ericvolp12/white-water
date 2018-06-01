package raft

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

func handleGetRequest(key string, s *Sailor, st *storage.State) (string, error) {
	gt := storage.GenerateTransaction(storage.GetOp, key, "")
	return st.ApplyTransaction(gt)
}

func (s *Sailor) getAttempt(msg messages.Message, st *storage.State) error {
	gt := storage.GenerateTransaction(storage.GetOp, msg.Key, "")
	item, err := st.ApplyTransaction(gt)

	reply := makeReply(s, &msg, "getResponse")
	if err != nil {
		reply.Error = err.Error() + " : Current Leader is " + s.leaderId
	} else {
		reply.Error = msg.Key + "= " + item + " : Current Leader is " + s.leaderId
	}
	return s.client.SendToBroker(reply)

}
