package raft

import (
	"math/rand"
	"time"
)

func new_time() time.Duration {
	return time.Duration((rand.Intn(150) + 150)) * time.Millisecond
}

func (s *Sailor) Timer(TIMEOUT_SIGNAL chan bool, RESET chan bool) {
	fmt.Printf("%s Timer Started\n", s.client.NodeName)
	timer := time.NewTimer(new_time())
	for {
		if s.state == leader {
			select {
			case <-RESET:
				timer.Reset(time.Duration(50) * time.Millisecond)
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
				//fmt.Printf("Timeout Occured: %s, %s, %s\n", s.client.NodeName, s.state, s.currentTerm)
			}
		}
	}
}
