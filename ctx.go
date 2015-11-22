package neptulon

// Ctx is the incoming message context.
type Ctx struct {
	Conn Conn
	Msg  []byte
	Res  []byte
}
