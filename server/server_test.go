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

	serverErrors := server.ListenAndServe()

	select {
	case <-serverErrors:
		t.Fatalf("ListenAndServe fail")
	case <-time.After(2 * time.Second):
	}
}

func TestListenAndServeFail(t *testing.T) {
	server := getServer(t)
	serverErrors := server.ListenAndServe()
	defer server.Shutdown()

	time.Sleep(1 * time.Second)

	server2 := getServer(t)
	serverErrors2 := server2.ListenAndServe()
	defer server2.Shutdown()

	select {
	case <-serverErrors2:
	case <-serverErrors:
		t.Fatalf("ListenAndServe should fail")
	case <-time.After(1 * time.Second):
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
