# WatsonTcp for Go

This directory contains an experimental Go implementation of WatsonTcp. The API mirrors the C# library where possible and provides simple client and server types.

## Installing

```
go get github.com/yourname/watsontcp-go
```

Import the desired packages:

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
```

### Client
```go
package main

import (
    "github.com/yourname/watsontcp-go/client"
    "github.com/yourname/watsontcp-go/message"
    "log"
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

Both the client and server expose `Options` fields for debug logging. Set
`DebugMessages` to `true` and provide a `Logger` function with the same signature
as `fmt.Printf` to receive logs whenever messages are sent or received.

```go
opts := client.DefaultOptions()
opts.Logger = log.Printf
opts.DebugMessages = true
c := client.New("127.0.0.1:9000", nil, cb, &opts)
```

## Differences from the C# Version

The Go implementation provides the same framing protocol and message structure as the C# library but is a smaller code base with fewer features:

- Only a subset of configuration options are implemented (e.g. keepalive and authentication options are present but not fully enforced).
- Event callbacks are simple function pointers rather than events/delegates.
- Client and server statistics are limited to byte and message counts.

Because the wire protocol is shared, a Go client can communicate with a C# server and vice versa if the framing and message fields match. Features that are not implemented in one language (such as preshared-key authentication) must be handled manually or disabled on the peer.


## Examples

The `examples/` directory contains small programs demonstrating library features. Notable examples include:

- `Test.Metadata` – sending messages with metadata maps.
- `Test.Parallel` – multiple clients sending in parallel to a single server.
- `Test.Reconnect` – reconnect logic that repeatedly connects and disconnects.

Run `go build ./examples/<ExampleName>` to compile an example.
