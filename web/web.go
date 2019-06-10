package web

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
)

// Web insance
type WEB struct {
}

type Page struct {
	CacheEntries []cache.StringEntry
}

func (web *WEB) getCacheEntries() ([]cache.StringEntry, error) {
	res, err := http.Get("http://localhost:8282/cache/all")
	if err != nil {
		return nil, err
	}
	contents, err := ioutil.ReadAll(res.Body)
	res.Body.Close()
	if err != nil {
		return nil, err
	}

	var cacheEntries []cache.StringEntry
	err = json.Unmarshal(contents, &cacheEntries)
	if err != nil {
		return nil, err
	}

	return cacheEntries, nil
}

func handleError(err error, w http.ResponseWriter) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte("500 - Internal Server Error!"))
	w.Write([]byte(err.Error()))
}

// TODO
func (web *WEB) index(w http.ResponseWriter, req *http.Request) {
	cacheEntries, err := web.getCacheEntries()
	if err != nil {
		handleError(err, w)
		return
	}

	p := &Page{
		CacheEntries: cacheEntries,
	}

	t, err := template.New("index.html").Funcs(template.FuncMap{
		"toHumanTime": func(timestamp int) string {
			if timestamp == 0 {
				return "∞"
			}
			return time.Unix(int64(timestamp), 0).Format("15:04:05 02.01.2006")
		},
		"getKey": func(addr string, recordType string) string {
			return fmt.Sprintf("%s.%s.", addr, recordType)
		},
	}).ParseFiles("web/static/index.html", "web/static/template.html")
	if err != nil {
		handleError(err, w)
		return
	}

	if err := t.Execute(w, p); err != nil {
		log.Printf("HTML template parsing failed: %s", err)
	}
}

// Run the Web HTTP server
func Run() error {
	web := new(WEB)

	mux := http.NewServeMux()
	mux.HandleFunc("/", web.index)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web/static"))))

	// mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
	//     http.NotFound(w, req)
	// })

	s := &http.Server{Addr: ":8080", Handler: mux, WriteTimeout: 1 * time.Second}
	log.Printf("Starting Web GUI server on %s", s.Addr)
	return s.ListenAndServe()
}
