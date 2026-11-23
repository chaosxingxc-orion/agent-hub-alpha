package e2e

// W6 (B0-RV-04): the reconciliation read face answers only facts this host
// can observe — result-record presence, machine state, ledger torn-tail.
// It has no serve/conflict/re-execute/restart field: recovery decides with
// facts the host does not own (current execution authority, remote task,
// business receipt), and unknown means query or wait, never restart.
//
// RED-BEFORE on the old API (recorded in the W6 commit): the old
// queryForReconciliation returned the SAME (false,false) pair for
// "never adopted" and "dispatched with no result" — two world states that
// demand different recovery actions.

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
)

func adoptOnly(t *testing.T, h *t2Host, op string) {
	t.Helper()
	if o := h.machine.Adopt(requestForT2(op, "er-1", "ev-1", t2Digest('a'), "br-"+op)); o.Status != "adopted" {
		t.Fatalf("adopt %s: %q: %s", op, o.Status, o.Detail)
	}
}

func dispatchOnly(t *testing.T, h *t2Host, op string) {
	t.Helper()
	if o := h.machine.Dispatch(op, chainGrants()); o.Status != "dispatched" {
		t.Fatalf("dispatch %s: %q: %s", op, o.Status, o.Detail)
	}
}

func putResultDigest(t *testing.T, h *t2Host, op, digest string) {
	t.Helper()
	if err := h.store.put(op, storedResult{
		Receipt:       "r-" + op,
		Facts:         []string{"fact-" + op},
		BusinessReqID: "br-" + op,
		EvaluationRev: "er-1",
		CmdDigest:     digest,
	}); err != nil {
		t.Fatalf("put %s: %v", op, err)
	}
}

func putResultOnly(t *testing.T, h *t2Host, op string) {
	t.Helper()
	putResultDigest(t, h, op, t2Digest('a'))
}

// TestT2Store_GetIsStorageOnly pins the store-level fact: presence and the
// record, nothing else. The store does not know whether the operation was
// completed (commitOutcome puts before MarkCompleted), so no "committed"
// verdict may be attributed to it.
func TestT2Store_GetIsStorageOnly(t *testing.T) {
	rs, err := newResultStore(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if r, ok := rs.get("op-none"); ok || r.Receipt != "" {
		t.Fatalf("missing op = (%+v,%v), want zero record and false", r, ok)
	}
	if err := rs.put("op-one", storedResult{
		Receipt: "r-one", Facts: []string{"f"}, BusinessReqID: "br-one", EvaluationRev: "er-1", CmdDigest: t2Digest('b'),
	}); err != nil {
		t.Fatal(err)
	}
	r, ok := rs.get("op-one")
	if !ok || r.Receipt != "r-one" || r.BusinessReqID != "br-one" || r.CmdDigest != t2Digest('b') {
		t.Fatalf("stored record = (%+v,%v), want the put payload", r, ok)
	}
}

// TestT2Host_InspectFactsStateMatrix pins the facts the host can report in
// every reachable local state. The dispatched/with-result row is the B/C
// crash window: the store has a payload while the machine has not recorded
// completion — the state the old read face reported as a committed/serve
// result.
func TestT2Host_InspectFactsStateMatrix(t *testing.T) {
	cases := []struct {
		name         string
		setup        func(t *testing.T, h *t2Host, op string)
		wantRecorded bool
		wantState    adopt.State
		wantDigest   string
	}{
		{
			name:         "no local record",
			setup:        func(t *testing.T, h *t2Host, op string) {},
			wantRecorded: false,
			wantState:    "",
		},
		{
			name: "adopted",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
			},
			wantRecorded: false,
			wantState:    adopt.StateAdopted,
		},
		{
			name: "dispatched without result",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
			},
			wantRecorded: false,
			wantState:    adopt.StateDispatched,
		},
		{
			name: "dispatched with result (crash window)",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				putResultOnly(t, h, op)
			},
			wantRecorded: true,
			wantState:    adopt.StateDispatched,
			wantDigest:   t2Digest('a'),
		},
		{
			name: "stored result without machine record",
			setup: func(t *testing.T, h *t2Host, op string) {
				// Store-only state: the payload survives, the ledger does
				// not. A SECOND digest proves StoredCmdDigest comes from
				// the record (a constant 'a' mutation is red).
				putResultDigest(t, h, op, t2Digest('b'))
			},
			wantRecorded: true,
			wantState:    "",
			wantDigest:   t2Digest('b'),
		},
		{
			name: "unresolved",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				if o := h.machine.MarkUnresolved(op); o.Status != string(adopt.StateUnresolved) {
					t.Fatalf("unresolve: %q: %s", o.Status, o.Detail)
				}
			},
			wantRecorded: false,
			wantState:    adopt.StateUnresolved,
		},
		{
			name: "failed",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				if o := h.machine.MarkFailed(op); o.Status != string(adopt.StateFailed) {
					t.Fatalf("fail: %q: %s", o.Status, o.Detail)
				}
			},
			wantRecorded: false,
			wantState:    adopt.StateFailed,
		},
		{
			name: "completed with result",
			setup: func(t *testing.T, h *t2Host, op string) {
				if o := h.adoptAndDispatch(op, "er-1", "ev-1", t2Digest('a'), "br-"+op, "r-"+op, []string{"f"}); o.Status != "dispatched" {
					t.Fatalf("setup dispatch: %q: %s", o.Status, o.Detail)
				}
				if o := h.machine.MarkCompleted(op); o.Status != string(adopt.StateCompleted) {
					t.Fatalf("complete: %q: %s", o.Status, o.Detail)
				}
			},
			wantRecorded: true,
			wantState:    adopt.StateCompleted,
			wantDigest:   t2Digest('a'),
		},
		{
			name: "completed without result",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				if o := h.machine.MarkCompleted(op); o.Status != string(adopt.StateCompleted) {
					t.Fatalf("complete: %q: %s", o.Status, o.Detail)
				}
			},
			wantRecorded: false,
			wantState:    adopt.StateCompleted,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newT2Host(t, t.TempDir())
			op := "op-matrix"
			tc.setup(t, h, op)
			facts := h.inspectOperation(op)
			if facts.ResultRecorded != tc.wantRecorded {
				t.Errorf("ResultRecorded = %v, want %v", facts.ResultRecorded, tc.wantRecorded)
			}
			if facts.MachineState != tc.wantState {
				t.Errorf("MachineState = %q, want %q", facts.MachineState, tc.wantState)
			}
			if facts.TornTail {
				t.Error("TornTail = true on an intact ledger")
			}
			if facts.StoredCmdDigest != tc.wantDigest {
				t.Errorf("StoredCmdDigest = %q, want %q", facts.StoredCmdDigest, tc.wantDigest)
			}
		})
	}
}

// TestT2Host_InspectHasNoSideEffects pins the matrix row-6 boundary: the
// read face does not write, schedule, or transition. Mutation anchor: any
// transition in inspectOperation (e.g. MarkUnresolved) appends an event and
// rewrites adopt.jsonl — both comparisons below turn red.
func TestT2Host_InspectHasNoSideEffects(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	op := "op-side-effect"
	adoptOnly(t, h, op)
	dispatchOnly(t, h, op)
	putResultOnly(t, h, op)

	ledgerBefore, err := os.ReadFile(filepath.Join(dir, "adopt.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	storeBefore, err := os.ReadFile(filepath.Join(dir, "t2-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	eventsBefore, err := h.machine.Events()
	if err != nil {
		t.Fatal(err)
	}

	_ = h.inspectOperation(op)
	_ = h.inspectOperation(op)

	ledgerAfter, err := os.ReadFile(filepath.Join(dir, "adopt.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	storeAfter, err := os.ReadFile(filepath.Join(dir, "t2-results.json"))
	if err != nil {
		t.Fatal(err)
	}
	eventsAfter, err := h.machine.Events()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(ledgerBefore, ledgerAfter) {
		t.Error("inspectOperation rewrote adopt.jsonl")
	}
	if !bytes.Equal(storeBefore, storeAfter) {
		t.Error("inspectOperation rewrote t2-results.json")
	}
	if len(eventsBefore) != len(eventsAfter) {
		t.Errorf("inspectOperation appended ledger events: %d -> %d", len(eventsBefore), len(eventsAfter))
	}
}

// TestT2Host_InspectReportsTornTail pins the missing-record caveat: a torn
// durable tail makes absence untrustworthy (an unrecorded dispatch may
// exist), so the facts view must carry the flag.
func TestT2Host_InspectReportsTornTail(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	op := "op-torn"
	adoptOnly(t, h, op)
	dispatchOnly(t, h, op)
	putResultOnly(t, h, op)
	if facts := h.inspectOperation(op); facts.TornTail {
		t.Fatal("TornTail = true before any corruption")
	}

	f, err := os.OpenFile(filepath.Join(dir, "adopt.jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("torn-garbage-without-newline"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	restarted := newT2Host(t, dir)
	facts := restarted.inspectOperation(op)
	if !facts.TornTail {
		t.Error("TornTail = false after an unparseable tail — absence of a record would be read as evidence of absence")
	}
	if facts.MachineState != adopt.StateDispatched {
		t.Errorf("MachineState = %q, want dispatched (the valid prefix still replays)", facts.MachineState)
	}
	if !facts.ResultRecorded {
		t.Error("ResultRecorded = false — the result store survived the torn ledger")
	}
}

// TestT2Host_RetryRefusesNonCompletedStates pins that retry never treats a
// non-completed state as re-executable. The ledger-event count is the
// discriminating assertion: deleting the guard falls through to branch C's
// Adopt, which appends a "reused"/"conflict" event even though the model is
// never called. Mutation anchor: remove the guard -> event count grows.
func TestT2Host_RetryRefusesNonCompletedStates(t *testing.T) {
	cases := []struct {
		name       string
		setup      func(t *testing.T, h *t2Host, op string)
		wantBranch string // "" = must refuse
	}{
		{
			name: "adopted",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
			},
		},
		{
			name: "dispatched without result",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
			},
		},
		{
			name: "dispatched with result (crash window)",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				putResultOnly(t, h, op)
			},
		},
		{
			name: "unresolved",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				_ = h.machine.MarkUnresolved(op)
			},
		},
		{
			name: "failed",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				_ = h.machine.MarkFailed(op)
			},
		},
		{
			name: "completed without result",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				_ = h.machine.MarkCompleted(op)
			},
		},
		{
			name: "stored result without machine record",
			setup: func(t *testing.T, h *t2Host, op string) {
				// The store has a payload but the ledger has no record (e.g.
				// adopt.jsonl was lost). Unknown state: must not re-execute.
				putResultOnly(t, h, op)
			},
		},
		{
			name: "completed with result serves",
			setup: func(t *testing.T, h *t2Host, op string) {
				if o := h.adoptAndDispatch(op, "er-1", "ev-1", t2Digest('a'), "br-"+op, "r-"+op, []string{"f"}); o.Status != "dispatched" {
					t.Fatalf("setup dispatch: %q: %s", o.Status, o.Detail)
				}
				_ = h.machine.MarkCompleted(op)
			},
			wantBranch: "A-retry",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newT2Host(t, t.TempDir())
			op := "op-retry"
			tc.setup(t, h, op)
			eventsBefore, err := h.machine.Events()
			if err != nil {
				t.Fatal(err)
			}
			cm := newCountingModel(nil)
			first := &firstExecInput{req: requestForT2(op, "er-1", "ev-1", t2Digest('a'), "br-"+op), evidence: creditEvidence{}}
			// chainGrants carries dispatch authority: without the guard the
			// entry must fall through to branch C's Adopt and append a
			// reused/conflict event (the discriminating assertion below).
			out, err := h.nodeEntryExec(op, chainGrants(), "history-verification", nil, first, Ports{Model: cm})
			if tc.wantBranch == "" {
				if err == nil {
					t.Fatalf("retry of a non-completed op returned branch %q, want a refusal", out.Branch)
				}
			} else if err != nil || out.Branch != tc.wantBranch {
				t.Fatalf("retry = (%v, %q), want (nil, %q)", err, out.Branch, tc.wantBranch)
			}
			if cm.calls != 0 {
				t.Errorf("model calls = %d, want 0", cm.calls)
			}
			eventsAfter, err := h.machine.Events()
			if err != nil {
				t.Fatal(err)
			}
			if len(eventsBefore) != len(eventsAfter) {
				t.Errorf("retry appended ledger events: %d -> %d (the guard was bypassed and Adopt ran)", len(eventsBefore), len(eventsAfter))
			}
		})
	}
}

// TestT2OpReadFactsHasNoActionVerdict is a structural pin on the read shape:
// the facts type is EXACTLY the four scalar fact fields, carries no JSON
// tags, and — because opReadFacts is an anonymous struct alias — cannot have
// methods declared on it at all. The audit's requirement is that a re-run
// permission must not be expressible here.
//
// R3 structural change: R1/R2 tried to enumerate method names and nested
// fields; both were narrower than the declaration (pointer/unexported
// receivers and verdict-named tags survived). The alias closes the method
// face at the language level; this test closes the field/tag face exactly.
func TestT2OpReadFactsHasNoActionVerdict(t *testing.T) {
	typ := reflect.TypeOf(opReadFacts{})
	want := []string{"ResultRecorded", "StoredCmdDigest", "MachineState", "TornTail"}
	if typ.NumField() != len(want) {
		t.Fatalf("opReadFacts has %d fields, want exactly %d (%v) — a new field must be argued as a FACT, not slipped in as a verdict", typ.NumField(), len(want), want)
	}
	for i, name := range want {
		f := typ.Field(i)
		if f.Name != name {
			t.Errorf("opReadFacts field %d = %q, want %q (exact facts shape)", i, f.Name, name)
		}
		switch f.Type.Kind() {
		case reflect.Bool, reflect.String:
			// scalar facts only
		default:
			t.Errorf("opReadFacts field %q has kind %s — the facts shape must stay scalar (no payload)", f.Name, f.Type.Kind())
		}
		if tag := f.Tag.Get("json"); tag != "" {
			t.Errorf("opReadFacts field %q carries json tag %q — the facts view is not a serialization shape; a verdict-named tag must not be expressible", f.Name, tag)
		}
	}
	// The anonymous alias is the primary guarantee: methods cannot be
	// declared on it. These two checks catch a future edit that turns it
	// into a defined type (Name becomes non-empty) — NumMethod alone would
	// miss an UNEXPORTED method on such a redefinition.
	if name := typ.Name(); name != "" {
		t.Errorf("opReadFacts must remain an anonymous struct alias (name=%q); a defined type can carry methods", name)
	}
	if n := typ.NumMethod(); n != 0 {
		t.Errorf("opReadFacts has %d method(s) — a method can express a verdict (e.g. SafeToRestart); the facts shape must have none", n)
	}
}

// RETIRED (R2 structural change): opReadFacts now carries no payload and no
// slice/map/pointer field, so the "facts view hands out the store's backing
// arrays" probe has no surface. The exact shape is enforced by
// TestT2OpReadFactsHasNoActionVerdict (four scalar fields, no methods); the
// payload read is nodeEntry, which is separately authorized.

// testWorld is the evidence the HOST does not own: current execution
// authority, whether the remote side accepted the call, and whether a
// business receipt exists. A real recovery layer obtains these from its own
// systems; the test double makes the composition explicit.
type testWorld struct {
	holdsAuthority  bool
	remoteAccepted  bool
	businessReceipt bool
}

// testRecoveryDecision is a TEST-SIDE consumer model. It takes the host's
// facts plus the world evidence and decides. The words serve / conflict /
// query-or-wait / start-new-attempt live HERE, not in the read face — this
// is the composition the audit requires and the old read face tried to
// pre-answer.
func testRecoveryDecision(facts opReadFacts, wantedDigest string, world testWorld) string {
	switch {
	case facts.TornTail:
		// A missing record is not evidence of absence while the durable
		// ledger is damaged.
		return "query-or-wait"
	case facts.MachineState == adopt.StateCompleted && facts.ResultRecorded:
		if facts.StoredCmdDigest == wantedDigest {
			return "serve"
		}
		return "conflict"
	case facts.MachineState == "" && !facts.ResultRecorded:
		// No local record at all: a new attempt is allowed only when the
		// caller still holds authority AND the world shows no prior effect.
		if !world.holdsAuthority || world.remoteAccepted || world.businessReceipt {
			return "query-or-wait"
		}
		return "start-new-attempt"
	default:
		// adopted / dispatched / unresolved / failed / completed-without-
		// result: local knowledge is insufficient; never restart.
		return "query-or-wait"
	}
}

// TestT2Host_FactsAreSufficientForRecoveryDecisions pins that the facts view
// is enough for a consumer to reach the correct outcome WITHOUT the read
// face returning it. The positive control (start-new-attempt) keeps the
// model from passing by answering "query-or-wait" to everything.
func TestT2Host_FactsAreSufficientForRecoveryDecisions(t *testing.T) {
	cases := []struct {
		name   string
		setup  func(t *testing.T, h *t2Host, op string)
		digest string
		world  testWorld
		want   string
	}{
		{
			name: "completed same material",
			setup: func(t *testing.T, h *t2Host, op string) {
				if o := h.adoptAndDispatch(op, "er-1", "ev-1", t2Digest('a'), "br-"+op, "r-"+op, []string{"f"}); o.Status != "dispatched" {
					t.Fatalf("setup dispatch: %q: %s", o.Status, o.Detail)
				}
				_ = h.machine.MarkCompleted(op)
			},
			digest: t2Digest('a'),
			want:   "serve",
		},
		{
			name: "completed other material",
			setup: func(t *testing.T, h *t2Host, op string) {
				if o := h.adoptAndDispatch(op, "er-1", "ev-1", t2Digest('a'), "br-"+op, "r-"+op, []string{"f"}); o.Status != "dispatched" {
					t.Fatalf("setup dispatch: %q: %s", o.Status, o.Detail)
				}
				_ = h.machine.MarkCompleted(op)
			},
			digest: t2Digest('b'),
			want:   "conflict",
		},
		{
			name:   "no record and clean world",
			setup:  func(t *testing.T, h *t2Host, op string) {},
			digest: t2Digest('a'),
			world:  testWorld{holdsAuthority: true},
			want:   "start-new-attempt",
		},
		{
			name:   "no record but no authority",
			setup:  func(t *testing.T, h *t2Host, op string) {},
			digest: t2Digest('a'),
			want:   "query-or-wait",
		},
		{
			name:   "no record but remote accepted",
			setup:  func(t *testing.T, h *t2Host, op string) {},
			digest: t2Digest('a'),
			world:  testWorld{holdsAuthority: true, remoteAccepted: true},
			want:   "query-or-wait",
		},
		{
			name: "stored result without machine record",
			setup: func(t *testing.T, h *t2Host, op string) {
				// Store/ledger disagreement: local knowledge is incomplete,
				// so even a clean world and held authority must not start a
				// new attempt (the consumer goes to query-or-wait).
				putResultDigest(t, h, op, t2Digest('b'))
			},
			digest: t2Digest('b'),
			world:  testWorld{holdsAuthority: true},
			want:   "query-or-wait",
		},
		{
			name: "dispatched without result",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
			},
			digest: t2Digest('a'),
			world:  testWorld{holdsAuthority: true},
			want:   "query-or-wait",
		},
		{
			name: "dispatched with result (crash window)",
			setup: func(t *testing.T, h *t2Host, op string) {
				adoptOnly(t, h, op)
				dispatchOnly(t, h, op)
				putResultOnly(t, h, op)
			},
			digest: t2Digest('a'),
			world:  testWorld{holdsAuthority: true},
			want:   "query-or-wait",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newT2Host(t, t.TempDir())
			op := "op-decision"
			tc.setup(t, h, op)
			if got := testRecoveryDecision(h.inspectOperation(op), tc.digest, tc.world); got != tc.want {
				t.Errorf("decision = %q, want %q", got, tc.want)
			}
		})
	}
}

// TestT2Host_TornLedgerNoRecordIsNotFresh pins the cross of the two facts:
// a torn durable ledger AND no local record. The absence of a record is not
// evidence that nothing happened (an unrecorded dispatch may exist), so the
// consumer must not start a new attempt even with a clean world and held
// authority. Mutation anchors:
//   - TornTail: TornTail && recorded        -> red (TornTail false here)
//   - TornTail: TornTail && State(op)!=""   -> red (state empty here)
//   - delete the consumer's TornTail branch -> red (decides start-new-attempt)
func TestT2Host_TornLedgerNoRecordIsNotFresh(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	// A valid event first, so adopt.jsonl exists and the garbage becomes a
	// torn tail rather than a corrupt-from-empty file.
	adoptOnly(t, h, "op-other")

	f, err := os.OpenFile(filepath.Join(dir, "adopt.jsonl"), os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString("torn-garbage-without-newline"); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}

	restarted := newT2Host(t, dir)
	facts := restarted.inspectOperation("op-unknown")
	if !facts.TornTail {
		t.Error("TornTail = false on a torn ledger with no local record")
	}
	if facts.ResultRecorded {
		t.Error("ResultRecorded = true for an op with no record")
	}
	if facts.MachineState != "" {
		t.Errorf("MachineState = %q, want empty for an op with no record", facts.MachineState)
	}
	if got := testRecoveryDecision(facts, t2Digest('a'), testWorld{holdsAuthority: true}); got != "query-or-wait" {
		t.Errorf("decision = %q, want query-or-wait — a torn ledger cannot license a new attempt", got)
	}
}

// TestT2Host_LedgerLossDoesNotRewriteStoredProvenance pins the store-only
// guard against the overwrite path: a stored record carries a knowledge
// pointer (KnowledgeUsed exists on the base tree), the ledger has no record,
// and a fresh registry serves the SAME method ID. Without the guard the
// entry re-executes the method and overwrites the stored pointer through the
// same receipt; with the guard it refuses and the pointer is untouched.
// Mutation anchor: delete the store-only guard -> red on both assertions.
func TestT2Host_LedgerLossDoesNotRewriteStoredProvenance(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	if err := h.store.put("op-loss", storedResult{
		Receipt:       "method:history-check:result",
		Facts:         []string{"fact-loss"},
		BusinessReqID: "br-loss",
		EvaluationRev: "er-1",
		KnowledgeUsed: "declared-v1",
		CmdDigest:     t2Digest('f'),
	}); err != nil {
		t.Fatal(err)
	}

	restarted := newT2Host(t, dir) // ledger empty, store survives
	restarted.registry.mustPublish(t, knowledgeMethod{
		ID: "history-check", Version: "2.0.0",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true,
	})
	first := &firstExecInput{req: requestForT2("op-loss", "er-1", "ev-1", t2Digest('f'), "br-loss"), evidence: creditEvidence{}}
	if _, err := restarted.nodeEntryExec("op-loss", chainGrants(), "history-verification",
		map[string]string{"overdue_months": "0", "business_request_id": "br-loss"}, first, Ports{Model: newCountingModel(nil)}); err == nil {
		t.Fatal("retry re-executed with a store-only record")
	}
	r, ok := restarted.store.get("op-loss")
	if !ok || r.KnowledgeUsed != "declared-v1" || r.CmdDigest != t2Digest('f') {
		t.Fatalf("ledger loss rewrote the stored provenance: %+v", r)
	}
}
