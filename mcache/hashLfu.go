package mcache

import (
	"crypto/md5"
	"math"
	"runtime"
	"sync"
)

// HashLfuCache is a thread-safe fixed size HashLFU cache.
type HashLfuCache struct {
	list     []*HashLfuCacheOne
	sliceNum int
	size     int
}

type HashLfuCacheOne struct {
	lfu  SimpleLFUCache
	lock sync.RWMutex
}

// NewHashLFU creates an SimpleLFU of the given size.
func NewHashLFU(size, sliceNum int) (*HashLfuCache, error) {
	return NewHashLfuWithEvict(size, sliceNum, nil)
}

// NewHashLfuWithEvict constructs a fixed size cache with the given eviction
// callback.
func NewHashLfuWithEvict(size, sliceNum int, onEvicted func(key interface{}, value interface{}, expirationTime int64)) (*HashLfuCache, error) {
	if 0 == sliceNum {
		sliceNum = runtime.NumCPU()
	}
	if size < sliceNum {
		size = sliceNum
	}

	lfuLen := int(math.Ceil(float64(size / sliceNum)))
	var h HashLfuCache
	h.size = size
	h.sliceNum = sliceNum
	h.list = make([]*HashLfuCacheOne, sliceNum)
	for i := 0; i < sliceNum; i++ {
		l, _ := NewSimpleLFU(lfuLen, onEvicted)
		h.list[i] = &HashLfuCacheOne{
			lfu: l,
		}
	}

	return &h, nil
}

// Purge is used to completely clear the cache.
func (h *HashLfuCache) Purge() {
	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		h.list[i].lfu.Purge()
		h.list[i].lock.Unlock()
	}
}

// PurgeOverdue is used to completely clear the overdue cache.
func (h *HashLfuCache) PurgeOverdue() {
	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		h.list[i].lfu.PurgeOverdue()
		h.list[i].lock.Unlock()
	}
}

// Add adds a value to the cache. Returns true if an eviction occurred.
func (h *HashLfuCache) Add(key interface{}, value interface{}, expirationTime int64) (evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	evicted = h.list[sliceKey].lfu.Add(key, value, expirationTime)
	h.list[sliceKey].lock.Unlock()
	return evicted
}

// Get looks up a key's value from the cache.
func (h *HashLfuCache) Get(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	value, expirationTime, ok = h.list[sliceKey].lfu.Get(key)
	h.list[sliceKey].lock.Unlock()
	return value, expirationTime, ok
}

// Contains checks if a key is in the cache, without updating the
// recent-ness or deleting it for being stale.
func (h *HashLfuCache) Contains(key interface{}) bool {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.RLock()
	containKey := h.list[sliceKey].lfu.Contains(key)
	h.list[sliceKey].lock.RUnlock()
	return containKey
}

// Peek returns the key value (or undefined if not found) without updating
// the "recently used"-ness of the key.
func (h *HashLfuCache) Peek(key interface{}) (value interface{}, expirationTime int64, ok bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.RLock()
	value, expirationTime, ok = h.list[sliceKey].lfu.Peek(key)
	h.list[sliceKey].lock.RUnlock()
	return value, expirationTime, ok
}

// ContainsOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (h *HashLfuCache) ContainsOrAdd(key interface{}, value interface{}, expirationTime int64) (ok, evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	defer h.list[sliceKey].lock.Unlock()

	if h.list[sliceKey].lfu.Contains(key) {
		return true, false
	}
	evicted = h.list[sliceKey].lfu.Add(key, value, expirationTime)
	return false, evicted
}

// PeekOrAdd checks if a key is in the cache without updating the
// recent-ness or deleting it for being stale, and if not, adds the value.
// Returns whether found and whether an eviction occurred.
func (h *HashLfuCache) PeekOrAdd(key interface{}, value interface{}, expirationTime int64) (previous interface{}, ok, evicted bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	defer h.list[sliceKey].lock.Unlock()

	previous, expirationTime, ok = h.list[sliceKey].lfu.Peek(key)
	if ok {
		return previous, true, false
	}

	evicted = h.list[sliceKey].lfu.Add(key, value, expirationTime)
	return nil, false, evicted
}

// Remove removes the provided key from the cache.
func (h *HashLfuCache) Remove(key interface{}) (present bool) {
	sliceKey := h.modulus(&key)

	h.list[sliceKey].lock.Lock()
	present = h.list[sliceKey].lfu.Remove(key)
	h.list[sliceKey].lock.Unlock()
	return
}

// Resize changes the cache size.
func (h *HashLfuCache) Resize(size int) (evicted int) {
	if size < h.sliceNum {
		size = h.sliceNum
	}

	lfuLen := int(math.Ceil(float64(size / h.sliceNum)))

	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		evicted = h.list[i].lfu.Resize(lfuLen)
		h.list[i].lock.Unlock()
	}
	return evicted
}

func (h *HashLfuCache) ResizeWeight(percentage int) {
	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.Lock()
		h.list[i].lfu.ResizeWeight(percentage)
		h.list[i].lock.Unlock()
	}
}

// Keys returns a slice of the keys in the cache, from oldest to newest.
func (h *HashLfuCache) Keys() []interface{} {

	var keys []interface{}

	allKeys := make([][]interface{}, h.sliceNum)

	var oneKeysMaxLen int

	for s := 0; s < h.sliceNum; s++ {
		h.list[s].lock.RLock()

		if h.list[s].lfu.Len() > oneKeysMaxLen {
			oneKeysMaxLen = h.list[s].lfu.Len()
		}

		oneKeys := make([]interface{}, h.list[s].lfu.Len())
		oneKeys = h.list[s].lfu.Keys()
		h.list[s].lock.RUnlock()

		allKeys[s] = oneKeys
	}

	for i := 0; i < h.list[0].lfu.Len(); i++ {
		for c := 0; c < len(allKeys); c++ {
			if len(allKeys[c]) > i {
				keys = append(keys, allKeys[c][i])
			}
		}
	}

	return keys
}

// Len returns the number of items in the cache.
func (h *HashLfuCache) Len() int {
	var length = 0

	for i := 0; i < h.sliceNum; i++ {
		h.list[i].lock.RLock()
		length = length + h.list[i].lfu.Len()
		h.list[i].lock.RUnlock()
	}
	return length
}

func (h *HashLfuCache) modulus(key *interface{}) int {
	str := InterfaceToString(*key)
	return int(md5.Sum([]byte(str))[0]) % h.sliceNum
}
