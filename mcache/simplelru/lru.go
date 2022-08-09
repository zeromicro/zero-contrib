package simplelru

import (
	"container/list"
	"errors"
	"time"
)

// EvictCallback is used to get a callback when a cache entry is evicted
type EvictCallback func(key interface{}, value interface{}, expirationTime int64)

// LRU implements a non-thread safe fixed size LRU cache
type LRU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   EvictCallback
}

// entry is used to hold a value in the evictList
type entry struct {
	key            interface{}
	value          interface{}
	expirationTime int64
}

// NewLRU constructs an LRU of the given size
func NewLRU(size int, onEvict EvictCallback) (*LRU, error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &LRU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LRU) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*entry).value, v.Value.(*entry).expirationTime)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// PurgeOverdue is used to completely clear the overdue cache.
func (c *LRU) PurgeOverdue() {
	for _, ent := range c.items {
		// 判断此值是否已经超时,如果超时则进行删除
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
		}
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *LRU) Add(key, value interface{}, expirationTime int64) (ok bool) {
	if ent, ok := c.items[key]; ok {
		c.evictList.MoveToFront(ent)
		ent.Value.(*entry).value = value
		ent.Value.(*entry).expirationTime = expirationTime
		return true
	}
	if c.evictList.Len() >= c.size {
		c.removeOldest()
	}
	ent := &entry{key, value, expirationTime}

	c.items[key] = c.evictList.PushFront(ent)
	return true
}

// Get looks up a key's value from the cache.
func (c *LRU) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	if ent, ok := c.items[key]; ok {
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, false
		}
		c.evictList.MoveToFront(ent)
		return ent.Value.(*entry).value, ent.Value.(*entry).expirationTime, true
	}
	return nil, 0, false
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *LRU) Contains(key interface{}) (ok bool) {
	ent, ok := c.items[key]
	if ok {
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
			return !ok
		}
	}
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LRU) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, ok
		}
		return ent.Value.(*entry).value, ent.Value.(*entry).expirationTime, true
	}
	return nil, 0, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *LRU) Remove(key interface{}) (ok bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return ok
	}
	return ok
}

// RemoveOldest removes the oldest item from the cache.
func (c *LRU) RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
			return c.RemoveOldest()
		}

		c.removeElement(ent)
		return ent.Value.(*entry).key, ent.Value.(*entry).value, ent.Value.(*entry).expirationTime, true
	}
	return nil, nil, 0, false
}

// GetOldest returns the oldest entry
func (c *LRU) GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		if checkExpirationTime(ent.Value.(*entry).expirationTime) {
			c.removeElement(ent)
			return c.GetOldest()
		}
		return ent.Value.(*entry).key, ent.Value.(*entry).value, ent.Value.(*entry).expirationTime, true
	}
	return nil, nil, 0, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LRU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*entry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *LRU) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *LRU) Resize(size int) (evicted int) {
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
func (c *LRU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *LRU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	delete(c.items, e.Value.(*entry).key)
	if c.onEvict != nil {
		c.onEvict(e.Value.(*entry).key, e.Value.(*entry).value, e.Value.(*entry).expirationTime)
	}
}

// checkExpirationTime is Determine if the cache has expired
func checkExpirationTime(expirationTime int64) (ok bool) {
	if 0 != expirationTime && expirationTime <= time.Now().UnixNano()/1e6 {
		return true
	}
	return false
}
