package mcache

import (
	"sync"
)

// LruCache is a thread-safe fixed size SimpleLRU cache.
type LruCache struct {
	lru  SimpleLRUCache
	lock sync.RWMutex
}

// NewLRU creates an LRU of the given size.
func NewLRU(size int) (*LruCache, error) {
	return NewLruWithEvict(size, nil)
}

// NewLruWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewLruWithEvict(size int, onEvicted func(key interface{}, value interface{}, expirationTime int64)) (*LruCache, error) {
	lru, err := NewSimpleLRU(size, SimpleLRUEvictCallback(onEvicted))
	if err != nil {
		return nil, err
	}
	c := &LruCache{
		lru: lru,
	}
	return c, nil
}

// Purge is used to completely clear the cache.
func (c *LruCache) Purge() {
	c.lock.Lock()
	c.lru.Purge()
	c.lock.Unlock()
}

// PurgeOverdue is used to completely clear the overdue cache.
func (c *LruCache) PurgeOverdue() {
	c.lock.Lock()
	c.lru.PurgeOverdue()
	c.lock.Unlock()
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (c *LruCache) Add(key, value interface{}, expirationTime int64) (evicted bool) {
	c.lock.Lock()
	evicted = c.lru.Add(key, value, expirationTime)
	c.lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (c *LruCache) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	value, expirationTime, ok = c.lru.Get(key)
	c.lock.Unlock()
	return value, expirationTime, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (c *LruCache) Contains(key interface{}) bool {
	c.lock.RLock()
	containKey := c.lru.Contains(key)
	c.lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (c *LruCache) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	c.lock.RLock()
	value, expirationTime, ok = c.lru.Peek(key)
	c.lock.RUnlock()
	return value, expirationTime, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LruCache) ContainsOrAdd(key, value interface{}, expirationTime int64) (ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	if c.lru.Contains(key) {
		return true, false
	}
	evicted = c.lru.Add(key, value, expirationTime)
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (c *LruCache) PeekOrAdd(key, value interface{}, expirationTime int64) (previous interface{}, ok, evicted bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	previous, expirationTime, ok = c.lru.Peek(key)
	if ok {
		return previous, true, false
	}

	evicted = c.lru.Add(key, value, expirationTime)
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (c *LruCache) Remove(key interface{}) (present bool) {
	c.lock.Lock()
	present = c.lru.Remove(key)
	c.lock.Unlock()
	return
}

// Resize changes the cache size.
func (c *LruCache) Resize(size int) (evicted int) {
	c.lock.Lock()
	evicted = c.lru.Resize(size)
	c.lock.Unlock()
	return evicted
}

// RemoveOldest removes the oldest item from the cache.
func (c *LruCache) RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	key, value, expirationTime, ok = c.lru.RemoveOldest()
	c.lock.Unlock()
	return
}

// GetOldest returns the oldest simpleLFUEntry
func (c *LruCache) GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool) {
	c.lock.Lock()
	key, value, expirationTime, ok = c.lru.GetOldest()
	c.lock.Unlock()
	return
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (c *LruCache) Keys() []interface{} {
	c.lock.RLock()
	keys := c.lru.Keys()
	c.lock.RUnlock()
	return keys
}

// Len returns the number of items in the cache.
func (c *LruCache) Len() int {
	c.lock.RLock()
	length := c.lru.Len()
	c.lock.RUnlock()
	return length
}
