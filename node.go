package main

import (
	"flag"
	"fmt"
	"sync"

	zeromq "github.com/ericvolp12/white-water/zeromq"
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

	client := zeromq.CreateClient(*pubEndpoint, *routerEndpoint, *nodeName, peers)

	helloIncoming := make(chan zeromq.Message)

	client.Subscribe("hello", &helloIncoming)

	wg := sync.WaitGroup{}

	go client.ReceiveMessages()

	wg.Add(1)

	for range helloIncoming {
		fmt.Printf("Hello Handler Firing...\n")
		err := client.SendToBroker(zeromq.Message{Type: "helloResponse", Source: *nodeName})
		if err != nil {
			break
		}
	}

	wg.Wait()

}
