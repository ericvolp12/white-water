package raft

func handleGetRequest(string key, s *Sailor, st *storage.State) (string, error) {
	gt := GenerateTransaction(storage.GetOp, key, "")
	return st.ApplyTransaction(gt)
}
