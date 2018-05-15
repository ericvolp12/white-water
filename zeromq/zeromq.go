package zeromq

import (
	"encoding/json"
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

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

// Client defines a messenger client, they can send or receive messages
type Client struct {
	req      *zmq.Socket // Requester socket for the client
	sub      *zmq.Socket // Subscriber socket for the client
	nodeName string      // Name of this node
	peers    []string    // Names of peers of this node
}

// CreateClient defines a client based on input information
func CreateClient(pubEndpoint string, routerEndpoint string, nodeName string, peers []string) Client {

	// Initialize Requester
	requester, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		fmt.Printf("Error creating requester: %v\n", err)
	}
	fmt.Println("Requester client to connect to router...")
	// Connect requester to router
	requester.Connect(routerEndpoint)
	// Set identity to our node name so we can get messages
	err = requester.SetIdentity(nodeName)
	if err != nil {
		fmt.Printf("Error identifying as: %v, %v\n", nodeName, err)
	}
	// Initialize Subscriber
	sub, err := zmq.NewSocket(zmq.SUB)
	if err != nil {
		fmt.Printf("Error creating subscriber: %v\n", err)
	}
	fmt.Println("Subscriber client trying to connect to Pub Endpoint...")
	// Connect subscriber to publisher endpoint
	sub.Connect(pubEndpoint)
	// Linger set to 0 so we don't deadlock
	err = sub.SetLinger(0)
	if err != nil {
		fmt.Printf("Error lingering: %v\n", err)
	}
	// Set subscription topic to our node name
	err = sub.SetSubscribe(nodeName)
	if err != nil {
		fmt.Printf("Error subscribing to: %v, %v\n", nodeName, err)
	}
	// Create our client object
	client := Client{req: requester, sub: sub, nodeName: nodeName, peers: peers}

	// Send it back
	return client
}

// DeleteClient closes open connections and deletes our client
func DeleteClient(c *Client) {
	// Close the requester
	c.req.Close()
	// Close the subscriber
	c.sub.Close()
}

// ReceiveMessages starts a duty loop to handle ZeroMQ Messages on our subscriber
func ReceiveMessages(c *Client) {
	for {
		msg, err := c.sub.RecvMessage(0)
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
			// Send a hello response
			err = sendMessage(c, Message{Type: "helloResponse", Source: c.nodeName})
			if err != nil {
				break
			}

		}
	}
}

// sendMessage sends a message on a client's requester
func sendMessage(c *Client, msg Message) error {
	// Marshal message object into json
	formattedMsg, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	// Send message via requester
	sent, err := c.req.SendMessage(formattedMsg)
	if err != nil {
		return fmt.Errorf("Error sending message: %v", err)
	}
	if sent <= 0 {
		return fmt.Errorf("Error sending message: zero bytes sent")
	}
	return nil
}

// Broadcast sends a message to all of a client's peers
func Broadcast(c *Client, msg Message) error {
	msg.Destination = c.peers
	fmt.Printf("Broadcasting message to all peers...\n")
	return sendMessage(c, msg)
}

// SendToPeer sends to a specific peer
func SendToPeer(c *Client, msg Message, peer string) error {
	msg.Destination = []string{peer}
	fmt.Printf("Sending message one peer...\n")
	return sendMessage(c, msg)
}

// SendToPeers sends to a set of peers
func SendToPeers(c *Client, msg Message, peers []string) error {
	msg.Destination = peers
	fmt.Printf("Sending message to specific peers...\n")
	return sendMessage(c, msg)
}

// SendToBroker sends a message to the broker and no one else
func SendToBroker(c *Client, msg Message) error {
	msg.Destination = []string{}
	fmt.Printf("Sending message to broker...\n")
	return sendMessage(c, msg)
}
