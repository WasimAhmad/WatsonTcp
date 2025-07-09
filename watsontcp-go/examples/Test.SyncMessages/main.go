package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
)

func main() {
	var conn net.Conn
	cb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) { conn = c },
		OnMessage: func(id string, msg *message.Message, data []byte) {
			if msg.SyncRequest {
				resp := &message.Message{SyncResponse: true, ConversationGUID: msg.ConversationGUID}
				hdr, _ := message.BuildHeader(resp)
				conn.Write(hdr)
				resp.ContentLength = int64(len("world"))
				conn.Write([]byte("world"))
			}
		},
	}
	srv := server.New("127.0.0.1:9001", nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	cl := client.New("127.0.0.1:9001", nil, client.Callbacks{}, nil)
	if err := cl.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cl.Disconnect()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, data, err := cl.SendSync(ctx, &message.Message{}, []byte("hello"))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("response: %s (%v)\n", string(data), resp.SyncResponse)
}
