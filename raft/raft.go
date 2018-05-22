package raft

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry *chan messages.Message, timeouts *chan bool) {
	select {
	case msg := <-gets:
		//get message handle
	case msg := <-sets:
		//set message handler
	case msg := <-requestVote:
		//requestVote message handle
	case msg := <-appendEntry:
		//append message handler
	case msg := <-timeouts:
		//timeouts message handle

	}
}
