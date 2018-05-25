package raft

import (
	"sync"
	"time"
)

type operation int

const (
	getOp operation = iota
	setOp
	deleteOp
)

type entry struct {
	op    operation
	key   string
	value string
}

type leaderState struct {
	nextIndex  []uint
	matchIndex []uint
}

type volatileState struct {
	commitIndex uint
	lastApplied uint
}

type position int

const (
    follower position = iota
    candidate
    leader
)

type Sailor struct {
	// Contains filtered or unexported fields
    state  position
	log             []entry
	currentTerm     uint
	votedFor        int
	volatile        *volatileState
	leader          *leaderState
	lastMessageTime time.Time
	electionLock    sync.RWMutex
}

type appendMessage struct {
	term         uint
	leaderId     int
	prevLogIndex uint
	prevLogTerm  uint
	entries      []entry
	leaderCommit uint
}

type requestVote struct {
    term    uint
    candidateId uint
    lastLogIndex    uint
    lastLogTerm     uint
    votGranted   bool
}
