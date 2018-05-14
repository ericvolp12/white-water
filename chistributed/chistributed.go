package chistributed

import (
	"fmt"

	zmq "github.com/pebbe/zmq4"
)

func main() {
	requester, _ := zmq.NewSocket(zmq.REQ)
	defer requester.Close()
	requester.Connect("tcp://127.0.0.1:23311")

	for request := 0; request < 10; request++ {
		requester.Send("Hello", 0)
		reply, _ := requester.Recv(0)
		fmt.Printf("Received reply %d [%s]\n", request, reply)
	}
}
