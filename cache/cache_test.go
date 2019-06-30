package cache

import (
	"testing"
	"time"

	"github.com/dvlahovski/go-dnscached/test"
	"github.com/miekg/dns"
)

func TestCreation(t *testing.T) {
	config := test.GetStubConfig()
	cache := NewCache(*config)

	if cache.capacity != config.Cache.MaxEntries {
		t.Fail()
	}
}

func TestInsertionSucceeds(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}
}

func TestInsertionFailsOnEmptyAnswers(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	// overwrite answer to be empty
	msg.Answer = msg.Answer[:0]
	ok = cache.Insert("google.bg", *msg)
	if ok {
		t.Fatal("insertion shoud have failed")
	}
}

func TestInsertionOverCapacity(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	msg := test.GetDnsMsgAnswer()

	config.Cache.MaxEntries = 1
	cache := NewCache(*config)

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	ok = cache.Insert("dir.bg", *msg)
	if ok {
		t.Fatal("insertion shoud have failed")
	}
}

func TestInsertionTwice(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	ok = cache.Insert("google.bg", *msg)
	if ok {
		t.Fatal("insertion shoud have failed")
	}
}

func TestInsertionFromParams(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	ok = cache.InsertFromParams("google.bg", "1.2.3.4", dns.TypeA, 120)
	if !ok {
		t.Fatal("insert failed")
	}

	ok = cache.InsertFromParams("google.bg", "2a00:1450:4017:805::2003", dns.TypeAAAA, 240)
	if !ok {
		t.Fatal("insert failed")
	}
}

func TestGetExisting(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	_, ok = cache.Get("google.bg")
	if !ok {
		t.Fatal("get failed")
	}
}

func TestGetNonExisting(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)

	_, ok = cache.Get("google.bg")
	if ok {
		t.Fatal("get should fail")
	}
}

func TestGetEntry(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	now := time.Now().Unix()
	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	entry, ok := cache.GetEntry("google.bg")
	if !ok {
		t.Fatal("get entry failed")
	}

	if entry.ttl != int(msg.Answer[0].Header().Ttl)+int(now) {
		t.Fatalf("expected %d ttl, got %d", int(msg.Answer[0].Header().Ttl)+int(now), entry.ttl)
	}

	if entry.hits != 0 {
		t.Fatal("hits should be 0")
	}

	_, ok = cache.Get("google.bg")
	if !ok {
		t.Fatal("get failed")
	}

	entry, ok = cache.GetEntry("google.bg")
	if !ok {
		t.Fatal("get entry failed")
	}

	if entry.hits != 1 {
		t.Fatal("hits should be 1")
	}
}

func TestDelete(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	ok = cache.Delete("google.bg")
	if !ok {
		t.Fatal("delete failed")
	}

	ok = cache.Delete("google.bg")
	if ok {
		t.Fatal("delete should return false")
	}

	_, ok = cache.Get("google.bg")
	if ok {
		t.Fatal("get should return false")
	}
}

func TestFlush(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	config.Cache.FlushInterval = 1
	config.Cache.MinTTL = 0
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()
	msg.Answer[0].Header().Ttl = 1

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	_, ok = cache.Get("google.bg")
	if !ok {
		t.Fatal("get failed")
	}

	time.Sleep(2 * time.Second)

	_, ok = cache.Get("google.bg")
	if ok {
		t.Fatal("get should fail")
	}
}

func TestDoesNotFlushZeroTtl(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	config.Cache.FlushInterval = 1
	config.Cache.MinTTL = 0
	cache := NewCache(*config)
	msg := test.GetDnsMsgAnswer()
	msg.Answer[0].Header().Ttl = 0

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	_, ok = cache.Get("google.bg")
	if !ok {
		t.Fatal("get failed")
	}

	time.Sleep(2 * time.Second)

	_, ok = cache.Get("google.bg")
	if !ok {
		t.Fatal("get failed")
	}
}

func TestToStringEntry(t *testing.T) {
	var entry Entry
	entry.ttl = 101
	entry.Value = *test.GetDnsMsgAnswer()

	strEntry := entry.ToStringEntry()
	if strEntry.Key != "google.bg" {
		t.Fatal("Key is incorrect")
	}

	if strEntry.Value[0] != "1.2.3.4" {
		t.Fatal("IP is incorrect")
	}

	if strEntry.Ttl != 101 {
		t.Fatal("TTL is incorrect")
	}

	if strEntry.Type != "A" {
		t.Fatal("Type is incorrect")
	}
}
