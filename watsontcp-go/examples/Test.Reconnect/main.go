package main

import (
	"fmt"
	"log"
	"math/rand"
	"time"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
)

const addr = "127.0.0.1:9000"

func main() {
	for {
		cli := client.New(addr, nil, client.Callbacks{
			OnMessage: func(msg *message.Message, data []byte) {
				fmt.Printf("server: %s\n", string(data))
			},
		}, nil)

		if err := cli.Connect(); err != nil {
			log.Println("connect:", err)
			time.Sleep(time.Second)
			continue
		}

		fmt.Printf("%s connected\n", time.Now().UTC().Format(time.RFC3339))
		time.Sleep(time.Duration(rand.Intn(2000)+1000) * time.Millisecond)
		cli.Disconnect()
	}
}
