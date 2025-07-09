# WatsonTcp for Go

WatsonTcp provides a simple TCP messaging framework with message framing,
authentication, and connection management. This Go implementation shares the
same wire protocol as the C# library and can communicate with it directly.

## Features

- Message framing compatible with WatsonTcp for C#
- TLS encryption support
- Optional preshared key authentication
- Idle timeouts and keepalive settings
- Send and receive byte slices or streams
- Synchronous request/response messaging
- Connection filters (allow/deny lists)
- Connection limit enforcement
- Runtime statistics (bytes and messages sent/received)
- Optional debug logging with customizable logger

## Installation

```
go get github.com/yourname/watsontcp-go
```

Import the packages you need:

```go
import (
    "github.com/yourname/watsontcp-go/client"
    "github.com/yourname/watsontcp-go/server"
)
```

## Basic Usage

### Server
```go
package main

import (
    "fmt"
    "log"

    "github.com/yourname/watsontcp-go/message"
    "github.com/yourname/watsontcp-go/server"
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
```

### Client
```go
package main

import (
    "log"

    "github.com/yourname/watsontcp-go/client"
    "github.com/yourname/watsontcp-go/message"
)

func main() {
    cb := client.Callbacks{}
    cb.OnMessage = func(msg *message.Message, data []byte) {
        log.Printf("server: %s", string(data))
    }

    c := client.New("127.0.0.1:9000", nil, cb, nil)
    if err := c.Connect(); err != nil {
        log.Fatal(err)
    }
    defer c.Disconnect()

    c.Send(&message.Message{}, []byte("hello"))
}
```

### Debug Logging

Both the client and server accept `Options` structures which enable debug
logging. Set `DebugMessages` to `true` and provide a `Logger` function matching
`fmt.Printf` to receive logs whenever messages are sent or received.

```go
opts := client.DefaultOptions()
opts.Logger = log.Printf
opts.DebugMessages = true
c := client.New("127.0.0.1:9000", nil, cb, &opts)
```

## Examples

The `examples` directory contains small programs that demonstrate most
library features:

- `Test.Client` – interactive console client
- `Test.Server` – simple console server
- `Test.FileTransfer` – streaming data between client and server
- `Test.Metadata` – sending messages with metadata maps
- `Test.Parallel` – multiple clients sending concurrently
- `Test.Reconnect` – reconnect logic for unreliable networks
- `Test.SyncMessages` – synchronous request/response messaging

Build an example with:

```
go build ./examples/<ExampleName>
```

## Compatibility with C#

Because the framing and message structure mirror the C# implementation, a Go
client can communicate with a C# server and vice versa. Ensure that features
such as preshared key authentication are enabled or disabled consistently on
both sides.
