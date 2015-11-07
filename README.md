# Neptulon

[![Build Status](https://travis-ci.org/neptulon/neptulon.svg?branch=master)](https://travis-ci.org/neptulon/neptulon)
[![GoDoc](https://godoc.org/github.com/neptulon/neptulon?status.svg)](https://godoc.org/github.com/neptulon/neptulon)

Neptulon is a socket framework with middleware support. Framework core is built on listener and context objects. Each message on each connection creates a context which is then passed on to the registered middleware for handling. Client server communication is full-duplex bidirectional.

Neptulon framework is only ~400 lines of code, which makes it easy to fork, specialize, and maintain for specific purposes, if you need to.

## TLS Only

Currently we only support TLS for communication. Raw TCP/UDP and DTLS support is planned for future iterations.

## JSON-RPC 2.0

[jsonrpc](https://github.com/neptulon/jsonrpc) package contains JSON-RPC 2.0 implementation on top of Neptulon. You can see a basic example below.

## Example

Following example creates a TLS listener with JSON-RPC 2.0 protocol and starts listening for 'ping' requests and replies with a typical 'pong'.

```go
nep, _ := neptulon.NewServer(cert, privKey, nil, "127.0.0.1:3000", true)
rpc, _ := jsonrpc.NewServer(nep)
route, _ := jsonrpc.NewRouter(rpc)

route.Request("ping", func(ctx *jsonrpc.ReqCtx) {
	ctx.Res = "pong"
})

nep.Run()
```

## Users

[Titan](https://github.com/nb-titan/titan) mobile messaging server is written entirely using the Neptulon framework. It uses JSON-RPC 2.0 package over Neptulon to act as the server part of a mobile messaging app. You can visit its repo to see a complete use case of Neptulon framework.

## Testing

All the tests can be executed with `GORACE="halt_on_error=1" go test -race -cover ./...` command. Optionally you can add `-v` flag to observe all connection logs.

## License

[MIT](LICENSE)
