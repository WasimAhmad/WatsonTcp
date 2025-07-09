package main

import (
	"fmt"
	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
	"log"
)

func main() {
	cb := server.Callbacks{}
	cb.OnMessage = func(id string, msg *message.Message, data []byte) {
		fmt.Printf("%s: %s\n", id, string(data))
	}
	srv := server.New("127.0.0.1:9000", nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()
	select {}
}
