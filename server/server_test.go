package server

import (
	"testing"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/test"
)

func GetServer(t *testing.T) *Server {
	config := test.GetStubConfig()
	cache := cache.NewCache(*config)
	client := new(test.StubDnsClient)

	server, err := NewServer(*cache, *config, client)
	if err != nil {
		t.Fatalf("server creation error: %s", err.Error())
	}

	return server
}

func TestCreation(t *testing.T) {
	server := GetServer(t)
	defer server.Shutdown()
}

func TestListenAndServe(t *testing.T) {
	server := GetServer(t)
	defer server.Shutdown()

	errors := make(chan error)
	go func() {
		errors <- server.ListenAndServe()
	}()

	select {
	case <-errors:
		t.Fatalf("ListenAndServe fail")
	case <-time.After(2 * time.Second):
	}
}

func TestListenAndServeFail(t *testing.T) {
	errors := make(chan error)
	server := GetServer(t)
	go func() {
		errors <- server.ListenAndServe()
	}()
	defer server.Shutdown()

	time.Sleep(3 * time.Second)

	errors2 := make(chan error)
	server2 := GetServer(t)
	go func() {
		errors2 <- server.ListenAndServe()
	}()
	defer server2.Shutdown()

	select {
	case <-errors2:
	case <-errors:
		t.Fatalf("ListenAndServe should fail")
	case <-time.After(3 * time.Second):
		t.Fatalf("ListenAndServe should fail")
	}
}

func TestHandleRequest(t *testing.T) {
	server := GetServer(t)
	msg := test.GetDnsMsgQuestion()
	respWriter := new(test.StubResponseWriter)
	server.HandleRequest(respWriter, msg)

	respMsg := respWriter.Msg
	if respMsg.Question[0] != msg.Question[0] {
		t.Errorf("mismatching question")
	}
}
