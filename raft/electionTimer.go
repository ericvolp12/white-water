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
	timer := time.NewTimer(new_time())
	for {
		if s.state == leader {
			timer.Reset(time.Duration(50) * time.Millisecond)
			<-timer.C
			TIMEOUT_SIGNAL <- true
		} else {
			select {
			case <-TIMEOUT_SIGNAL:
				timer.Reset(new_time())
			case <-timer.C:
				TIMEOUT_SIGNAL <- true
				fmt.Printf("Timeout Occured\n")
			}
		}
	}
}
