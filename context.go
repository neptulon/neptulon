package neptulon

// Context encapsulates connection, request, and reponse objects.
type Context struct {
	conn *Conn
	msg  []byte
}
