package mcache

import (
	"container/list"
	"errors"
	"time"
)

// SimpleLRUEvictCallback is used to get a callback when a cache simpleLFUEntry is evicted
type SimpleLRUEvictCallback func(key interface{}, value interface{}, expirationTime int64)

// SimpleLRU implements a non-thread safe fixed size SimpleLRU cache
type SimpleLRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   SimpleLRUEvictCallback
}

// simpleLRUEntry is used to hold a value in the evictList
type simpleLRUEntry struct {
	key            interface{}
	value          interface{}
	expirationTime int64
}

// NewSimpleLRU constructs an SimpleLRU of the given size
func NewSimpleLRU(size int, onEvict SimpleLRUEvictCallback) (*SimpleLRU, error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &SimpleLRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *SimpleLRU) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*simpleLRUEntry).value, v.Value.(*simpleLRUEntry).expirationTime)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// PurgeOverdue is used to completely clear the overdue cache.
func (c *SimpleLRU) PurgeOverdue() {
	for _, ent := range c.items {
		// 判断此值是否已经超时,如果超时则进行删除
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
		}
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *SimpleLRU) Add(key, value interface{}, expirationTime int64) (ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*simpleLRUEntry).value = value
		ent.Value.(*simpleLRUEntry).expirationTime = expirationTime
		return true
	}
	if c.evictList.Len() >= c.size {
		c.removeOldest()
	}
	ent := &simpleLRUEntry{key, value, expirationTime}

	c.items[key] = c.evictList.PushFront(ent)
	return true
}

// Get looks up a key's value from the cache.
func (c *SimpleLRU) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	if ent, ok := c.items[key]; ok {
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, false
		}
		c.evictList.MoveToFront(ent)
		return ent.Value.(*simpleLRUEntry).value, ent.Value.(*simpleLRUEntry).expirationTime, true
	}
	return nil, 0, false
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *SimpleLRU) Contains(key interface{}) (ok bool) {
	ent, ok := c.items[key]
	if ok {
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
			return !ok
		}
	}
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *SimpleLRU) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, ok
		}
		return ent.Value.(*simpleLRUEntry).value, ent.Value.(*simpleLRUEntry).expirationTime, true
	}
	return nil, 0, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *SimpleLRU) Remove(key interface{}) (ok bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return ok
	}
	return ok
}

// RemoveOldest removes the oldest item from the cache.
func (c *SimpleLRU) RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
			return c.RemoveOldest()
		}

		c.removeElement(ent)
		return ent.Value.(*simpleLRUEntry).key, ent.Value.(*simpleLRUEntry).value, ent.Value.(*simpleLRUEntry).expirationTime, true
	}
	return nil, nil, 0, false
}

// GetOldest returns the oldest simpleLFUEntry
func (c *SimpleLRU) GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		if checkExpirationTime(ent.Value.(*simpleLRUEntry).expirationTime) {
			c.removeElement(ent)
			return c.GetOldest()
		}
		return ent.Value.(*simpleLRUEntry).key, ent.Value.(*simpleLRUEntry).value, ent.Value.(*simpleLRUEntry).expirationTime, true
	}
	return nil, nil, 0, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *SimpleLRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*simpleLRUEntry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *SimpleLRU) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *SimpleLRU) Resize(size int) (evicted int) {
	diff := c.Len() - size
	if diff < 0 {
		diff = 0
	}
	for i := 0; i < diff; i++ {
		c.removeOldest()
	}
	c.size = size
	return diff
}

// removeOldest removes the oldest item from the cache.
func (c *SimpleLRU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *SimpleLRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	delete(c.items, e.Value.(*simpleLRUEntry).key)
	if c.onEvict != nil {
		c.onEvict(e.Value.(*simpleLRUEntry).key, e.Value.(*simpleLRUEntry).value, e.Value.(*simpleLRUEntry).expirationTime)
	}
}

// SimpleCheckExpirationTime is Determine if the cache has expired
func checkExpirationTime(expirationTime int64) (ok bool) {
	if 0 != expirationTime && expirationTime <= time.Now().UnixNano()/1e6 {
		return true
	}
	return false
}
