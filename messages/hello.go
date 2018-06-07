package messages

import "log"

// HandleSingleHello is a handler for hello messages to let broker know node is alive
func (client *Client) HandleSingleHello() {
	var msg *Message
	for msg == nil {
		msg = client.ReceiveMessage()
		if msg != nil && msg.Type != "hello" {
			client.buffer <- *msg
			msg = nil
		}
	}
	if msg.Type != "hello" {
		log.Fatal("NO HELLO MESSAGE!\n")
	}
	err := client.SendToBroker(Message{Type: "helloResponse", Source: client.NodeName})
	if err != nil {
		log.Fatal("ERROR SENDING HELLO MESSAGE!\n")
	}

}
