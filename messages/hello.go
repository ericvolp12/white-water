package messages

import "log"

func (client *Client) HandleSingleHello() {
	var msg *Message = nil
	for msg == nil {
		msg = client.ReceiveMessage()
	}
	if msg.Type != "hello" {
		log.Fatal("NO HELLO MESSAGE!\n")
	}
	err := client.SendToBroker(Message{Type: "helloResponse", Source: client.NodeName})
	if err != nil {
		log.Fatal("ERROR SENDING HELLO MESSAGE!\n")
	}

}
