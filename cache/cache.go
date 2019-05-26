package cache

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

// Calculate the min TTL of a slice of dns Answers
func calcTTL(value dns.Msg) uint32 {
	minTTL := value.Answer[0].Header().Ttl
	for _, answer := range value.Answer {
		if answer.Header().Ttl < minTTL {
			minTTL = answer.Header().Ttl
		}
	}

	return minTTL
}

// Create a dummy placeholder dns.Msg from domain namen, ip, record type, ttl
func createPlaceholderMsg(key string, ip string, recordType uint16, ttl int) (*dns.Msg, error) {
	var typeStr string

	if recordType == dns.TypeA {
		typeStr = "A"
	} else if recordType == dns.TypeAAAA {
		typeStr = "AAAA"
	} else {
		return nil, fmt.Errorf("Invalid RR")
	}

	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	// msg.SetQuestion("google.bg.", dns.TypeA)
	msg.SetQuestion(dns.Fqdn(key), recordType)

	var err error
	msg.Answer = make([]dns.RR, 1)
	RR := fmt.Sprintf("%s %d IN %s %s", dns.Fqdn(key), ttl, typeStr, ip)
	msg.Answer[0], err = dns.NewRR(RR)
	if err != nil {
		return nil, fmt.Errorf("Invalid RR")
	}

	return msg, nil
}

// Entry is the cache's internal entry representation
type Entry struct {
	ttl   int
	hits  int
	value dns.Msg
}

// Cache object
type Cache struct {
	cache         map[string]Entry
	capacity      int
	flushInterval int
	lock          sync.Mutex
	config        config.Config
}

// NewCache returns a new cache instance
func NewCache(cfg config.Config) *Cache {
	c := new(Cache)
	c.cache = make(map[string]Entry)
	c.lock = *new(sync.Mutex)
	c.config = cfg
	c.capacity = cfg.Cache.MaxEntries
	c.flushInterval = cfg.Cache.FlushInterval

	if c.capacity <= 0 {
		log.Printf("bad capacity value in config, setting to 1000")
		c.capacity = 1000
	}

	c.hardcodeRecords(cfg.Entries)

	c.start()
	return c
}

// Populate the cache with hardcoded records from the config
func (c *Cache) hardcodeRecords(entries []config.CacheEntry) {
	for _, entry := range entries {
		var recordType uint16
		if entry.Type == "A" {
			recordType = dns.TypeA
		} else if entry.Type == "AAAA" {
			recordType = dns.TypeAAAA
		} else {
			log.Printf("skipping hardode entry")
			continue
		}

		msg, err := createPlaceholderMsg(entry.Key, entry.Value.String(), recordType, entry.Ttl)
		if err != nil {
			log.Printf("skipping hardode entry")
			continue
		}

		c.Insert(dns.Fqdn(entry.Key)+entry.Type+".", *msg)
	}
}

// Flush all the records with expired ttl
func (c *Cache) flush() {
	c.lock.Lock()
	defer c.lock.Unlock()

	now := time.Now().Unix()
	for key, entry := range c.cache {
		if entry.ttl == 0 {
			continue
		}

		if int64(entry.ttl) <= now {
			log.Printf("deleting key %s", key)
			delete(c.cache, key)
		}
	}
}

// Start the ticker that flushes every flushInterval seconds
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

// Insert a DNS msg in the cache
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
	ttl := calcTTL(value)
	if ttl == 0 {
		entry.ttl = 0
	} else {
		if ttl < c.config.Cache.MinTTL {
			ttl = c.config.Cache.MinTTL
		}

		entry.ttl = int(time.Now().Unix() + int64(ttl))
	}

	log.Printf("insert %s ttl %d", key, ttl)
	entry.hits = 0
	entry.value = value

	c.cache[key] = *entry

	return true
}

// InsertFromParams - insert and entry from separate params
func (c *Cache) InsertFromParams(key string, ip string, recordType uint16, ttl int) bool {
	msg, err := createPlaceholderMsg(key, ip, recordType, ttl)
	if err != nil {
		return false
	}

	recordTypeStr := ""
	if recordType == dns.TypeA {
		recordTypeStr = "A"
	} else if recordType == dns.TypeAAAA {
		recordTypeStr = "AAAA"
	}

	return c.Insert(dns.Fqdn(key)+recordTypeStr, *msg)
}

// Get a DNS msg from the cache
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

// GetEntry returns the internal entry
func (c *Cache) GetEntry(key string) (Entry, bool) {
	entry, ok := c.cache[key]
	return entry, ok
}

// Delete an entry
func (c *Cache) Delete(key string) bool {
	c.lock.Lock()
	defer c.lock.Unlock()

	_, ok := c.cache[key]
	delete(c.cache, key)
	return ok
}

// MarshalJSON return a json representation of an entry
func (e Entry) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString("{")

	value := e.value.String()
	value = strings.Replace(value, "\n", "  ", -1)
	value = strings.Replace(value, "\t", " ", -1)

	buffer.WriteString(fmt.Sprintf("\"value\":\"%s\"", value))
	buffer.WriteString(",")
	buffer.WriteString(fmt.Sprintf("\"ttl\":%d", e.ttl))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// MarshalJSON returns a json representation of the cache's contents
func (c *Cache) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.cache)
}
