package server

import (
	"fmt"
	"log"
	"reflect"

	"github.com/miekg/dns"
)

type Server struct {
	server *dns.Server
	// inErrors chan struct {}
	outErrors chan struct{}
}

func NewServer() *Server {
	s := new(Server)
	s.server = &dns.Server{Addr: ":3333", Net: "udp"}
	// s.inErrors = make(chan struct{})
	s.outErrors = make(chan struct{})

	return s
}

func (s *Server) Shutdown() error {
	return s.server.Shutdown()
}

func passThrough(dnsWriter dns.ResponseWriter, clientRequest *dns.Msg) {
	request := new(dns.Msg)
	request.Id = dns.Id()
	request.RecursionDesired = true
	request.Question = make([]dns.Question, len(clientRequest.Question))
	copy(request.Question, clientRequest.Question)

	client := new(dns.Client)
	serverResponse, _, err := client.Exchange(request, "95.87.194.5:53")

	reply := new(dns.Msg)

	if err != nil || serverResponse == nil {
		log.Printf("%s\n", err.Error())
		reply.SetRcode(clientRequest, dns.RcodeServerFailure)
		dnsWriter.WriteMsg(reply)
		return
	}

	if serverResponse.Rcode != dns.RcodeSuccess {
		reply.SetRcode(clientRequest, serverResponse.Rcode)
		dnsWriter.WriteMsg(reply)
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

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	// fmt.Printf("%+v\n", r)
	// fmt.Printf("%d\n", r.Question[0].Qtype)

	// TODO skip if not A or AAAA
	// TODO skip if question count != 1
	if (len(r.Question)) != 1 {

	}

	m := new(dns.Msg)
	m.Id = dns.Id()
	m.RecursionDesired = true
	m.Question = make([]dns.Question, len(r.Question))
	copy(m.Question, r.Question)

	c := new(dns.Client)
	in, _, err := c.Exchange(m, "95.87.194.5:53")

	// fmt.Printf("%+V\n", in)
	fmt.Printf("sizeof %d\n", reflect.TypeOf(in.Answer).Size())

	if err != nil {
		fmt.Printf("Fail: %s\n", err.Error())
	}

	// fmt.Printf("--------------\n")
	// fmt.Printf("%+v\n", in)
	// fmt.Printf("--------------\n")
	// fmt.Printf("%+V\n", in.Answer[0])
	// fmt.Printf("--------------\n")
	// if t, ok := in.Answer[0].(*dns.A); ok {
	//     fmt.Printf("asdasdsad %s\n", t.A)
	// } else {
	//     fmt.Printf("Fail: %+V\n", ok)
	// }
}

func (s *Server) ListenAndServe() chan struct{} {
	dns.HandleFunc(".", passThrough)

	go func() {
		if err := s.server.ListenAndServe(); err != nil {
			log.Printf("%s\n", err.Error())
			s.outErrors <- struct{}{}
		}

		defer s.server.Shutdown()
	}()

	// go func () {
	//     <- s.inErrors
	//     log.Printf("here")
	//     s.server.Shutdown()
	// }()

	return s.outErrors
}

// TODO
// func SetupLogging() {
//     file, err := os.OpenFile("log.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
//     if err != nil {
//         log.Fatalln("Failed to open log file", output, ":", err)
//     }

//     MyFile = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
// }
