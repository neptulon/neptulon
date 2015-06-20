package neptulon

import "sync"

// Session is a thread-safe data store.
type Session struct {
	error        error
	disconnected bool
	data         map[string]interface{}
	mutex        sync.RWMutex
}

// NewSession creates and returns a new session object.
func NewSession() *Session {
	return &Session{data: make(map[string]interface{})}
}

// Set stores a value for a given key in the session.
func (s *Session) Set(key string, val interface{}) {
	s.mutex.Lock()
	s.data[key] = val
	s.mutex.Unlock()
}

// Get retrieves a value for a given key in the session.
func (s *Session) Get(key string) interface{} {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	return s.data[key]
}
