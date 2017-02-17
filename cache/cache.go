package cache

import (
	"log"
	"sync"
	"time"

	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

// calculate the min TTL of a slice of dns Answers
func calcTTL(value dns.Msg) uint32 {
	minTTL := value.Answer[0].Header().Ttl
	for _, answer := range value.Answer {
		if answer.Header().Ttl < minTTL {
			minTTL = answer.Header().Ttl
		}
	}

	return minTTL
}

type Entry struct {
	ttl   int
	hits  int
	value dns.Msg
}

type Cache struct {
	cache         map[string]Entry
	capacity      int
	flushInterval int
	lock          sync.Mutex
	config        config.Config
}

func NewCache(config config.Config) *Cache {
	c := new(Cache)
	c.cache = make(map[string]Entry)
	c.lock = *new(sync.Mutex)
	c.config = config
	c.capacity = config.Cache.MaxEntries
	c.flushInterval = config.Cache.FlushInterval

	if c.capacity <= 0 {
		log.Printf("bad capacity value in config, setting to 1000")
		c.capacity = 1000
	}

	c.start()
	return c
}

func (c *Cache) flush() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()
	for key, entry := range c.cache {
		if int64(entry.ttl) <= now {
			log.Printf("deleting key %s", key)
			delete(c.cache, key)
		}
	}
}

func (c *Cache) start() {
	ticker := time.NewTicker(time.Duration(c.flushInterval) * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				c.flush()
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
}

func (c *Cache) Insert(key string, value dns.Msg) bool {
	if len(value.Answer) <= 0 {
		log.Printf("expecting at least one answer in the msg")
		return false
	}

	c.lock.Lock()
	defer c.lock.Unlock()

	if len(c.cache) >= c.capacity {
		return false
	}

	if _, ok := c.cache[key]; ok {
		log.Printf("cache item (%s) exists on insert", key)
		return false
	}

	entry := new(Entry)
	entry.ttl = int(time.Now().Unix() + int64(calcTTL(value)))
	log.Printf("%s ttl %d", key, calcTTL(value))
	entry.hits = 0
	entry.value = value

	c.cache[key] = *entry

	return true
}

func (c *Cache) Get(key string) (dns.Msg, bool) {
	c.lock.Lock()
	defer c.lock.Unlock()

	entry, ok := c.cache[key]
	if !ok {
		return dns.Msg{}, false
	}

	entry.hits++
	c.cache[key] = entry

	return c.cache[key].value, true
}

func (c *Cache) Delete(key string) {
	c.lock.Lock()
	defer c.lock.Unlock()

	delete(c.cache, key)
}
