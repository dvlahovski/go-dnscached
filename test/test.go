package test

import (
	"net"
	"os"

	"github.com/dvlahovski/go-dnscached/config"
	"github.com/miekg/dns"
)

func GetStubConfig() *config.Config {
	cfg := new(config.Config)

	cfg.Cache.MaxEntries = 1000
	cfg.Cache.MinTTL = 60
	cfg.Cache.FlushInterval = 30
	cfg.Cache.Policy = config.PolicyDefault

	cfg.Server.Address = "0.0.0.0:1234"
	cfg.Server.Servers = make([]string, 1)
	cfg.Server.Servers[0] = "8.8.8.8:53"

	return cfg
}

func GetDnsMsg() *dns.Msg {
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	// msg.SetQuestion("google.bg.", dns.TypeA)

	var err error
	msg.Answer = make([]dns.RR, 1)
	msg.Answer[0], err = dns.NewRR("google.bg. 300 IN A 93.123.23.52")
	if err != nil {
		os.Exit(-1)
	}

	return msg
}

func GetDnsMsgQuestion() *dns.Msg {
	msg := new(dns.Msg)
	msg.Id = dns.Id()
	msg.RecursionDesired = true
	msg.SetQuestion("google.bg.", dns.TypeA)

	return msg
}

type StubResponseWriter struct {
	Msg *dns.Msg
}

func (s *StubResponseWriter) LocalAddr() net.Addr {
	return nil
}
func (s *StubResponseWriter) RemoteAddr() net.Addr {
	return nil
}
func (s *StubResponseWriter) WriteMsg(msg *dns.Msg) error {
	s.Msg = msg
	return nil
}
func (s *StubResponseWriter) Write([]byte) (int, error) {
	return 0, nil
}
func (s *StubResponseWriter) Close() error {
	return nil
}
func (s *StubResponseWriter) TsigStatus() error {
	return nil
}
func (s *StubResponseWriter) TsigTimersOnly(bool) {}
func (s *StubResponseWriter) Hijack()             {}
