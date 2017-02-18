package cache_test

import (
	"fmt"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

func ExampleNewCache() {
	config, err := config.Load("../config/config.json")
	if err != nil {
		return
	}

	_ = cache.NewCache(*config)
}

func ExampleCache_Insert() {
	config, err := config.Load("../config/config.json")
	if err != nil {
		return
	}

	cache := cache.NewCache(*config)
	ok := cache.Insert("google.bg", dns.Msg{})

	if !ok {
		fmt.Printf("insertion failed")
	}
}

func ExampleCache_InsertFromParams() {
	config, err := config.Load("../config/config.json")
	if err != nil {
		return
	}

	cache := cache.NewCache(*config)
	ok := cache.InsertFromParams("google.bg", "1.2.3.4", dns.TypeA, 120)
	fmt.Printf("%t\n", ok)
	// Output:
	// true
}
