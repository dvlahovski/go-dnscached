package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/dvlahovski/go-dnscached/api"
	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/config"
	"github.com/dvlahovski/go-dnscached/server"
)

func main() {
	file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalln("Failed to open log file")
	}
	defer file.Close()

	multi := io.MultiWriter(file, os.Stdout)

	log.SetOutput(multi)
	log.SetPrefix("go-dnscached: ")
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)

	log.Printf("Daemon started")
	defer log.Printf("Daemon shutdown")

	config, err := config.Load()
	if err != nil {
		return
	}
	fmt.Printf("%v\n", config)

	cache := cache.NewCache(*config)

	server, err := server.NewServer(*cache, *config)
	if err != nil {
		log.Printf("server creation error: %s", err.Error())
		return
	}

	// serverErrors := server.ListenAndServe()
    go func() {
        log.Fatal(server.ListenAndServe())
    }()

	go func() {
		log.Fatal(api.Run(server, cache))
	}()

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt, syscall.SIGTERM)

	select {
	case sig := <-sigs:
		log.Printf("Caught signal: %s", sig)
		if err := server.Shutdown(); err != nil {
			log.Printf("Server shutdown error: %s", err.Error())
		}
	// case <-serverErrors:
	// 	log.Printf("Server error")
	}

}
