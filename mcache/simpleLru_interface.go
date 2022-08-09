package mcache

type SimpleLRUCache interface {
	Add(key, value interface{}, expirationTime int64) (ok bool)

	Get(key interface{}) (value interface{}, expirationTime int64, ok bool)

	Contains(key interface{}) (ok bool)

	Peek(key interface{}) (value interface{}, expirationTime int64, ok bool)

	Remove(key interface{}) (ok bool)

	RemoveOldest() (key interface{}, value interface{}, expirationTime int64, ok bool)

	GetOldest() (key interface{}, value interface{}, expirationTime int64, ok bool)

	Keys() []interface{}

	Len() int

	Purge()

	PurgeOverdue()

	Resize(int) int
}
