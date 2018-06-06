package raft

// structures.go holds all of the structure data for the Raft implementation.
// The structures hold all the data for the node structs and the RPC messages.

import (
	"time"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
)

// The entry struct holds all the information for a single entry in the log.
type entry struct {
	Term  uint                // The term when the log entry was proposed
	Trans storage.Transaction // The transaction to apply to the state machine
	votes uint                // Some scratch space for the leader to count votes
	Id    int                 // The ID of the messages that proposed it, needed to craft the set-reply
}

// The leaderState struct holds all the information that the leader needs to keep track of
type leaderState struct {
	nextIndex  map[string]uint // A map from the node-name to the nextIndex, as described in Raft
	matchIndex map[string]uint // A map from the node-name to the matchIndex, as described in Raft
}

// The volatileState struct holds all the information that doesn't need
// to be kept in persistant storage, as described in the Raft paper
type volatileState struct {
	commitIndex uint
	lastApplied uint
}

// position is a type alias to hold our current position, leader, candidate, or follower
type position int

const (
	follower  position = iota // The follower constant
	candidate                 // The candidate constant
	leader                    // The leader constant
)

// The sailor struct holds everything needed in the raft implementation
type Sailor struct {
	// Contains filtered or unexported fields
	client          *messages.Client // The network client
	state           position         // Our current position as a leader, candidate, or follower
	log             []entry          // The log, made of log entries
	lastCommitIndex uint             // The log index of our last committed value
	currentTerm     uint             // The current Raft term
	votedFor        string           // The name of the last node we voted for
	numVotes        int              // The number of votes we have as a candidate
	volatile        *volatileState   // The volatile data that doesn't need persistance
	leader          *leaderState     // The data we need to keep track of as leader
	leaderId        string           // The ID of the last known leader
	timer           *time.Timer      // The timer we use to signal when it is time for an election
}

// The appendMessage struct holds the data needed for an appendEntries RPC
type appendMessage struct { //type="appendEntries"
	Term         uint    // The current term
	LeaderId     string  // The current leader ID
	PrevLogIndex uint    // The log index before the ones we're sending
	PrevLogTerm  uint    // The term of the log index before the ones we're sending
	Entries      []entry // The new data to add to the log
	LeaderCommit uint    // The index that the leader has committed to
}

// The appendReply struct holds the data needed for a reply to an appendEntries RPC
type appendReply struct { //type="appendReply"
	Term      uint // The term of the follower
	Success   bool // Whether or not the call was successful
	PrepLower uint // The last log entry the follower had upon recieving the RPC
	PrepUpper uint // The follower's last long entry after the RPC
	ComLower  uint // The follower's commit index upon recieving the RPC
	ComUpper  uint // The follower's commit index after the RPC
}

// The requestVote struct holds the data needed for a requestVote RPC
type requestVote struct { //type="requestVote"
	Term         uint   // The term of the candidate
	CandidateId  string // The ID of the candidate
	LastLogIndex uint   // The index of the last entry in the candidate's log
	LastLogTerm  uint   // The term of the last entry in the candidate's log
}

// The reply struct holds the data needed for a reply to a requestVote RPC
type reply struct { //type="voteReply"
	Term        uint // The term of the voter
	VoteGranted bool // Whether or not the candidate got the vote
	Success     bool // Whether or not the call was successful
}

// InitializeSailor takes a network client and returns a pointer
// to a newly constructed Sailor struct.
func InitializeSailor(c *messages.Client) *Sailor {
	s := Sailor{}
	s.state = follower
	s.client = c
	s.currentTerm = 1
	s.volatile = &volatileState{}
	return &s
}
