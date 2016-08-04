package gorgonzola

import (
	"testing"
	"time"
)

func TestSingleElement(t *testing.T) {

	cache := NewTTLList()
	cache.Add("test", time.Second*2)

	if cache.list.Front() == nil {
		t.Errorf("Element has not been added")
	}

	if cache.list.Front().Value.(*TTLItem).item.(string) != "test" {
		t.Errorf("No front element with a value?")
	}

	<-time.After(time.Second * 3)

	if cache.list.Front() != nil {
		t.Errorf("Element has not been removed")
	}
}

func TestDoubleElement(t *testing.T) {

	cache := NewTTLList()
	cache.Add("test", time.Second*1)
	cache.Add("long", time.Second*8)

	<-time.After(time.Second * 3)

	if cache.list.Len() != 1 {
		t.Errorf("Element has not been removed")
	}
}
