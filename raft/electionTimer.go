package raft

import (
	"fmt"
	"math/rand"
	"time"
)

func new_time() time.Duration {
	return time.Duration((rand.Intn(150) + 150)) * time.Millisecond
}

func (s *Sailor) timer(TIMEOUT_SIGNAL chan bool) {
	last := s.lastMessageTime
	timer := time.NewTimer(new_time())
	for {
		if last != s.lastMessageTime {
			last = s.lastMessageTime
			timer.Reset(new_time())
		}
		<-timer.C
		TIMEOUT_SIGNAL <- true
		fmt.Printf("Timeout Occured\n")
	}
	/*
		last := s.lastMessageTime
		for {
			if last != s.lastMessageTime { // triggers timer restart
				timeout = new_timer //time.Duration((rand.Intn(150) + 150)) * time.Millisecond
			}
			last = s.lastMessageTime
			if time.Since(s.lastMessageTime) > timeout {
				fmt.Printf("TIMEOUT")
				*TIMEOUT_SIGNAL <- true // Sending TIMEOUT signal to raft thread
			}
		}
	*/
}
