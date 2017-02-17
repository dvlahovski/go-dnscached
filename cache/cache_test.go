package cache

import (
	"testing"
	"time"

	"github.com/dvlahovski/go-dnscached/test"
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
	msg := test.GetDnsMsg()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}
}

func TestInsertionFailsOnEmptyAnswers(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsg()

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
	msg := test.GetDnsMsg()

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
	msg := test.GetDnsMsg()

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	ok = cache.Insert("google.bg", *msg)
	if ok {
		t.Fatal("insertion shoud have failed")
	}
}

func TestGetExisting(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	cache := NewCache(*config)
	msg := test.GetDnsMsg()

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

func TestFlush(t *testing.T) {
	var ok bool
	config := test.GetStubConfig()
	config.Cache.FlushInterval = 1
	cache := NewCache(*config)
	msg := test.GetDnsMsg()
	msg.Answer[0].Header().Ttl = 1

	ok = cache.Insert("google.bg", *msg)
	if !ok {
		t.Fatal("insertion failed")
	}

	_, ok = cache.Get("google.bg")
	if ok != true {
		t.Fatal("get failed")
	}

	time.Sleep(2 * time.Second)

	_, ok = cache.Get("google.bg")
	if ok {
		t.Fatal("get should fail")
	}
}
