package storage

import (
	"errors"
	"fmt"
	"time"
)

type Operation int

const (
	GetOp Operation = iota
	SetOp
	DeleteOp
)

// ValueStamp is a struct that holds a value paired with a timestamp
type ValueStamp struct {
	Value   string    // The value being stored
	Updated time.Time // The time it was last updated
}

// Transaction stores a transaction relating the type of transaction made, the key, and the value stamp
type Transaction struct {
	TType Operation // Type of transaction, either a get or a set
	Key   string    // Key we are getting ot setting
	Value string    // A Value, if needed for the transaction
}

// State contains the transaction log and the state map for indexing into
type State struct {
	Map map[string]ValueStamp // The map relating keys to values
}

// getLocal gets a string from the local state without the timestamp
func (s *State) getLocal(key string) (string, error) {
	val := s.Map[key]
	if val.Updated.IsZero() {
		return "", fmt.Errorf("No such key: %v", key)
	}
	return val.Value, nil
}

// getLocalWithStamp gets a string from the local state with timestamp
func (s *State) getLocalWithStamp(key string) (ValueStamp, error) {
	val := s.Map[key]
	if val.Updated.IsZero() {
		return ValueStamp{}, fmt.Errorf("No such key: %v", key)
	}
	return val, nil
}

// setLocal takes a key and value and creates a new ValueStamp and commits it
// to the local state
func (s *State) setLocal(key string, value string) error {
	val := ValueStamp{Value: value, Updated: time.Now()}
	s.Map[key] = val
	return nil
}

// GetWithStamp gets a ValueStamp from the cluster
func (s *State) GetWithStamp(key string) (ValueStamp, error) {
	return s.getLocalWithStamp(key)
}

// Get gets a string from the local machine without the timestamp
func (s *State) Get(key string) (string, error) {
	return s.getLocal(key)
}

// Set takes a key and value and creates a new ValueStamp and commits it
func (s *State) Set(key string, value string) error {
	return s.setLocal(key, value)
}

// InitializeState starts up the message listeners for the interface
func InitializeState() State {
	state := State{}

	state.Map = make(map[string]ValueStamp)
	return state
}

// ApplyTransaction takes a transaction struct and
// applies the described effect to the state machine
func (s *State) ApplyTransaction(t Transaction) (string, error) {
	switch t.TType {
	case GetOp:
		return s.getLocal(t.Key)
	case SetOp:
		return "", s.setLocal(t.Key, t.Value)
	default:
		return "", errors.New("Invalid Transaction!")
	}
}

// GenerateTransaction creates a transaction struct
func GenerateTransaction(opType Operation, key string, value string) Transaction {
	t := Transaction{opType, key, value}
	return t
}
