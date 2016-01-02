# Neptulon

[![Build Status](https://travis-ci.org/neptulon/neptulon.svg?branch=master)](https://travis-ci.org/neptulon/neptulon)
[![GoDoc](https://godoc.org/github.com/neptulon/neptulon?status.svg)](https://godoc.org/github.com/neptulon/neptulon)

Neptulon is a bidirectional RPC framework with middleware support. Framework core is built on listener and context objects. Each message on each connection creates a context which is then passed on to the registered middleware for handling. Client server communication is full-duplex bidirectional.

Neptulon framework is only ~400 lines of code, which makes it easy to fork, specialize, and maintain for specific purposes, if you need to.

## TCP/TLS and WebSocket Support

Currently the TCP + TLS support is in place. WebSocket support is being planned ([see #49](https://github.com/neptulon/neptulon/issues/49)). UDP/SRTP + DTLS might we considered for future if the need arises.

## Communication Protocols

Neptulon is built for speed and massive scalability. For that reason, the protocol is very simple:

```
+-------------------------------+---------+
| 4 Bytes Header (payload size) | Payload |
+-------------------------------+---------+
```

This simplicity makes client writing a breeze. We also plan to add support for WebSocket protocol for Web clients.

## JSON-RPC 2.0

[jsonrpc](https://github.com/neptulon/jsonrpc) package contains JSON-RPC 2.0 implementation on top of Neptulon. You can see a basic example below.

## Example

Following is a raw TCP server for echoing all incoming messages as is to the client.

```go
s, err := neptulon.NewTCPServer("127.0.0.1:3001", false)
if err != nil {
	log.Fatalln("Failed to start Neptulon server:", err)
}

// middleware for echoing all incoming messages as is
s.MiddlewareIn(func(ctx *client.Ctx) {
	ctx.Client.Send(ctx.Msg)
	ctx.Next()
})

s.Start()
```

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

## Comparison to Other Frameworks

We designed Neptulon with a singular focus with minimal dependencies. Using another framework is risky on long term if your project grows as you'll have to maintain it yourself at some point. Below are the options that we evaluated and why Neptulon came to be.

**Neptulon (Go)**:
* Very small codebase with singular focus making it possible to specialize and maintain if necessary.
* Middleware based as in Express.
* Blazingly fast and secure with optional TLS session reuse and client certificate authentication, making it fit for millions of connections per machine.
* Uses very basic `header[payload-size]+payload` protocol with optional JSON-RPC package.
* Other platform clients are all < ~200 lines of code, thanks to simple protocol.
* Big bet on Go's future and big dependency on Go runtime and std lib.

**Go net/rpc**:
* Not bidirectional so not evaluated. Can possibly be made bidirectional with custom codec but adds bloat.

**Koding Kite (Go)**:
* Large codebase with extra features.
* WebSocket based. Adds HTTP/WebSockets as dependencies which makes client connection phase very bloated.
* Similar interface to `go/http` package.
* Uses `dnode` protocol.
* Can reuse existing WebSocket client packages on any platform.
* Big bet on Go's future and big dependency on Go runtime and std lib.

**Google gRPC (Go/C)**:
* Go and C mixed code base which is gigantic for what it does. Huge amount of extra features.
* Monolithic with plugins.
* Uses protocol-buffers with optional JSON plugin.
* Different platform clients are provided but they are quite big in code size.
* Common platform is written in C and hence does not have a singular runtime or std lib dependency.

**Node.js + WebSockets (JavaScript)**:
* This was our initial approach. Existing libraries and frameworks are amazing. On the other hand, getting TLS session reuse, client cert authentication, etc. to work requires more effort than writing entire Neptulon framework.
* Lots of moving parts are out of our control.
* Big bet on Node's future and big dependency on V8 runtime and std lib.

## License

[MIT](LICENSE)
