package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/server"
	"github.com/miekg/dns"
)

func getParams(req *http.Request, key string) (string, bool) {
	if req.Method != http.MethodGet {
		return "", false
	}

	values := req.URL.Query()
	value := values.Get(key)

	if value == "" {
		return "", false
	}

	return value, true
}

func requiredParam(w http.ResponseWriter, req *http.Request, key string) (string, bool) {
	value, exists := getParams(req, key)

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad Request!"))
		return "", false
	}

	return value, true
}

type API struct {
	server *server.Server
	cache  *cache.Cache
}

func (api *API) cacheList(w http.ResponseWriter, req *http.Request) {
	jsonString, err := json.Marshal(api.cache)
	if err != nil {
		log.Printf("%s", err)
		http.NotFound(w, req)
	}
	fmt.Fprintf(w, string(jsonString))
}

func (api *API) cacheGet(w http.ResponseWriter, req *http.Request) {
	value, exists := getParams(req, "key")

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad Request!"))
		return
	}

	entry, ok := api.cache.GetEntry(value)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad Request!"))
		return
	}

	jsonString, err := json.Marshal(entry)
	if err != nil {
		log.Printf("%s", err)
		http.NotFound(w, req)
	}
	fmt.Fprintf(w, string(jsonString))
}

func (api *API) cacheDelete(w http.ResponseWriter, req *http.Request) {
	value, exists := getParams(req, "key")

	if !exists {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("400 - Bad Request!"))
		return
	}

	status := api.cache.Delete(value)

	if status {
		fmt.Fprintf(w, "Successfully deleted %s", value)
	} else {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("403 - Bad Request!"))
		fmt.Fprintf(w, "\nNo such record %s", value)
	}
}

func (api *API) cacheInsert(w http.ResponseWriter, req *http.Request) {
	var all bool = true
	key, exists := requiredParam(w, req, "key")
	all = all && exists
	recordTypeStr, exists := requiredParam(w, req, "type")
	all = all && exists
	value, exists := requiredParam(w, req, "value")
	all = all && exists
	ttlStr, exists := requiredParam(w, req, "ttl")
	all = all && exists
	if !all {
		return
	}

	var recordType uint16
	if recordTypeStr == "A" {
		recordType = dns.TypeA
	} else if recordTypeStr == "AAAA" {
		recordType = dns.TypeAAAA
	}

	ttl, err := strconv.Atoi(ttlStr)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("403 - Bad Request!"))
		return
	}

	ok := api.cache.InsertFromParams(key, value, recordType, ttl)

	if !ok {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("403 - Bad Request!"))
	} else {
		fmt.Fprintf(w, "Successfully inserted %s", key)
	}
}

func Run(server *server.Server, cache *cache.Cache) {
	api := new(API)
	api.cache = cache
	api.server = server

	mux := http.NewServeMux()
	mux.HandleFunc("/cache/all", api.cacheList)
	mux.HandleFunc("/cache/get", api.cacheGet)
	mux.HandleFunc("/cache/delete", api.cacheDelete)
	mux.HandleFunc("/cache/insert", api.cacheInsert)

	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/" {
			fmt.Fprintf(w, "Welcome to the jungle!")
			return
		}
		http.NotFound(w, req)
	})

	s := &http.Server{Addr: ":8282", Handler: mux, WriteTimeout: 1 * time.Second}
	log.Printf("Starting server on %s", s.Addr)
	log.Fatal(s.ListenAndServe())
}
