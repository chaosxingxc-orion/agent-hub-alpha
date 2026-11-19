package main

import (
	"errors"
	"path/filepath"
	"testing"
)

// These checks qualify the BANK DOUBLE's advertised behavior. Without them,
// a fake that loses its state or creates duplicate effects could make the
// framework scenarios report misleading results.
func TestBankLostReplyRetainsQueryableDurableEffect(t *testing.T) {
	path := filepath.Join(t.TempDir(), "bank.json")
	p := bankPayload{"br-test", "customer-test", 50_000}
	if _, err := (bank{path}).submit(p, true); !errors.Is(err, errLostReply) {
		t.Fatalf("expected injected lost reply, got %v", err)
	}
	// A new provider handle reads the file, not the first handle's memory.
	r, found, err := (bank{path}).query(p.BusinessID)
	if err != nil || !found || r.Payload != p {
		t.Fatalf("persisted receipt unavailable: found=%v receipt=%+v err=%v", found, r, err)
	}
	s, err := (bank{path}).state()
	if err != nil || s.Effects != 1 || s.SubmitCalls != 1 || s.QueryCalls != 1 {
		t.Fatalf("provider metrics do not reflect durable acceptance: %+v err=%v", s, err)
	}
}

func TestBankDeduplicatesSamePayloadAndRejectsChangedPayload(t *testing.T) {
	b := bank{filepath.Join(t.TempDir(), "bank.json")}
	p := bankPayload{"br-test", "customer-test", 50_000}
	first, err := b.submit(p, false)
	if err != nil {
		t.Fatal(err)
	}
	second, err := b.submit(p, false)
	if err != nil || second != first {
		t.Fatalf("provider retry changed receipt: %+v -> %+v err=%v", first, second, err)
	}
	p.CreditUSD = 10_000
	if _, err := b.submit(p, false); err == nil {
		t.Fatal("provider accepted a changed amount under the same business ID")
	}
	s, err := b.state()
	if err != nil || s.Effects != 1 || s.SubmitCalls != 3 || s.Receipts[p.BusinessID] != first {
		t.Fatalf("duplicate/conflict altered the original effect: %+v err=%v", s, err)
	}
}
