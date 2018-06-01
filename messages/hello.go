package messages

import "log"

//import "fmt"

// helloHandler handles hello messages...
func HelloHandler(client *Client) {
	helloIncoming := make(chan Message, 500)

	client.Subscribe("hello", &helloIncoming)

	for range helloIncoming {
		//		fmt.Printf("Hello Handler Firing...\n")
		err := client.SendToBroker(Message{Type: "helloResponse", Source: client.NodeName})
		client.readyToSend = true
		if err != nil {
			break
		}
	}
}

func (client *Client) HandleSingleHello() {
	msg := client.ReceiveMessage()
	if msg.Type != "hello" {
		log.Fatal("NO HELLO MESSAGE!\n")
	}
	err := client.SendToBroker(Message{Type: "helloResponse", Source: client.NodeName})
	if err != nil {
		log.Fatal("ERROR SENDING HELLO MESSAGE!\n")
	}

}
