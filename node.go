package main

import (
	"flag"
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
	raft "github.com/ericvolp12/white-water/raft"
	storage "github.com/ericvolp12/white-water/storage"
)

type arrayFlags []string

// Just for debugging the peers object
func (i *arrayFlags) String() string {
	return "It's an array of peers"
}

// Used to set multiple peers in the CLI
func (i *arrayFlags) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// Initialize node's ZMQ messaging and White Water MsgHandler
func main() {
	var peers arrayFlags

	// Flag initialization
	pubEndpoint := flag.String("pub-endpoint",
		"tcp://127.0.0.1:23310", "URI of the broker's ZMQ PUB socket.")
	routerEndpoint := flag.String("router-endpoint",
		"tcp://127.0.0.1:34411", "URI of the broket's ZMQ ROUTER socket.")
	nodeName := flag.String("node-name", "", "The name of a node")
	flag.Var(&peers, "peer", "A peer for a given node.")

	debugFlag := flag.Bool("debug", false, "Whether or not to provide debugging output.")

	flag.Parse()

	if *debugFlag {
		fmt.Println("Node started with debugging...")
	}

	client := messages.CreateClient(*pubEndpoint, *routerEndpoint, *nodeName, peers)

	s := raft.InitializeSailor(&client)

	state := storage.InitializeState()
	client.HandleSingleHello()
	s.MsgHandler(&state, len(client.Peers))
}
