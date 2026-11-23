package main

import (
	"crypto/sha256"
	"encoding/hex"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/rfc8785"
)

func TestH03DigestOriginalNULAndPrivateRawAreDistinct(t *testing.T) {
	value := map[string]any{"input": "typed input", "goal": "frozen task", "binding": BusBinding{AgentID: "role", Version: "0.1.0"}, "model": "engineering-no-model"}
	raw, err := rfc8785.Canonicalize(jsonDoc(value))
	if err != nil {
		t.Fatal(err)
	}
	withNUL := sha256.Sum256(append(append([]byte{}, raw...), 0))
	expected := "sha256:" + hex.EncodeToString(withNUL[:])
	actual, err := canonicalHash(value)
	if err != nil || actual != expected || actual == h03RawHash(value) {
		t.Fatal("canonicalHash must retain original trailing NUL", actual, expected, err)
	}
	if got := sha("R1", "P1-v1", "sum", "iteration:0"); got != "sha256:592757eb3061ad13c9b3515dbe55103f9ee2b5cb747a9d0fe4cd263e5a677675" {
		t.Fatal("node ID algorithm changed", got)
	}
}

func TestH03BindingNilPreservesOriginalPlanHash(t *testing.T) {
	p := testPlan(goodSteps()...)
	a := planHash(p, map[string]string{})
	p.ExecutionBinding = nil
	if b := planHash(p, map[string]string{}); a != b {
		t.Fatal("nil profile changed legacy digest")
	}
}
