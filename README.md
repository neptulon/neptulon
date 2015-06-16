Neptulon
========

[![Build Status](https://travis-ci.org/nbusy/neptulon.svg?branch=master)](https://travis-ci.org/nbusy/neptulon) [![GoDoc](https://godoc.org/github.com/nbusy/neptulon?status.svg)](https://godoc.org/github.com/nbusy/neptulon)

Neptulon is a socket framework with middleware support. Framework core is built on listener and context objects. Each message on each connection creates a context which is then passed on to the registered middleware for handling. Client server communication is full-duplex bidirectional.

Framework core is a small ~1000 SLOC codebase which makes it easy to fork, specialize, and maintain for specific purposes, if you need to.

Example
-------

```go
nep, _ := neptulon.NewApp(cert, privKey, "127.0.0.1:3000", true)
jrpc, _ := jsonrpc.NewApp(nep)
rout, _ := jsonrpc.NewRouter(jrpc)

rout.Request("echo", func(conn *neptulon.Conn, req *jsonrpc.Request) (res interface{}, err *jsonrpc.ResError) {
	return req.Params, nil
})
```

Testing
-------

All the tests can be executed by `GORACE="halt_on_error=1" go test -race -cover ./...` command. Optionally you can add `-v` flag to observe all connection logs.

License
-------

[MIT](LICENSE)
