
demo:

len := 10  
Cache, _ := m_cache.NewLRU(Len)

Cache.Add(1,1,1614306658000)

Cache.Add(2,2,0)

Cache.Get(2)

