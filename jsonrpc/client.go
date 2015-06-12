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

// Dial creates a new client connection to a given network address with optional root CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func Dial(addr string, rootCA []byte, clientCert []byte, clientCertKey []byte, debug bool) (*Client, error) {
	c, err := neptulon.Dial(addr, rootCA, clientCert, clientCertKey, debug)
	if err != nil {
		return nil, err
	}

	return &Client{conn: c}, nil
}

// ReadMsg reads a message off of a client connection and returns a JSON-RPC Message object.
func (c *Client) ReadMsg() (*Message, error) {
	_, data, err := c.conn.Read()
	if err != nil {
		return nil, err
	}

	var msg Message
	if err = json.Unmarshal(data, &msg); err != nil {
		return nil, err
	}

	return &msg, nil
}

// WriteMsg writes a message to a client connection with error checking.
func (c *Client) WriteMsg(msg interface{}) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	if _, err := c.conn.Write(data); err != nil {
		return err
	}

	return nil
}

// Close closes a client connection.
func (c *Client) Close() error {
	return c.conn.Close()
}
