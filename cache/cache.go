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
	c.cache = make(map[string]dns.Msg)
	return c
}

func (c *Cache) LookUp(key string) bool {
	_, ok := c.cache[key]
	return ok
}

func (c *Cache) Insert(key string, value dns.Msg) {
	// TODO check if value already exists
	entry = new(CacheEntry)
	entry.ttl = 240
	entry.hits = 0
	entry.value = value

	c.cache[key] = entry
}

func (c *Cache) Delete(key string) {
	delete(c.cache, key)
}
