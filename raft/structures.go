package raft

import (
	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

type GetReply string

type entry struct {
	term  uint
	trans storage.Transaction
	votes uint
}

type leaderState struct {
	nextIndex  map[string]uint
	matchIndex map[string]uint
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
	lastCommitIndex uint
	currentTerm     uint
	votedFor        string
	numVotes        int
	volatile        *volatileState
	leader          *leaderState
}

type appendMessage struct { //type="appendEntries"
	Term         uint
	LeaderId     string
	PrevLogIndex uint
	PrevLogTerm  uint
	Entries      []entry
	LeaderCommit uint
}

type appendReply struct { //type="appendReply"
	Term       uint
	Success    bool
	MatchIndex uint
}

type requestVote struct { //type="requestVote"
	Term         uint
	CandidateId  string
	LastLogIndex uint
	LastLogTerm  uint
}

type reply struct { //type="voteReply"
	Term        uint
	VoteGranted bool
	Success     bool
}

func InitializeSailor(c *messages.Client) *Sailor {
	s := Sailor{}
	s.state = follower
	s.client = c
	s.currentTerm = 1
	s.volatile = &volatileState{}
	return &s
}
