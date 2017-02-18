# go-dnscached

This is a simple DNS caching server.

It uses this [DNS library](https://github.com/miekg/dns) for sending/receiving DNS queries.

Currently it caches only A and AAAA queries.
If it receives a different query - it acts as a forwarding DNS server without caching.

There is a json config file in `config/config.json`

By default it creates a log file and also logs to STDOUT

After installing, the server is run with `go run main.go` or with `go build main.go; ./main`
