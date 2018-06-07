package raft

// timer.go
// Utility functions for reseting the timer used in raft elections, timeouts,
// and heartbeats

import (
	"math/rand"
	"time"
)

// Returns a new random timer duration in milliseconds ranging
// from 150 ms to 300 ms as recommended by the authors of the Raft algorithm
// to be used for follower and candidate timout durations.
// Used in requestVote.go:handle_timeout(), handle_requestVote,
// and append.go:handleAppendEntries()
func new_time() time.Duration {
	return time.Duration((rand.Intn(300) + 150)) * time.Millisecond
}

// Returns a new timer duration of 50 milliseconds used for triggering a
// leader's heartbeat messages (appendEntries type)
// used in requestVote.go: handle_timeout(), and handle_voteReply
func leaderReset() time.Duration {
	return time.Duration(time.Duration(50) * time.Millisecond)

}
