package gorgonzola

import (
	"sync"
	"time"
)

type TTLDaemon struct {
	sync.RWMutex
	items map[string]*TTL
}

type TTL struct {
	sync.RWMutex
	expires *time.Time
	item    interface{}
}

func NewTTLDaemon() TTLDaemon {
	ttl := TTLDaemon{items: make(map[string]*TTL)}
	ttl.startCleanupTimer()
	return ttl
}

// Refresh the duration of a TTL object
func (item *TTL) touch() {
	item.Lock()
	defer item.Unlock()
	expiration := time.Now().Add(time.Second * 3)
	item.expires = &expiration
}

// Check if an object has expired
func (item *TTL) expired() bool {
	var value bool
	item.RLock()
	defer item.RUnlock()
	if item.expires == nil {
		value = true
	} else {
		value = item.expires.Before(time.Now())
	}
	return value
}

// Cleanup
func (w *TTLDaemon) cleanup() {
	w.Lock()
	defer w.Unlock()
	for key, item := range w.items {
		if item.expired() {
			delete(w.items, key)
		}
	}
}

// Starting cleanup
func (w *TTLDaemon) startCleanupTimer() {
	ticker := time.Tick(time.Second)
	go (func() {
		for {
			select {
			case <-ticker:
				w.cleanup()
			}
		}
	})()
}
