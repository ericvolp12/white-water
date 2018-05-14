package main

import (
	"encoding/json"
	"flag"
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

type arrayFlags []string

// Message is a Chistributed message struct
type Message struct {
	Type  string `json:"type"`
	ID    int    `json:"id"`
	Key   string `json:"key"`
	Error string `json:"error"`
	Value string `json:"value"`
}

func (i *arrayFlags) String() string {
	return "It's an array of peers"
}

func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	var peers arrayFlags

	// Flag initialization
	pubEndpoint := flag.String("pub-endpoint",
		"tcp://127.0.0.1:23310", "URI of the broker's ZMQ PUB socket.")
	routerEndpoint := flag.String("router-endpoint",
		"tcp://127.0.0.1:34411", "URI of the broket's ZMQ ROUTER socket.")
	nodeName := flag.String("node-name", "", "The name of a node")
	flag.Var(&peers, "peer", "A peer for a given node.")

	flag.Parse()

	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	requester.Connect(*routerEndpoint)

	sub, _ := zmq.NewSocket(zmq.PUB)
	defer sub.Close()
	sub.Connect(*pubEndpoint)

	sub.SetSubscribe(*nodeName)

	for {
		msg, err := sub.RecvMessageBytes(0)
		if err != nil {
			break
		}

		cMsg := &Message{}
		// Unwrap the message from JSON to Go
		json.Unmarshal(msg[1], cMsg)
		// Print the type of the message
		fmt.Printf("Type: %v", cMsg.Type)
	}
}
