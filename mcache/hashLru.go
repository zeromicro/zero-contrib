package mcache

import (
	"crypto/md5"
	"github.com/songangweb/mcache/simplelru"
	"math"
	"runtime"
	"sync"
)

// HashLruCache is a thread-safe fixed size LRU cache.
type HashLruCache struct {
	list     []*HashLruCacheOne
	sliceNum int
	size     int
}

type HashLruCacheOne struct {
	lru  simplelru.LRUCache
	lock sync.RWMutex
}

// NewHashLRU creates an LRU of the given size.
func NewHashLRU(size, sliceNum int) (*HashLruCache, error) {
	return NewHashLruWithEvict(size, sliceNum, nil)
}

// NewHashLruWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewHashLruWithEvict(size, sliceNum int, onEvicted func(key interface{}, value interface{}, expirationTime int64)) (*HashLruCache, error) {
	if 0 == sliceNum {
		sliceNum = runtime.NumCPU()
	}
	if size < sliceNum {
		size = sliceNum
	}

	lruLen := int(math.Ceil(float64(size / sliceNum)))
	var h HashLruCache
	h.size = size
	h.sliceNum = sliceNum
	h.list = make([]*HashLruCacheOne, sliceNum)
	for i := 0; i < sliceNum; i++ {
		l, _ := simplelru.NewLRU(lruLen, onEvicted)
		h.list[i] = &HashLruCacheOne{
			lru: l,
		}
	}

	return &h, nil
}

// Purge is used to completely clear the cache.
func (h *HashLruCache) Purge() {
	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		h.list[i].lru.Purge()
		h.list[i].lock.Unlock()
	}
}

// PurgeOverdue is used to completely clear the overdue cache.
func (h *HashLruCache) PurgeOverdue() {
	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		h.list[i].lru.PurgeOverdue()
		h.list[i].lock.Unlock()
	}
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (h *HashLruCache) Add(key interface{}, value interface{}, expirationTime int64) (evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	evicted = h.list[sliceKey].lru.Add(key, value, expirationTime)
	h.list[sliceKey].lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (h *HashLruCache) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	value, expirationTime, ok = h.list[sliceKey].lru.Get(key)
	h.list[sliceKey].lock.Unlock()
	return value, expirationTime, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (h *HashLruCache) Contains(key interface{}) bool {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.RLock()
	containKey := h.list[sliceKey].lru.Contains(key)
	h.list[sliceKey].lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (h *HashLruCache) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.RLock()
	value, expirationTime, ok = h.list[sliceKey].lru.Peek(key)
	h.list[sliceKey].lock.RUnlock()
	return value, expirationTime, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (h *HashLruCache) ContainsOrAdd(key interface{}, value interface{}, expirationTime int64) (ok, evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	defer h.list[sliceKey].lock.Unlock()

	if h.list[sliceKey].lru.Contains(key) {
		return true, false
	}
	evicted = h.list[sliceKey].lru.Add(key, value, expirationTime)
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (h *HashLruCache) PeekOrAdd(key interface{}, value interface{}, expirationTime int64) (previous interface{}, ok, evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	defer h.list[sliceKey].lock.Unlock()

	previous, expirationTime, ok = h.list[sliceKey].lru.Peek(key)
	if ok {
		return previous, true, false
	}

	evicted = h.list[sliceKey].lru.Add(key, value, expirationTime)
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (h *HashLruCache) Remove(key interface{}) (present bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	present = h.list[sliceKey].lru.Remove(key)
	h.list[sliceKey].lock.Unlock()
	return
}

// Resize changes the cache size.
func (h *HashLruCache) Resize(size int) (evicted int) {
	if size < h.sliceNum {
		size = h.sliceNum
	}

	lruLen := int(math.Ceil(float64(size / h.sliceNum)))

	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		evicted = h.list[i].lru.Resize(lruLen)
		h.list[i].lock.Unlock()
	}
	return evicted
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (h *HashLruCache) Keys() []interface{} {

	var keys []interface{}

	allKeys := make([][]interface{}, h.sliceNum)

	var oneKeysMaxLen int

	for s := 0; s < h.sliceNum; s++ {
		h.list[s].lock.RLock()

		if h.list[s].lru.Len() > oneKeysMaxLen {
			oneKeysMaxLen = h.list[s].lru.Len()
		}

		oneKeys := make([]interface{}, h.list[s].lru.Len())
		oneKeys = h.list[s].lru.Keys()
		h.list[s].lock.RUnlock()

		allKeys[s] = oneKeys
	}

	for i := 0; i < h.list[0].lru.Len(); i++ {
		for c := 0; c < len(allKeys); c++ {
			if len(allKeys[c]) > i {
				keys = append(keys, allKeys[c][i])
			}
		}
	}

	return keys
}

// Len returns the number of items in the cache.
func (h *HashLruCache) Len() int {
	var length = 0

	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.RLock()
		length = length + h.list[i].lru.Len()
		h.list[i].lock.RUnlock()
	}
	return length
}

func (h *HashLruCache) modulus(key *interface{}) int {
	str := InterfaceToString(*key)
	return int(md5.Sum([]byte(str))[0]) % h.sliceNum
}
