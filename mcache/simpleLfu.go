package mcache

import (
	"container/list"
	"errors"
	"math"
	"time"
)

// SimpleLFUEvictCallback is used to get a callback when a cache simpleLFUEntry is evicted
type SimpleLFUEvictCallback func(key interface{}, value interface{}, expirationTime int64)

// SimpleLFU implements a non-thread safe fixed size SimpleLFU cache
type SimpleLFU struct {
	size      int
	evictList *list.List
	items     map[interface{}]*list.Element
	onEvict   SimpleLFUEvictCallback
}

// simpleLFUEntry is used to hold a value in the evictList
type simpleLFUEntry struct {
	key            interface{}
	value          interface{}
	weight         int64 // 访问次数
	expirationTime int64
}

// NewSimpleLFU constructs an SimpleLFU of the given size
func NewSimpleLFU(size int, onEvict SimpleLFUEvictCallback) (*SimpleLFU, error) {
	if size <= 0 {
		return nil, errors.New("must provide a positive size")
	}
	c := &SimpleLFU{
		size:      size,
		evictList: list.New(),
		items:     make(map[interface{}]*list.Element),
		onEvict:   onEvict,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *SimpleLFU) Purge() {
	for k, v := range c.items {
		if c.onEvict != nil {
			c.onEvict(k, v.Value.(*simpleLFUEntry).value, v.Value.(*simpleLFUEntry).expirationTime)
		}
		delete(c.items, k)
	}
	c.evictList.Init()
}

// PurgeOverdue is used to completely clear the overdue cache.
func (c *SimpleLFU) PurgeOverdue() {
	for _, ent := range c.items {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
		}
	}
	c.evictList.Init()
}

// Add adds a value to the cache.  Returns true if an eviction occurred.
func (c *SimpleLFU) Add(key, value interface{}, expirationTime int64) (ok bool) {
	if ent, ok := c.items[key]; ok {
		ent.Value.(*simpleLFUEntry).value = value
		ent.Value.(*simpleLFUEntry).expirationTime = expirationTime
		ent.Value.(*simpleLFUEntry).weight++
		if (ent.Prev() != nil) && (ent.Prev().Value.(*simpleLFUEntry).weight < ent.Value.(*simpleLFUEntry).weight) {
			c.evictList.MoveBefore(ent, ent.Prev())
		}
		return true
	}

	if c.evictList.Len() >= c.size {
		c.removeOldest()
	}

	ent := &simpleLFUEntry{key, value, 1, expirationTime}
	c.items[key] = c.evictList.PushBack(ent)

	return true
}

// Get looks up a key's value from the cache.
func (c *SimpleLFU) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	if ent, ok := c.items[key]; ok {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, false
		}
		ent.Value.(*simpleLFUEntry).weight++
		if (ent.Prev() != nil) && (ent.Prev().Value.(*simpleLFUEntry).weight < ent.Value.(*simpleLFUEntry).weight) {
			c.evictList.MoveBefore(ent, ent.Prev())
		}
		return ent.Value.(*simpleLFUEntry).value, ent.Value.(*simpleLFUEntry).expirationTime, true
	}
	return nil, 0, false
}

// Contains checks if a key is in the cache, without updating the recent-ness
// or deleting it for being stale.
func (c *SimpleLFU) Contains(key interface{}) (ok bool) {
	ent, ok := c.items[key]
	if ok {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
			return !ok
		}
	}
	return ok
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *SimpleLFU) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	var ent *list.Element
	if ent, ok = c.items[key]; ok {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
			return nil, 0, ok
		}
		return ent.Value.(*simpleLFUEntry).value, ent.Value.(*simpleLFUEntry).expirationTime, true
	}
	return nil, 0, ok
}

// Remove removes the provided key from the cache, returning if the
// key was contained.
func (c *SimpleLFU) Remove(key interface{}) (ok bool) {
	if ent, ok := c.items[key]; ok {
		c.removeElement(ent)
		return ok
	}
	return ok
}

// RemoveOldest removes the oldest item from the cache.
func (c *SimpleLFU) RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	if ent := c.evictList.Back(); ent != nil {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
			return c.RemoveOldest()
		}
		c.removeElement(ent)

		return ent.Value.(*simpleLFUEntry).key, ent.Value.(*simpleLFUEntry).value, ent.Value.(*simpleLFUEntry).expirationTime, true
	}
	return nil, nil, 0, false
}

// GetOldest returns the oldest simpleLFUEntry
func (c *SimpleLFU) GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	ent := c.evictList.Back()
	if ent != nil {
		if SimpleCheckExpirationTime(ent.Value.(*simpleLFUEntry).expirationTime) {
			c.removeElement(ent)
			return c.GetOldest()
		}

		ent.Value.(*simpleLFUEntry).weight++
		return ent.Value.(*simpleLFUEntry).key, ent.Value.(*simpleLFUEntry).value, ent.Value.(*simpleLFUEntry).expirationTime, true
	}
	return nil, nil, 0, false
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *SimpleLFU) Keys() []interface{} {
	keys := make([]interface{}, len(c.items))
	i := 0
	for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
		keys[i] = ent.Value.(*simpleLFUEntry).key
		i++
	}
	return keys
}

// Len returns the number of items in the cache.
func (c *SimpleLFU) Len() int {
	return c.evictList.Len()
}

// Resize changes the cache size.
func (c *SimpleLFU) Resize(size int) (evicted int) {
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

// ResizeWeight changes the cache eight weight size.
func (c *SimpleLFU) ResizeWeight(percentage int) {
	if percentage > 0 || percentage < 100 {
		for ent := c.evictList.Back(); ent != nil; ent = ent.Prev() {
			ent.Value.(*simpleLFUEntry).weight = int64(math.Ceil(float64(ent.Value.(*simpleLFUEntry).weight / 100 * int64(percentage))))
		}
	}
}

// removeOldest removes the oldest item from the cache.
func (c *SimpleLFU) removeOldest() {
	ent := c.evictList.Back()
	if ent != nil {
		c.removeElement(ent)
	}
}

// removeElement is used to remove a given list element from the cache
func (c *SimpleLFU) removeElement(e *list.Element) {
	c.evictList.Remove(e)
	delete(c.items, e.Value.(*simpleLFUEntry).key)
	if c.onEvict != nil {
		c.onEvict(e.Value.(*simpleLFUEntry).key, e.Value.(*simpleLFUEntry).value, e.Value.(*simpleLFUEntry).expirationTime)
	}
}

// SimpleCheckExpirationTime is Determine if the cache has expired
func SimpleCheckExpirationTime(expirationTime int64) (ok bool) {
	if 0 != expirationTime && expirationTime <= time.Now().UnixNano()/1e6 {
		return true
	}
	return false
}
