package messages

import (
	"encoding/json"
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

// Message is a Chistributed message struct
type Message struct {
	Type        string   `json:"type"`
	ID          int      `json:"id,omitempty"`
	Destination []string `json:"destination,omitempty"`
	Key         string   `json:"key,omitempty"`
	Error       string   `json:"error,omitempty"`
	Source      string   `json:"source,omitempty"`
	Value       string   `json:"value,omitempty"`
}

// dumbMessage is a Chistributed message struct where you can only send to one dest
type dumbMessage struct {
	Type        string `json:"type"`
	ID          int    `json:"id"`
	Destination string `json:"destination"`
	Key         string `json:"key"`
	Error       string `json:"error,omitempty"`
	Source      string `json:"source"`
	Value       string `json:"value,omitempty"`
}

// Client defines a messenger client, they can send or receive messages
type Client struct {
	req      *zmq.Socket // Requester socket for the client
	sub      *zmq.Socket // Subscriber socket for the client
	NodeName string      // Name of this node
	Peers    []string    // Names of peers of this node
	buffer   chan Message
}

// CreateClient defines a client based on input information
func CreateClient(pubEndpoint string, routerEndpoint string, nodeName string, peers []string) Client {

	// Initialize Requester
	requester, err := zmq.NewSocket(zmq.REQ)
	if err != nil {
		fmt.Printf("Error creating requester: %v\n", err)
	}
	//	fmt.Println("Requester client to connect to router...")
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
	//fmt.Println("Subscriber client trying to connect to Pub Endpoint...")
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
	client := Client{req: requester, sub: sub, NodeName: nodeName, Peers: peers}
	client.buffer = make(chan Message, 500)
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

// ReceiveMessage pops a message off the queue and hands it to the
// calling process
func (c *Client) ReceiveMessage() *Message {
	// If we have a message in the buffer, send it back
	select {
	case msg := <-c.buffer:
		return &msg
	default:
	}
	// Actually get the message from ZMQ
	msg, err := c.sub.RecvMessage(zmq.DONTWAIT)
	if err != nil {
		return nil
		//log.Fatal("Error receiveing message: %v\n", err)
	}

	dMsg := dumbMessage{}
	// Unwrap the message from JSON to Go
	// Index 2 should be the json payload
	json.Unmarshal([]byte(msg[2]), &dMsg)

	// Convert to standard message from a DumbMessage
	cMsg := Message{}
	cMsg.Destination = []string{dMsg.Destination}
	cMsg.Type = dMsg.Type
	cMsg.Source = dMsg.Source
	cMsg.Value = dMsg.Value
	cMsg.Key = dMsg.Key
	cMsg.Value = dMsg.Value
	cMsg.Error = dMsg.Error
	cMsg.ID = dMsg.ID

	return &cMsg
}

// sendMessage sends a message on a client's requester
func (c *Client) sendMessage(msg Message) error {
	// Marshal message object into json
	// Because the docs are wrong we gotta send many messages
	if len(msg.Destination) > 0 {
		//fmt.Printf("Sending dumb messages, Borja be damned!\n")
		for _, dest := range msg.Destination {
			dMsg := dumbMessage{}
			dMsg.Destination = dest
			dMsg.Type = msg.Type
			dMsg.Source = msg.Source
			dMsg.Value = msg.Value
			dMsg.Key = msg.Key
			dMsg.Value = msg.Value
			dMsg.Error = msg.Error
			dMsg.ID = msg.ID
			dMsg.print()

			formattedMsg, err := json.Marshal(dMsg)
			if err != nil {
				fmt.Printf("Error marshalling message: %v\n", err)
				return err
			}
			// Send message via requester
			sent, err := c.req.SendMessage(formattedMsg)
			if err != nil {
				fmt.Printf("Error sending message: %v\n", err)
				return fmt.Errorf("Error sending message: %v", err)
			}
			if sent <= 0 {
				fmt.Printf("Error sending message: zero bytes sent\n")
				return fmt.Errorf("Error sending message: zero bytes sent")
			}
			// Wait for an ack response
			_, err = c.req.RecvMessage(0)
			if err != nil {
				fmt.Printf("Error in sendMessage (%s, %+v)\n", c.NodeName, err)
			}
			// Print out the ack
			//fmt.Printf("Ack: %v\n", ack)
		}
	} else {
		dMsg := dumbMessage{}
		dMsg.Destination = ""
		dMsg.Type = msg.Type
		dMsg.Source = msg.Source
		dMsg.Value = msg.Value
		dMsg.Key = msg.Key
		dMsg.Value = msg.Value
		dMsg.Error = msg.Error
		dMsg.ID = msg.ID
		dMsg.print()

		formattedMsg, err := json.Marshal(dMsg)
		if err != nil {
			fmt.Printf("Error marshalling message: %v\n", err)
			return err
		}
		// Send message via requester
		sent, err := c.req.SendMessage(formattedMsg)
		if err != nil {
			fmt.Printf("Error sending message: %v\n", err)
			return fmt.Errorf("Error sending message: %v", err)
		}
		if sent <= 0 {
			fmt.Printf("Error sending message: zero bytes sent\n")
			return fmt.Errorf("Error sending message: zero bytes sent")
		}
		// Wait for an ack response
		_, err = c.req.RecvMessage(0)
		// Print out the ack
		//fmt.Printf("Ack: %v\n", ack)
	}
	return nil
}

// Broadcast sends a message to all of a client's peers
func (c *Client) Broadcast(msg Message) error {
	msg.Destination = c.Peers
	//fmt.Printf("Broadcasting message to all peers... %s\n", c.NodeName)
	return c.sendMessage(msg)
}

// SendToPeer sends to a specific peer
func (c *Client) SendToPeer(msg Message, peer string) error {
	msg.Destination = []string{peer}
	//fmt.Printf("Sending message one peer... %+v\n", msg)
	return c.sendMessage(msg)
}

// SendToPeers sends to a set of peers
func (c *Client) SendToPeers(msg Message, peers []string) error {
	msg.Destination = peers
	//fmt.Printf("Sending message to specific peers...\n")
	return c.sendMessage(msg)
}

// SendToBroker sends a message to the broker and no one else
func (c *Client) SendToBroker(msg Message) error {
	msg.Destination = []string{}
	//fmt.Printf("Sending message to broker...\n")
	msg.Print()
	return c.sendMessage(msg)
}

// Print prints a message
func (m *Message) Print() {
	_, err := json.Marshal(*m)
	if err != nil {
		fmt.Printf("Error printing message: %v\n", err)
	}
	//fmt.Printf("Message: %v\n", string(formattedMsg))
}

// Print prints a message
func (m *dumbMessage) print() {
	_, err := json.Marshal(*m)
	if err != nil {
		fmt.Printf("Error printing message: %v\n", err)
	}
	//fmt.Printf("Message: %v\n", string(formattedMsg))
}
