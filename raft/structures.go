package raft

import "time"

type GetReply string

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
	Term         uint
	LeaderId     string
	PrevLogIndex uint
	PrevLogTerm  uint
	Entries      []entry
	LeaderCommit uint
}

type requestVote struct { //type="requestVote"
	Term         uint
	CandidateId  string
	LastLogIndex uint
	LastLogTerm  uint
}

type reply struct {
	Term        uint
	VoteGranted bool
	Success     bool
}
