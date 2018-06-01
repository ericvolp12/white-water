package raft

import (
	"fmt"
	"math/rand"
	"time"
)

func new_time() time.Duration {
	return time.Duration((rand.Intn(1500) + 1500)) * time.Millisecond
}

func (s *Sailor) Timer(TIMEOUT_SIGNAL chan bool, RESET chan bool) {
	fmt.Printf("%s Timer Started\n", s.client.NodeName)
	timer := time.NewTimer(new_time())
	for {
		if s.state == leader {
			select {
			case <-RESET:
				timer.Reset(time.Duration(500) * time.Millisecond)
				//fmt.Printf("Timer Reset Leader: %s\n", s.client.NodeName)
			case <-timer.C:
				TIMEOUT_SIGNAL <- true
				//fmt.Printf("HEARTBEAT Occured: %s, %s\n", s.client.NodeName, s.state)
			}
		} else {
			select {
			case <-RESET:
				timer.Reset(new_time())
				//fmt.Printf("Timer Reset: %s\n", s.client.NodeName)
			case <-timer.C:
				TIMEOUT_SIGNAL <- true
				fmt.Printf("	TIMEOUT OCCURED: %s, %s, %s\n", s.client.NodeName, s.state, s.currentTerm)
			}
		}
	}
}
