package compose

// Concurrency regression anchor for the per-call schema root.
//
// The validator used to keep the current root document in a package-level
// mutable variable; concurrent validations of different document types
// overwrote each other's root, so $defs resolution read the wrong document.
// Observed: DATA RACE plus tens of spurious rejections of unchanged valid
// input. The root is now an explicit per-call parameter, so this test's
// assertions are structural, not scheduling-dependent: a valid document must
// never be rejected because another goroutine validated something else.
//
// Reproduced from the external audit report (GC#26 / GH#53); adapted to this
// package's helpers only. Mutation anchor: reintroducing any package-level
// root that the validator reads makes this test fail under -race.
//
// The arms call the validate entry points directly rather than ComposeErrors:
// registering the graph library would flip the process-wide registration state
// that wiring_test.go's fail-closed pin requires to be unregistered first.
// The walk under test is the same one ComposeErrors drives.

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/validate"
)

func auditLoad(t *testing.T, parts ...string) map[string]any {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(append([]string{ExamplesDir}, parts...)...))
	if err != nil {
		t.Fatal(err)
	}
	var d map[string]any
	if err := json.Unmarshal(b, &d); err != nil {
		t.Fatal(err)
	}
	for k := range d {
		if len(k) > 0 && k[0] == '_' {
			delete(d, k)
		}
	}
	return d
}

// TestAuditConcurrentValidationRoots pins: an immutable contract/graph stays
// valid under concurrent validation of the other document type.
func TestAuditConcurrentValidationRoots(t *testing.T) {
	contract := auditLoad(t, "valid", "doc-extractor.json")
	graph := auditLoad(t, "compose", "valid", "parallel-split-budget-ok.json")
	if graphSchema == nil {
		t.Fatal("graphSchema not loaded — the package init could not find schema/")
	}

	// Serial baseline first: the concurrency arm may only claim what the
	// single-goroutine arm already proves (a fixture/registration artefact
	// must not be read as a race). Counted, not fatal, so one bar covers both
	// arms and a pre-fix run still reports its concurrency numbers.
	var rejects atomic.Int32
	if es := validate.ContractErrors(contract); len(es) > 0 {
		t.Logf("serial contract baseline rejected: %v", es)
		rejects.Add(1)
	}
	var graphErrs []string
	validate.SchemaErrorsDoc(graph, graphSchema, "<root>", &graphErrs)
	if len(graphErrs) > 0 {
		t.Logf("serial graph baseline rejected: %v", graphErrs)
		rejects.Add(1)
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
					if es := validate.ContractErrors(contract); len(es) > 0 {
						rejects.Add(1)
					}
				} else {
					var es []string
					validate.SchemaErrorsDoc(graph, graphSchema, "<root>", &es)
					if len(es) > 0 {
						rejects.Add(1)
					}
				}
			}
		}(worker)
	}
	close(start)
	wg.Wait()

	// The count is scheduling-dependent; the bar is "zero", never a number.
	if n := rejects.Load(); n != 0 {
		t.Errorf("unchanged valid documents rejected %d times (serial baseline + concurrent calls)", n)
	}
}

// TestAuditConcurrentEnvelopeRootIsolation pins the third entry: envelope
// validation runs against EnvelopeSchema, a root distinct from both the
// contract and the graph schema. A signed envelope is out of scope here, so
// the assertion is stability rather than emptiness: the identical input must
// produce the identical verdict set every round. Root cross-talk shows up as
// $defs-resolution errors appearing only under concurrency.
func TestAuditConcurrentEnvelopeRootIsolation(t *testing.T) {
	contract := auditLoad(t, "valid", "doc-extractor.json")
	doc := map[string]any{
		"envelope_version": "0.1.0",
		"contract":         contract,
		"contract_hash":    "sha256:" + strings.Repeat("0", 64),
		"signatures":       []any{},
	}
	verdict := func() string {
		errs, _ := validate.EnvelopeErrors(doc, &validate.TrustStore{}, false)
		sort.Strings(errs)
		return strings.Join(errs, "\n")
	}
	want := verdict()
	if want == "" {
		t.Fatal("envelope baseline unexpectedly clean — the stability arm needs a non-empty verdict")
	}
	var wg sync.WaitGroup
	var drift atomic.Int32
	const workers, rounds = 8, 30
	start := make(chan struct{})
	for worker := 0; worker < workers; worker++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			<-start
			for n := 0; n < rounds; n++ {
				if got := verdict(); got != want {
					drift.Add(1)
				}
			}
		}()
	}
	close(start)
	wg.Wait()
	if n := drift.Load(); n != 0 {
		t.Errorf("envelope verdict drifted %d times under concurrent validation (root cross-talk)", n)
	}
}
