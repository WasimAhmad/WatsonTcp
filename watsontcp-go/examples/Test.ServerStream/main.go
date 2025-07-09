package main

import (
	"log"
	"net"
	"os"

	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatalf("usage: %s <file>", os.Args[0])
	}
	file := os.Args[1]

	var srv *server.Server

	cb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			f, err := os.Open(file)
			if err != nil {
				log.Println(err)
				return
			}
			fi, _ := f.Stat()
			defer f.Close()
			srv.SendStream(id, &message.Message{}, f, fi.Size())
		},
	}

	srv = server.New("127.0.0.1:9200", nil, cb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()
	log.Printf("serving %s on 127.0.0.1:9200", file)
	select {}
}
