package server

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"net"
	"net/http"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

type DnsClient interface {
	Exchange(m *dns.Msg, address string) (*dns.Msg, time.Duration, error)
}

type HttpClient interface {
	Post(url, contentType string, body io.Reader) (resp *http.Response, err error)
}

// servers is the list of DNS servers that we forward to/ask
type Server struct {
	server       *dns.Server
	cache        *cache.Cache
	servers      []net.UDPAddr
	serversHttps []string
	dnsClient    DnsClient
	httpClient   HttpClient
}

// Get a new server ready to start serving
func NewServer(cache *cache.Cache, config *config.Config, dnsClient DnsClient, httpClient HttpClient) (*Server, error) {
	s := new(Server)
	addr, err := net.ResolveUDPAddr("udp", config.Server.Address)
	if err != nil {
		return nil, err
	}

	log.Printf("Server listening at %s", addr.String())
	s.server = &dns.Server{Addr: addr.String(), Net: "udp"}

	if len(config.Server.Servers) <= 0 {
		return nil, fmt.Errorf("no dns servers to use")
	}

	s.servers = make([]net.UDPAddr, len(config.Server.Servers))
	for i, addr := range config.Server.Servers {
		udpAddr, err := net.ResolveUDPAddr("udp", addr)
		if err != nil {
			return nil, err
		}

		s.servers[i] = *udpAddr
	}

	s.serversHttps = make([]string, len(config.Server.ServersHTTPS))
	copy(s.serversHttps, config.Server.ServersHTTPS)

	s.cache = cache
	s.dnsClient = dnsClient
	s.httpClient = httpClient

	return s, nil
}

// Shutdown gracefully
func (s *Server) Shutdown() error {
	return s.server.Shutdown()
}

// Get a random DNS server to query.
func (s *Server) getRandServer() string {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(s.servers)
	return s.servers[n].String()
}

// Get a random DNS server to query over HTTPs.
func (s *Server) getRandServerHttps() string {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(s.serversHttps)
	return s.serversHttps[n]
}

func (s *Server) makeDNSoverHTTPSrequest(url string, dnsMsg *dns.Msg) (*dns.Msg, error) {
	rawDns, err := dnsMsg.Pack()
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Post(url, "application/dns-message", bytes.NewBuffer(rawDns))
	if err != nil {
		return nil, err
	}

	contents, err := ioutil.ReadAll(resp.Body)
	resp.Body.Close()
	if err != nil {
		return nil, err
	}

	answer := new(dns.Msg)
	err = answer.Unpack(contents)
	if err != nil {
		return nil, err
	}

	return answer, nil
}

func (s *Server) callFirstSuccessfulServer(request *dns.Msg) (serverResponse *dns.Msg, err error) {
	for _, serverUrl := range s.serversHttps {
		serverResponse, err = s.makeDNSoverHTTPSrequest(serverUrl, request)
		if err != nil {
			log.Printf("server: %s, failed with: %s", serverUrl, err)
		} else {
			return
		}
	}

	for _, serverAddr := range s.servers {
		serverResponse, _, err = s.dnsClient.Exchange(request, serverAddr.String())
		if err != nil {
			log.Printf("server: %s, failed with: %s", serverAddr.String(), err)
		} else {
			return
		}
	}

	return
}

// Make a DNS request to a server
func (s *Server) makeRequest(questions []dns.Question) (dns.Msg, bool) {
	request := new(dns.Msg)
	request.Id = dns.Id()
	request.RecursionDesired = true
	request.Question = make([]dns.Question, len(questions))
	copy(request.Question, questions)

	serverResponse, err := s.callFirstSuccessfulServer(request)

	if err != nil {
		log.Printf("%s\n", err.Error())
		return dns.Msg{}, false
	}

	if serverResponse == nil {
		return dns.Msg{}, false
	}

	return *serverResponse, true
}

// If something went wrong - inform the client
func (s *Server) shouldSendErrorResponse(response dns.Msg, status bool) int {
	if !status {
		return dns.RcodeServerFailure
	}

	if response.Rcode != dns.RcodeSuccess {
		return response.Rcode
	}

	return dns.RcodeSuccess
}

// Act as a forwarding server without caching
// This is in the case where the query is not of type A or AAAA
func (s *Server) passThrough(dnsWriter dns.ResponseWriter, clientRequest *dns.Msg) {
	serverResponse, ok := s.makeRequest(clientRequest.Question)

	reply := new(dns.Msg)

	rcode := s.shouldSendErrorResponse(serverResponse, ok)
	if rcode != dns.RcodeSuccess {
		reply.SetRcode(clientRequest, rcode)
		dnsWriter.WriteMsg(reply)
		return
	}

	reply.SetReply(clientRequest)

	reply.Answer = make([]dns.RR, len(serverResponse.Answer))
	copy(reply.Answer, serverResponse.Answer)

	reply.Ns = make([]dns.RR, len(serverResponse.Ns))
	copy(reply.Ns, serverResponse.Ns)

	reply.Extra = make([]dns.RR, len(serverResponse.Extra))
	copy(reply.Extra, serverResponse.Extra)

	dnsWriter.WriteMsg(reply)
}

// Handle a client request
// Check if there is a cache record and return it or create it
// Ask one of the DNS servers if the record is not in the cache
func (s *Server) HandleRequest(dnsWriter dns.ResponseWriter, clientRequest *dns.Msg) {
	if (len(clientRequest.Question)) != 1 {
		s.passThrough(dnsWriter, clientRequest)
		return
	}

	if clientRequest.Question[0].Qtype != dns.TypeA && clientRequest.Question[0].Qtype != dns.TypeAAAA {
		s.passThrough(dnsWriter, clientRequest)
		return
	}

	question := clientRequest.Question[0].Name
	if clientRequest.Question[0].Qtype == dns.TypeA {
		question += "A"
	} else {
		question += "AAAA"
	}

	cachedMsg, hit := s.cache.Get(dns.Fqdn(question))

	reply := new(dns.Msg)
	response := dns.Msg{}

	if hit {
		response = cachedMsg
	} else {
		var ok bool
		response, ok = s.makeRequest(clientRequest.Question)

		rcode := s.shouldSendErrorResponse(response, ok)
		if rcode != dns.RcodeSuccess {
			reply.SetRcode(clientRequest, rcode)
			dnsWriter.WriteMsg(reply)
			return
		}

		s.cache.Insert(dns.Fqdn(question), response)
	}

	reply.SetReply(clientRequest)

	reply.Answer = make([]dns.RR, len(response.Answer))
	copy(reply.Answer, response.Answer)

	reply.Ns = make([]dns.RR, len(response.Ns))
	copy(reply.Ns, response.Ns)

	reply.Extra = make([]dns.RR, len(response.Extra))
	copy(reply.Extra, response.Extra)

	dnsWriter.WriteMsg(reply)
}

// Start the server
func (s *Server) ListenAndServe() error {
	dns.HandleFunc(".", s.HandleRequest)

	if err := s.server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}
