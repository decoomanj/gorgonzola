package gorgonzola

import (
	"container/list"
	"sync"
	"time"
)

type TTLList struct {
	sync.RWMutex
	list *list.List
}

type TTLItem struct {
	sync.RWMutex
	expires *time.Time
	item    interface{}
}

func NewTTLList() TTLList {
	ttlList := TTLList{list: list.New()}
	ttlList.startCleanupTimer()
	return ttlList
}

func (l *TTLList) Add(item interface{}, d time.Duration) {
	l.Lock()
	defer l.Unlock()
	createDate := time.Now().Add(d)
	l.list.PushBack(&TTLItem{
		expires: &createDate,
		item:    item,
	})
}

// Cleanup
func (l *TTLList) cleanup() {
	l.Lock()
	defer l.Unlock()

	e := l.list.Front()
	for {
		if e != nil {
			next := e.Next()
			if e.Value.(*TTLItem).expired() {
				l.list.Remove(e)
			}
			e = next
		} else {
			return
		}
	}
}

// Starting cleanup
func (l *TTLList) startCleanupTimer() {
	ticker := time.Tick(time.Second)
	go (func() {
		for {
			select {
			case <-ticker:
				l.cleanup()
			}
		}
	})()
}

// Check if an object has expired
func (item *TTLItem) expired() bool {
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
