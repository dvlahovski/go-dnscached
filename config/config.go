package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

const PolicyDefault = "default"
const PolicyKeepMostUsed = "keep-most-used"

type Config struct {
	Server  ServerConfig `json:"Server"`
	Cache   CacheConfig  `json:"Cache"`
	Entries []CacheEntry `json:"CacheEntries"`
}

type ServerConfig struct {
	Address string   `json:"Address"`
	Servers []string `json:"Servers"`
}

type CacheConfig struct {
	MaxEntries    int    `json:"MaxEntries"`
	MinTTL        int    `json:"MinTTL"`
	FlushInterval int    `json:"FlushInterval"`
	Policy        string `json:"Policy"`
}

type CacheEntry struct {
	Key   string `json:"Key"`
	Value net.IP `json:"Value"`
	Type  string `json:"Type"`
	Ttl   int    `json:"Ttl"`
}

func (c *Config) Valid() bool {
	if c.Cache.Policy != PolicyDefault && c.Cache.Policy != PolicyKeepMostUsed {
		return false
	}

	return true
}

func Load() (*Config, error) {
	file, err := os.Open("config/config.json")
	if err != nil {
		log.Printf("error opening config file %s", err)
		return nil, err
	}

	decoder := json.NewDecoder(file)
	config := new(Config)
	err = decoder.Decode(config)
	if err != nil {
		log.Printf("error decoding json config: %s", err)
		return nil, err
	}

	if !config.Valid() {
		log.Printf("invalid config")
		return nil, fmt.Errorf("invalid config")
	}

	return config, nil
}

func (c *Config) Store() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Printf("error opening config file")
		return
	}

	encoder := json.NewEncoder(file)
	err = encoder.Encode(&c)
	if err != nil {
		log.Printf("config save err: %s", err)
	}
}
