package main

import (
	"net/http"
	"sync"
	"testing"
	"time"
)

type h03BlockingTransport struct {
	next             http.RoundTripper
	once             sync.Once
	entered, release chan struct{}
}

func (b *h03BlockingTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	b.once.Do(func() { close(b.entered); <-b.release })
	return b.next.RoundTrip(r)
}

func TestH03ExecutionDiscoveryOutsideLockAndOneIntent(t *testing.T) {
	s, _, r, _ := h03ExecutionFixture(t)
	old := discoveryClient
	transport := &h03BlockingTransport{next: http.DefaultTransport, entered: make(chan struct{}), release: make(chan struct{})}
	discoveryClient = &http.Client{Transport: transport, Timeout: 5 * time.Second}
	t.Cleanup(func() { discoveryClient = old })
	in := CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}
	type result struct {
		out map[string]any
		err error
	}
	results := make(chan result, 2)
	go func() { out, err := s.BeginCall(in); results <- result{out, err} }()
	select {
	case <-transport.entered:
	case <-time.After(2 * time.Second):
		t.Fatal("discovery did not reach barrier")
	}
	beat := make(chan error, 1)
	go func() { beat <- s.Heartbeat(r.ID, r.Token) }()
	select {
	case err := <-beat:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(time.Second):
		t.Fatal("slow discovery held global lock")
	}
	go func() { out, err := s.BeginCall(in); results <- result{out, err} }()
	close(transport.release)
	sends := 0
	request := ""
	for i := 0; i < 2; i++ {
		select {
		case got := <-results:
			if got.err != nil {
				t.Fatal(got.err)
			}
			if got.out["action"] == "send" {
				sends++
			}
			c := got.out["call"].(*RemoteCall)
			if request != "" && request != c.RequestID {
				t.Fatal("concurrent request created second identity")
			}
			request = c.RequestID
		case <-time.After(3 * time.Second):
			t.Fatal("concurrent Begin timed out")
		}
	}
	if sends != 1 {
		t.Fatalf("expected exactly one first intent, got %d", sends)
	}
}
