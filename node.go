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
	Type        string   `json:"type"`
	ID          int      `json:"id"`
	Destination []string `json:"destination"`
	Key         string   `json:"key"`
	Error       string   `json:"error"`
	Source      string   `json:"source"`
	Value       string   `json:"value"`
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

	requester, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		fmt.Printf("Error creating requester: %v\n", err)
	}
	defer requester.Close()
	fmt.Println("Trying to connect to router...")
	requester.Connect(*routerEndpoint)

	err = requester.SetIdentity(*nodeName)
	if err != nil {
		fmt.Printf("Error identifying as: %v, %v\n", *nodeName, err)
	}

	sub, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		fmt.Printf("Error creating subscriber: %v\n", err)
	}
	defer sub.Close()
	fmt.Println("Trying to connect to Pub Endpoint...")
	sub.Connect(*pubEndpoint)

	topic := *nodeName + " "

	err = sub.SetLinger(0)
	if err != nil {
		fmt.Printf("Error lingering: %v\n", err)
	}

	err = sub.SetSubscribe(*nodeName)
	if err != nil {
		fmt.Printf("Error subscribing to: %v, %v\n", topic, err)
	}

	for {
		msg, err := sub.RecvMessage(0)
		if err != nil {
			break
		}
		fmt.Println("Message received:")

		cMsg := &Message{}
		// Unwrap the message from JSON to Go
		// Index 2 should be the json payload
		json.Unmarshal([]byte(msg[2]), cMsg)
		// Print the type of the message
		fmt.Printf("\tType: %v\n", cMsg.Type)
		fmt.Printf("\tFull message: %v\n", msg[2])
		if cMsg.Type == "hello" {
			reply, err := json.Marshal(Message{Type: "helloResponse", Source: *nodeName})
			if err != nil {
				break
			}
			sent, err := requester.SendMessage(reply)
			if sent <= 0 || err != nil {
				break
			}

		}
	}
}
