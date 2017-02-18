package server

import (
	"testing"
	"time"

	"github.com/dvlahovski/go-dnscached/cache"
	"github.com/dvlahovski/go-dnscached/test"
)

func getServer(t *testing.T) *Server {
	config := test.GetStubConfig()
	cache := cache.NewCache(*config)

	server, err := NewServer(*cache, *config)
	if err != nil {
		t.Fatalf("server creation error: %s", err.Error())
	}

	return server
}

func TestCreation(t *testing.T) {
	server := getServer(t)
	defer server.Shutdown()
}

func TestListenAndServe(t *testing.T) {
	server := getServer(t)
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
	server := getServer(t)
	go func() {
		errors <- server.ListenAndServe()
	}()
	defer server.Shutdown()

	time.Sleep(3 * time.Second)

	errors2 := make(chan error)
	server2 := getServer(t)
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

// func TestHandleRequest(t *testing.T) {
// 	server := getServer(t)
// 	msg := test.GetDnsMsgQuestion()
// 	respWriter := new(test.StubResponseWriter)
// 	server.handleRequest(respWriter, msg)

// 	respMsg := respWriter.Msg
// 	fmt.Printf("%v", respMsg)
// 	fmt.Printf("%v", msg)
// 	if respMsg.Question[0] != msg.Question[0] {
// 		t.Errorf("mismatching question")
// 	}

// }
