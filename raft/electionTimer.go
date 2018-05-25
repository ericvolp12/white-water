package raft

import(
    "fmt"
    "time"
    "math/rand"
)

func (s *Sailor) timer(TIMEOUT_SIGNAL *chan bool) {
    timeout := time.Duration((rand.Intn(150) + 150)) * time.Millisecond //Random timeout in MS
    last := s.lastMessageTime
    for {
        if (last != s.lastMessageTime) { // triggers timer restart
            timeout = time.Duration((rand.Intn(150) + 150)) * time.Millisecond
        }
        last = s.lastMessageTime
        if(time.Since(s.lastMessageTime) > timeout) {
            fmt.Printf("TIMEOUT")
            *TIMEOUT_SIGNAL <- true // Sending TIMEOUT signal to raft thread
        }
    }
}
