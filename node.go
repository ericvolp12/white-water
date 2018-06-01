package main

import (
	"flag"
	"fmt"

	messages "github.com/ericvolp12/white-water/messages"
)

type arrayFlags []string

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

	debugFlag := flag.Bool("debug", false, "Whether or not to provide debugging output.")

	flag.Parse()

	if *debugFlag {
		fmt.Println("Node started with debugging...")
	}

	client := messages.CreateClient(*pubEndpoint, *routerEndpoint, *nodeName, peers)

	//s := raft.InitializeSailor(&client)

	//state := storage.InitializeState()
	client.HandleSingleHello()

}

func initializeChannels(client *messages.Client) (gets, sets, requestVote, appendEntry chan messages.Message, timereset, timeouts chan bool) {
	gets = make(chan messages.Message, 500)
	sets = make(chan messages.Message, 500)
	requestVote = make(chan messages.Message, 500)
	appendEntry = make(chan messages.Message, 500)
	timeouts = make(chan bool, 500)
	timereset = make(chan bool, 500)
	client.Subscribe("get", &gets) // CHECK TYPE
	client.Subscribe("set", &sets) // CHECK TYPE
	client.Subscribe("requestVote", &requestVote)
	client.Subscribe("voteReply", &requestVote)
	client.Subscribe("appendEntries", &appendEntry)
	client.Subscribe("appendReply", &appendEntry)
	return
}
