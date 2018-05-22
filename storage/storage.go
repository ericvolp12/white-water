package storage

import (
	"fmt"
	"sync"
	"time"

	messages "github.com/ericvolp12/white-water/messages"
)

// ValueStamp is a struct that holds a value paired with a timestamp
type ValueStamp struct {
	Value   string    // The value being stored
	Updated time.Time // The time it was last updated
}

// Transaction stores a transaction relating the type of transaction made, the key, and the value stamp
type Transaction struct {
	TType string     // Type of transaction, either a get or a set
	Key   string     // Key we are getting ot setting
	Stamp ValueStamp // A Value and Timestamp pairing
}

// State contains the transaction log and the state map for indexing into
type State struct {
	Log []Transaction         // The transaction log, ordered by time
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

// Get gets a string from the cluster without the timestamp
func (s *State) Get(key string) (string, error) {
	return s.getLocal(key)
}

// Set takes a key and value and creates a new ValueStamp and commits it
func (s *State) Set(key string, value string) error {
	return s.setLocal(key, value)
}

// Initialize starts up the message listeners for the interface
func Initialize(client *messages.Client) {
	state := State{}

	state.Map = make(map[string]ValueStamp)

	wg := sync.WaitGroup{}

	go state.getHandler(client)
	wg.Add(1)

	go state.setHandler(client)
	wg.Add(1)

	wg.Wait()
}

// getHandler handles get messages...
func (s *State) getHandler(client *messages.Client) {
	getIncoming := make(chan messages.Message, 500)

	client.Subscribe("get", &getIncoming)

	for msg := range getIncoming {
		fmt.Printf("Get Handler Firing...\n")
		// Attempt to get value
		val, err := s.Get(msg.Key)

		// If there is no value stored for key
		if err != nil {
			// Send an error message to the broker
			err = client.SendToBroker(messages.Message{
				Type:   "getResponse",
				Source: client.NodeName,
				ID:     msg.ID,
				Error:  err.Error(),
			})
			if err != nil {
				break
			}
		}
		// If we found a value, send a response to the broker
		err = client.SendToBroker(messages.Message{
			Type:   "getResponse",
			Source: client.NodeName,
			ID:     msg.ID,
			Key:    msg.Key,
			Value:  val,
		})
		if err != nil {
			break
		}
	}
}

// setHandler handles set messages...
func (s *State) setHandler(client *messages.Client) {
	setIncoming := make(chan messages.Message, 500)

	client.Subscribe("set", &setIncoming)

	for msg := range setIncoming {
		fmt.Printf("Set Handler Firing...\n")
		// Attempt to set value
		err := s.Set(msg.Key, msg.Value)

		// If something went horribly wrong
		if err != nil {
			// Send an error message to the broker
			err = client.SendToBroker(messages.Message{
				Type:  "setResponse",
				ID:    msg.ID,
				Error: err.Error(),
			})
			if err != nil {
				break
			}
		}
		// If we set a value, send a response to the broker
		err = client.SendToBroker(messages.Message{
			Type:   "setResponse",
			Source: client.NodeName,
			ID:     msg.ID,
			Key:    msg.Key,
			Value:  msg.Value,
		})
		if err != nil {
			break
		}
	}
}
