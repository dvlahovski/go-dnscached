package cache

import (
	"github.com/miekg/dns"
)

type CacheEntry struct {
	ttl   int
	hits  int
	value dns.Msg
}

type Cache struct {
	cache map[string]CacheEntry
}

func NewCache() *Cache {
	c := new(Cache)
	c.cache = make(map[string]CacheEntry)
	return c
}

func (c *Cache) LookUp(key string) bool {
	_, ok := c.cache[key]
	return ok
}

func (c *Cache) Insert(key string, value dns.Msg) bool {
	if _, ok := c.cache[key]; ok {
		return false
	}

	entry := new(CacheEntry)
	entry.ttl = 240
	entry.hits = 0
	entry.value = value

	c.cache[key] = *entry

	return true
}

func (c *Cache) Get(key string) (dns.Msg, bool) {
	entry, ok := c.cache[key]
	if !ok {
		return dns.Msg{}, false
	}

	entry.hits++
	c.cache[key] = entry
	return c.cache[key].value, true
}

func (c *Cache) Delete(key string) {
	delete(c.cache, key)
}
