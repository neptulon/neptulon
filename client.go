package neptulon

// Client is a Neptulon connection client.
type Client interface {
	Send(msg []byte) error
}
