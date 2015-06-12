package jsonrpc

import (
	"encoding/json"

	"github.com/nbusy/neptulon"
)

// Client is a client implementation for JSON-RPC 2.0 protocol for Neptulon framework.
// Client implementations in other programming languages might be provided in separate repositories so check the documentation.
type Client struct {
	conn *neptulon.Conn
}

// NewClient creates a new client connection to a given network address with optional root CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func NewClient(addr string, rootCA []byte, clientCert []byte, clientCertKey []byte, debug bool) (*Client, error) {
	c, err := neptulon.Dial(addr, rootCA, clientCert, clientCertKey, debug)
	if err != nil {
		return nil, err
	}

	return &Client{conn: c}, nil
}

// ReadMsg reads a message off of a client connection and returns a Message object.
func (c *Client) ReadMsg() (*Message, error) {
	n, data, err := c.conn.Read()
	if err != nil {
		return nil, err
	}

	var msg Message
	if err = json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}
