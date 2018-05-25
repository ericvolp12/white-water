package raft

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry *chan messages.Message, timeouts *chan bool) {
	select {
	case msg := <-*timeouts:
		//timeouts message handle
		//Max
	case msg := <-*gets:
		//get message handle
		//Joseph
	case msg := <-*sets:
		//set message handler
		//Joseph
	case msg := <-*requestVote:
		//requestVote message handle
		//Max
	case msg := <-*appendEntry:
		//append message handler
		//Max

	}
}


// ON CONVERSION TO CANDIDATE:
func (s *Sailor) handle_timeout(msg messages.Message) {
    s.electionLock.Lock();
    s.electionLock.Unlock();
}
