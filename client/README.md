# Neptulon Go Client

[![Build Status](https://travis-ci.org/neptulon/client.svg?branch=master)](https://travis-ci.org/neptulon/client)
[![GoDoc](https://godoc.org/github.com/neptulon/client?status.svg)](https://godoc.org/github.com/neptulon/client)

Neptulon client implementation in Go. Client-server connection is always full-duplex bidirectional.

## Example

Example assumes that there is a Neptulon server running on local network address 127.0.0.1:3001 running a single echo middleware which echoes all incoming messages back.

```go
import (
	"log"

	"github.com/neptulon/client"
)

func main() {
	c := client.NewClient(nil, nil)
	c.MiddlewareIn(func(ctx *client.Ctx) {
		log.Println("Server's reply:", ctx.Msg)
		ctx.Next()
	})
	c.ConnectTCP("127.0.0.1:3001", false)
	c.Send([]byte("echo"))
	c.Close()
	// Output: Server's reply: echo
}
```

## License

[MIT](LICENSE)
