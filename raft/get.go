package raft

import storage "github.com/ericvolp12/white-water/storage"

func handleGetRequest(key string, s *Sailor, st *storage.State) (string, error) {
	gt := storage.GenerateTransaction(storage.GetOp, key, "")
	return st.ApplyTransaction(gt)
}
