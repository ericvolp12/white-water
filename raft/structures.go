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
    term  uint
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
    client          *messages.Client
	state           position
	log             []entry
	currentTerm     uint
	votedFor        int
    numVotes        int
	volatile        *volatileState
	leader          *leaderState
	lastMessageTime time.Time
}

type appendMessage struct { //type="appendEntries"
	term         uint
	leaderId     int
	prevLogIndex uint
	prevLogTerm  uint
	entries      []entry
	leaderCommit uint
}

type requestVote struct { //type="requestVote"
	term         uint
	candidateId  uint
	lastLogIndex uint
	lastLogTerm  uint
}

type reply struct {
    term uint
    voteGranted bool
    success bool
}
