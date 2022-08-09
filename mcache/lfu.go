package mcache

import (
	"sync"
)

// LfuCache is a thread-safe fixed size SimpleLRU cache.
type LfuCache struct {
	lfu  SimpleLFUCache
	lock sync.RWMutex
}

// NewLFU creates an SimpleLRU of the given size.
func NewLFU(size int) (*LfuCache, error) {
	return NewLfuWithEvict(size, nil)
}

// NewLfuWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewLfuWithEvict(size int, onEvicted func(key interface{}, value interface{}, expirationTime int64)) (*LfuCache, error) {
	lfu, _ := NewSimpleLFU(size, SimpleLFUEvictCallback(onEvicted))
	c := &LfuCache{
		lfu: lfu,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LfuCache) Purge() {
	c.lock.Lock()
	c.lfu.Purge()
	c.lock.Unlock()
}

// PurgeOverdue is used to completely clear the overdue cache.
func (c *LfuCache) PurgeOverdue() {
	c.lock.Lock()
	c.lfu.PurgeOverdue()
	c.lock.Unlock()
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *LfuCache) Add(key, value interface{}, expirationTime int64) (evicted bool) {
	c.lock.Lock()
	evicted = c.lfu.Add(key, value, expirationTime)
	c.lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (c *LfuCache) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	value, expirationTime, ok = c.lfu.Get(key)
	c.lock.Unlock()
	return value, expirationTime, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *LfuCache) Contains(key interface{}) bool {
	c.lock.RLock()
	containKey := c.lfu.Contains(key)
	c.lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LfuCache) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	c.lock.RLock()
	value, expirationTime, ok = c.lfu.Peek(key)
	c.lock.RUnlock()
	return value, expirationTime, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LfuCache) ContainsOrAdd(key, value interface{}, expirationTime int64) (ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lfu.Contains(key) {
		return true, false
	}
	evicted = c.lfu.Add(key, value, expirationTime)
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LfuCache) PeekOrAdd(key, value interface{}, expirationTime int64) (previous interface{}, ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previous, expirationTime, ok = c.lfu.Peek(key)
	if ok {
		return previous, true, false
	}

	evicted = c.lfu.Add(key, value, expirationTime)
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (c *LfuCache) Remove(key interface{}) (present bool) {
	c.lock.Lock()
	present = c.lfu.Remove(key)
	c.lock.Unlock()
	return
}

// Resize changes the cache size.
func (c *LfuCache) Resize(size int) (evicted int) {
	c.lock.Lock()
	evicted = c.lfu.Resize(size)
	c.lock.Unlock()
	return evicted
}

func (c *LfuCache) ResizeWeight(percentage int) {
	c.lock.Lock()
	c.lfu.ResizeWeight(percentage)
	c.lock.Unlock()
}

// RemoveOldest removes the oldest item from the cache.
func (c *LfuCache) RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	key, value, expirationTime, ok = c.lfu.RemoveOldest()
	c.lock.Unlock()
	return
}

// GetOldest returns the oldest simpleLFUEntry
func (c *LfuCache) GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	key, value, expirationTime, ok = c.lfu.GetOldest()
	c.lock.Unlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LfuCache) Keys() []interface{} {
	c.lock.RLock()
	keys := c.lfu.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *LfuCache) Len() int {
	c.lock.RLock()
	length := c.lfu.Len()
	c.lock.RUnlock()
	return length
}
