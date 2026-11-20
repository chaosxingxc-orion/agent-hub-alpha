package main

// Concurrency acceptance for the per-call schema root (external audit GC#26):
// the FLOW entry (ComposeErrors) must stay correct while another document type
// is validated concurrently. This arm lives in cmd/compose — a separate test
// binary — because internal/compose's wiring pin requires the process to start
// with the graph library UNREGISTERED; registering here cannot disturb it.
//
// Mutation anchor: reintroducing a package-level root that only ComposeErrors
// sets (the half-revert that the validate-package arms cannot see, since they
// bypass this entry) makes the flow arm reject valid graphs or drift, and
// -race reports the write/read pair.

import (
	"path/filepath"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/compose"
	"github.com/openjiuwen/agent-festival/go/internal/validate"
)

func auditLoadFixture(t *testing.T, parts ...string) map[string]any {
	t.Helper()
	for _, root := range []string{"../../schema", "../../../schema"} {
		fx, err := validate.LoadFixture(filepath.Join(append([]string{root, "examples"}, parts...)...))
		if err != nil {
			continue
		}
		if fx.ParseError != "" {
			t.Fatalf("%v parse error: %s", parts, fx.ParseError)
		}
		return fx.Doc
	}
	t.Fatalf("fixture %v not found", parts)
	return nil
}

// TestAuditFlowEntryConcurrentWithContract pins: ComposeErrors (the production
// flow entry) and ContractErrors (a different root) may run concurrently over
// unchanged valid input without rejecting anything.
func TestAuditFlowEntryConcurrentWithContract(t *testing.T) {
	compose.RegisterGraph() // production main() does this before any ComposeErrors call
	graph := auditLoadFixture(t, "compose", "valid", "parallel-split-budget-ok.json")
	contract := auditLoadFixture(t, "valid", "doc-extractor.json")

	// Serial baseline: both entries must accept before concurrency is claimed.
	var rejects atomic.Int32
	if r := compose.ComposeErrors(graph); len(r.Errors) > 0 {
		t.Fatalf("serial flow baseline rejected: %v", r.Errors)
	}
	if es := validate.ContractErrors(contract); len(es) > 0 {
		t.Fatalf("serial contract baseline rejected: %v", es)
	}

	const workers, rounds = 8, 30
	var wg sync.WaitGroup
	start := make(chan struct{})
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			<-start
			for n := 0; n < rounds; n++ {
				if i%2 == 0 {
					if r := compose.ComposeErrors(graph); len(r.Errors) > 0 {
						rejects.Add(1)
					}
				} else {
					if es := validate.ContractErrors(contract); len(es) > 0 {
						rejects.Add(1)
					}
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()

	if n := rejects.Load(); n != 0 {
		t.Errorf("unchanged valid documents rejected %d times across the flow/contract entries", n)
	}
}
