package server

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

type Server struct {
	server    *dns.Server
	outErrors chan struct{}
	cache     cache.Cache
	servers   []net.UDPAddr
}

func NewServer(cache cache.Cache, config config.Config) (*Server, error) {
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

	s.outErrors = make(chan struct{})
	s.cache = cache

	return s, nil
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown()
}

func (s *Server) getRandServer() string {
	rand.Seed(time.Now().Unix())
	n := rand.Int() % len(s.servers)
	return s.servers[n].String()
}

func (s *Server) makeRequest(questions []dns.Question) (dns.Msg, bool) {
	request := new(dns.Msg)
	// request.SetEdns0(4096, true)
	request.Id = dns.Id()
	request.RecursionDesired = true
	request.Question = make([]dns.Question, len(questions))
	copy(request.Question, questions)

	client := new(dns.Client)
	serverAddr := s.getRandServer()
	serverResponse, _, err := client.Exchange(request, serverAddr)

	if err != nil {
		log.Printf("%s\n", err.Error())
		return dns.Msg{}, false
	}

	if serverResponse == nil {
		return dns.Msg{}, false
	}

	return *serverResponse, true
}

func (s *Server) shouldSendErrorResponse(response dns.Msg, status bool) int {
	if !status {
		return dns.RcodeServerFailure
	}

	if response.Rcode != dns.RcodeSuccess {
		return response.Rcode
	}

	return dns.RcodeSuccess
}

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

func (s *Server) handleRequest(dnsWriter dns.ResponseWriter, clientRequest *dns.Msg) {
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

	cachedMsg, hit := s.cache.Get(question)

	reply := new(dns.Msg)
	response := dns.Msg{}

	if hit {
		response = cachedMsg
	} else {
		var ok = false
		response, ok = s.makeRequest(clientRequest.Question)

		rcode := s.shouldSendErrorResponse(response, ok)
		if rcode != dns.RcodeSuccess {
			reply.SetRcode(clientRequest, rcode)
			dnsWriter.WriteMsg(reply)
			return
		}

		s.cache.Insert(question, response)
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

func (s *Server) ListenAndServe() chan struct{} {
	dns.HandleFunc(".", s.handleRequest)

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Printf("%s\n", err.Error())
			s.outErrors <- struct{}{}
		}

		defer s.server.Shutdown()
	}()

	return s.outErrors
}
