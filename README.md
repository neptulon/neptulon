Neptulon
========

[![Build Status](https://travis-ci.org/nbusy/neptulon.svg?branch=master)](https://travis-ci.org/nbusy/neptulon) [![GoDoc](https://godoc.org/github.com/nbusy/neptulon?status.svg)](https://godoc.org/github.com/nbusy/neptulon)

Neptulon is a socket framework with middleware support. Framework core is built on listener and context objects. Each message on each connection creates a context which is then passed on to the registered middleware for handling. Client server communication is full-duplex bidirectional.

Framework core is a small ~1000 SLOC codebase which makes it easy to fork, specialize, and maintain for specific purposes, if you need to.

Example
-------

Following example creates a TLS listener with JSON-RPC 2.0 protocol and starts listening for 'ping' requests and replies with a typical 'pong'.

// todo: move json-rpc example into its own readme and example_test.go files and link them from here and replace below one with a simple byte ping/pong and a separate sender example

```go
nep, _ := neptulon.NewApp(cert, privKey, nil, "127.0.0.1:3000", true)
rpc, _ := jsonrpc.NewApp(nep)
route, _ := jsonrpc.NewRouter(rpc)

route.Request("ping", func(ctx *jsonrpc.ReqContext) {
	ctx.Res = "pong"
})

nep.Run()
```

Testing
-------

All the tests can be executed by `GORACE="halt_on_error=1" go test -race -cover ./...` command. Optionally you can add `-v` flag to observe all connection logs.

License
-------

[MIT](LICENSE)
