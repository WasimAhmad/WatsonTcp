package main

import (
	"fmt"
	"log"
	"time"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
)

func main() {
	srvCb := server.Callbacks{
		OnMessage: func(id string, msg *message.Message, data []byte) {
			fmt.Printf("server received '%s' with %d metadata entries\n", string(data), len(msg.Metadata))
		},
	}
	srv := server.New("127.0.0.1:9102", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(time.Second)

	cli := client.New("127.0.0.1:9102", nil, client.Callbacks{}, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	md := map[string]any{
		"foo":    "bar",
		"number": 42,
	}
	msg := &message.Message{Metadata: md}
	if err := cli.Send(msg, []byte("hello")); err != nil {
		log.Fatal(err)
	}

	time.Sleep(time.Second)
}
