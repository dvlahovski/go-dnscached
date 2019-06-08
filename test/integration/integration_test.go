package integration

import (
	"strings"
	"testing"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/server"
	"github.com/dvlahovski/go-dnscached/test"
	"github.com/miekg/dns"
)

func Init(t *testing.T) (*server.Server, *cache.Cache, *test.StubDnsClient) {
	config := test.GetStubConfig()
	cache := cache.NewCache(*config)
	client := new(test.StubDnsClient)

	server, err := server.NewServer(*cache, *config, client)
	if err != nil {
		t.Fatalf("server creation error: %s", err.Error())
	}

	return server, cache, client
}

func compareRR(a, b dns.RR) bool {
	return strings.Replace(a.String(), "\t", "", -1) == strings.Replace(b.String(), "\t", "", -1)
}

func TestCacheMissScenario(t *testing.T) {
	serv, cacheClient, dnsClient := Init(t)
	msg := test.GetDnsMsgQuestion()
	respWriter := new(test.StubResponseWriter)
	dnsClient.SetReply(test.GetDnsMsgAnswer())

	_, ok := cacheClient.Get("google.bg.A.")
	if ok {
		t.Fatalf("message should not be present in cache")
	}

	serv.HandleRequest(respWriter, msg)

	respMsg := respWriter.Msg
	if respMsg.Question[0] != msg.Question[0] {
		t.Fatalf("mismatching question")
	}

	cachedMsg, ok := cacheClient.Get("google.bg.A.")
	if !ok {
		t.Fatalf("message should be present in cache")
	}

	if cachedMsg.Question[0] != msg.Question[0] {
		t.Fatalf("sent and cached messages' questions should be identical")
	}

	if !compareRR(cachedMsg.Answer[0], test.GetDnsMsgAnswer().Answer[0]) {
		t.Fatalf("the cached and received DNS messages should be identical")
	}
}

func TestCacheHitScenario(t *testing.T) {
	serv, cacheClient, _ := Init(t)
	msg := test.GetDnsMsgQuestion()
	respWriter := new(test.StubResponseWriter)

	ok := cacheClient.Insert("google.bg.A.", *test.GetDnsMsgAnswer())
	if !ok {
		t.Fatalf("message insert failed")
	}

	serv.HandleRequest(respWriter, msg)

	respMsg := respWriter.Msg
	if respMsg.Question[0] != msg.Question[0] {
		t.Fatalf("mismatching question")
	}

	if !compareRR(respMsg.Answer[0], test.GetDnsMsgAnswer().Answer[0]) {
		t.Fatalf("response should be equal to the message inserted in the cache")
	}
}
