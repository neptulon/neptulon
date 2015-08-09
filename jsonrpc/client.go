package jsonrpc

import (
	"encoding/json"
	"errors"

	"github.com/nbusy/neptulon"
)

// Client is a client implementation for JSON-RPC 2.0 protocol for Neptulon framework.
// Client implementations in other programming languages might be provided in separate repositories so check the documentation.
type Client struct {
	conn *neptulon.Conn
}

// Dial creates a new client connection to a given network address with optional CA and/or a client certificate (PEM encoded X.509 cert/key).
// Debug mode logs all raw TCP communication.
func Dial(addr string, ca []byte, clientCert []byte, clientCertKey []byte, debug bool) (*Client, error) {
	c, err := neptulon.Dial(addr, ca, clientCert, clientCertKey, debug)
	if err != nil {
		return nil, err
	}

	return &Client{conn: c}, nil
}

// SetReadDeadline set the read deadline for the connection in seconds.
func (c *Client) SetReadDeadline(seconds int) {
	c.conn.SetReadDeadline(seconds)
}

// ReadMsg reads a message off of a client connection and returns a request, response, or notification message depending on what server sent.
func (c *Client) ReadMsg() (req *Request, res *Response, not *Notification, err error) {
	_, data, err := c.conn.Read()
	if err != nil {
		return
	}

	var msg message
	if err = json.Unmarshal(data, &msg); err != nil {
		return
	}

	// if incoming message is a request or response
	if msg.ID != "" {
		// if incoming message is a request
		if msg.Method != "" {
			req = &Request{ID: msg.ID, Method: msg.Method, Params: msg.Params}
			return
		}

		// if incoming message is a response
		res = &Response{ID: msg.ID, Result: msg.Result, Error: msg.Error}
		return
	}

	// if incoming message is a notification
	if msg.Method != "" {
		not = &Notification{Method: msg.Method, Params: msg.Params}
	}

	err = errors.New("Received a malformed message.")
	return
}

// WriteRequest writes a JSON-RPC request to a client connection with structured params object and auto generated request ID.
func (c *Client) WriteRequest(method string, params interface{}) (reqID string, err error) {
	id, err := neptulon.GenUID()
	if err != nil {
		return "", err
	}

	return id, c.WriteMsg(Request{ID: id, Method: method, Params: params})
}

// WriteRequestArr writes a JSON-RPC request to a client connection with array params and auto generated request ID.
func (c *Client) WriteRequestArr(method string, params ...interface{}) (reqID string, err error) {
	id, err := neptulon.GenUID()
	if err != nil {
		return "", err
	}

	return id, c.WriteMsg(Request{ID: id, Method: method, Params: params})
}

// WriteMsg writes any JSON-RPC message to a client connection.
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
