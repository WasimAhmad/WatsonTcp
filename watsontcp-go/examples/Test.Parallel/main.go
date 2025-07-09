package main

import (
	"crypto/md5"
	"fmt"
	"log"
	"math/rand"
	"sync"
	"time"

	"github.com/yourname/watsontcp-go/client"
	"github.com/yourname/watsontcp-go/message"
	"github.com/yourname/watsontcp-go/server"
)

const (
	serverAddr    = "127.0.0.1:9101"
	clientThreads = 4
	numIterations = 1000
)

func main() {
	data := make([]byte, 262144)
	fmt.Printf("data md5: %x\n", md5.Sum(data))

	srvCb := server.Callbacks{
		OnMessage: func(id string, msg *message.Message, d []byte) {
			fmt.Printf("[server] %s %x (%d bytes)\n", id, md5.Sum(d), len(d))
		},
	}
	srv := server.New(serverAddr, nil, srvCb, nil)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
	defer srv.Stop()

	time.Sleep(time.Second)

	var wg sync.WaitGroup
	for i := 0; i < clientThreads; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cli := client.New(serverAddr, nil, client.Callbacks{}, nil)
			if err := cli.Connect(); err != nil {
				log.Println("connect:", err)
				return
			}
			defer cli.Disconnect()
			for j := 0; j < numIterations; j++ {
				time.Sleep(time.Duration(rand.Intn(25)) * time.Millisecond)
				cli.Send(&message.Message{}, data)
			}
			fmt.Println("[client] finished")
		}()
	}

	wg.Wait()
}
