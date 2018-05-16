package main

import (
	"flag"
	"sync"

	messages "github.com/ericvolp12/white-water/messages"
	storage "github.com/ericvolp12/white-water/storage"
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

	flag.Parse()

	client := messages.CreateClient(*pubEndpoint, *routerEndpoint, *nodeName, peers)

	wg := sync.WaitGroup{}

	go client.ReceiveMessages()
	wg.Add(1)

	go storage.Initialize(&client)
	wg.Add(1)

	wg.Wait()

}
