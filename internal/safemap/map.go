package safemap

import (
	"errors"
	"sync"
)

// SafeMap provides a thread-safe map[string]string with basic CRUD operations.
type SafeMap struct {
	store map[string]string
	mu    sync.Mutex
}

// New creates a new instance of SafeMap.
func New() *SafeMap {
	return &SafeMap{store: make(map[string]string)}
}

// Set adds or updates a key-value pair in the SafeMap.
// Returns an error if the key or value is empty.
func (sm *SafeMap) Set(key, value string) error {
	if key == "" || value == "" {
		return errors.New("SafeMap: Key or value cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	sm.store[key] = value

	return nil
}

// Get retrieves the value for a given key from the SafeMap.
// Returns an error if the key is empty or not found.
func (sm *SafeMap) Get(key string) (string, error) {
	if key == "" {
		return "", errors.New("SafeMap: Key cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	val, ok := sm.store[key]
	if !ok {
		return "", errors.New("SafeMap: Key not found")
	}

	return val, nil
}

// Delete removes a key-value pair from the SafeMap.
// Returns an error if the key is empty.
func (sm *SafeMap) Delete(key string) error {
	if key == "" {
		return errors.New("SafeMap: Key cannot be empty")
	}

	sm.mu.Lock()
	defer sm.mu.Unlock()

	delete(sm.store, key)

	return nil
}

// Keys returns a slice of all keys present in the SafeMap.
func (sm *SafeMap) Keys() []string {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	keys := make([]string, 0, len(sm.store))
	for key := range sm.store {
		keys = append(keys, key)
	}

	return keys
}
