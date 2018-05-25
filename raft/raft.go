package raft

func (s *Sailor) MsgHandler(gets, sets, requestVote, appendEntry *chan messages.Message, timeouts *chan bool) {
	for {
		select {
		case msg := <-*timeouts:
			//timeouts message handle
			//Max
		default:
			select {
			case s.state == follower:
				select {
				//sets - J
				//gets - J
				//append - J
				//requestVote - M
				}
			case s.state == candidate:
				select {
				//requestVote - M
				//append -J
				}
			case s.state == leader:
				select {
				//gets -J
				//sets -J
				//append -J
				//requestVote -M
				//Propose-value?
				}
			}
		}
	}
}

// ON CONVERSION TO CANDIDATE:
func (s *Sailor) handle_timeout(msg messages.Message) {
}
