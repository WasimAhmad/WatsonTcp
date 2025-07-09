package main

import (
	"io"
	"log"
	"net"
	"os"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
)

func main() {
	sample := []byte("sample data for streaming")
	os.WriteFile("sample.txt", sample, 0644)

	var srv *server.Server

	srvCb := server.Callbacks{
		OnConnect: func(id string, c net.Conn) {
			f, err := os.Open("sample.txt")
			if err != nil {
				log.Println(err)
				return
			}
			fi, _ := f.Stat()
			defer f.Close()
			srv.SendStream(id, &message.Message{}, f, fi.Size())
		},
		OnStream: func(id string, msg *message.Message, r io.Reader) {
			out, _ := os.Create("from_client.txt")
			io.Copy(out, r)
			out.Close()
			log.Println("server received file from client")
		},
	}

	srv = server.New("127.0.0.1:9100", nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	cliCb := client.Callbacks{
		OnStream: func(msg *message.Message, r io.Reader) {
			out, _ := os.Create("from_server.txt")
			io.Copy(out, r)
			out.Close()
			log.Println("client received file from server")
		},
	}

	cli := client.New("127.0.0.1:9100", nil, cliCb, nil)
	if err := cli.Connect(); err != nil {
		log.Fatal(err)
	}
	defer cli.Disconnect()

	f, _ := os.Open("sample.txt")
	fi, _ := f.Stat()
	cli.SendStream(&message.Message{}, f, fi.Size())
	f.Close()

	select {}
}
