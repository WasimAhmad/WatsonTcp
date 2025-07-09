package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
)

func main() {
	cb := client.Callbacks{}
	cb.OnMessage = func(msg *message.Message, data []byte) {
		fmt.Printf("server: %s\n", string(data))
	}
	c := client.New("127.0.0.1:9000", nil, cb, nil)
	if err := c.Connect(); err != nil {
		panic(err)
	}
	defer c.Disconnect()
	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("msg> ")
		if !scanner.Scan() {
			break
		}
		txt := scanner.Text()
		if txt == "" {
			continue
		}
		c.Send(&message.Message{}, []byte(txt))
	}
}
