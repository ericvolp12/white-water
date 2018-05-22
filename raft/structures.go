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

type Sailor struct {
	// Contains filtered or unexported fields
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
