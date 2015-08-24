package neptulon

import "sync"

// ConcurrentMap is a thread-safe map.
type ConcurrentMap struct {
	data  map[interface{}]interface{}
	mutex sync.RWMutex
}

// NewConcurrentMap creates and returns a new thread-safe map.
func NewConcurrentMap() *ConcurrentMap {
	return &ConcurrentMap{data: make(map[interface{}]interface{})}
}

// Set stores a value for a given key.
func (c *ConcurrentMap) Set(key interface{}, val interface{}) {
	c.mutex.Lock()
	c.data[key] = val
	c.mutex.Unlock()
}

// Get retrieves a value for a given key.
func (c *ConcurrentMap) Get(key interface{}) (val interface{}, ok bool) {
	c.mutex.RLock()
	val, ok = c.data[key]
	c.mutex.RUnlock()
	return
}
