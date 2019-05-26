package config

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
)

// PolicyDefault is the default caching policy
const PolicyDefault = "default"

// PolicyKeepMostUsed TODO
const PolicyKeepMostUsed = "keep-most-used"

// Config is the layout struct of the JSON config
type Config struct {
	Server  ServerConfig `json:"Server"`
	Cache   CacheConfig  `json:"Cache"`
	Entries []CacheEntry `json:"CacheEntries"`
}

// ServerConfig is the server specific configuration
type ServerConfig struct {
	Address string   `json:"Address"`
	Servers []string `json:"Servers"`
	ServersHTTPS []string `json:"ServersHTTPS"`
}

// CacheConfig is the cache specific configuration
type CacheConfig struct {
	MaxEntries    int    `json:"MaxEntries"`
	MinTTL        uint32 `json:"MinTTL"`
	FlushInterval int    `json:"FlushInterval"`
	Policy        string `json:"Policy"`
}

// CacheEntry is the entry layout of the cache prefill entries in the config
type CacheEntry struct {
	Key   string `json:"Key"`
	Value net.IP `json:"Value"`
	Type  string `json:"Type"`
	Ttl   int    `json:"Ttl"`
}

// Valid checks if the loaded config is valid
func (c *Config) Valid() bool {
	if c.Cache.Policy != PolicyDefault && c.Cache.Policy != PolicyKeepMostUsed {
		return false
	}

	return true
}

// Load the contents of the JSON config file and make some validations
func Load(config_path string) (*Config, error) {
	file, err := os.Open(config_path)
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

// Store the config obj in the json file
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
