package e2e

// Machine-level pin for equivalent grant regroupings (external audit GC#23 /
// GH#56): the real Dispatch gate — not just the fixture predicate — must accept
// the SAME authority written as two grants when it accepts it as one.
//
// The scenario is the audit's: a provider splits "{modify, execute}" into
// "{modify}" + "{execute}" without changing the resource, the actions or the
// effect ceiling. Admission accepts both forms; dispatch must agree.
//
// Mutation anchors: reverting adopt.GrantsCover to the single-grant rule makes
// the split case rejected; dropping the per-action effect check lets the
// mixed-effect case through (see internal/adopt/grants_split_test.go).

import (
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/ledger"
)

func twoActionToolGrant(effect string) ledger.Grant {
	return ledger.Grant{
		Kind:     "tool",
		Selector: "tool://credit-transfer",
		Actions:  []string{"execute", "write"},
		Effect:   effect,
	}
}

// adoptWithGrants adopts op with the given grants pinned at decision time.
func adoptWithGrants(t *testing.T, h *t2Host, op string, grants []ledger.Grant) {
	t.Helper()
	req := requestForT2(op, "er-split", "ev-split", t2Digest('a'), "br-split")
	req.GrantsAtDecision = grants
	if out := h.machine.Adopt(req); out.Status != "adopted" {
		t.Fatalf("adopt with pinned grants: %+v", out)
	}
}

// TestDispatchAcceptsEquivalentGrantRegrouping pins the audit's exact shape at
// the real gate: merged and split representations of one authority both
// dispatch.
func TestDispatchAcceptsEquivalentGrantRegrouping(t *testing.T) {
	pinned := []ledger.Grant{twoActionToolGrant("irreversible")}

	h := newT2Host(t, t.TempDir())
	adoptWithGrants(t, h, "op-merged", pinned)
	if out := h.machine.Dispatch("op-merged", pinned); out.Status != "dispatched" {
		t.Fatalf("merged representation rejected: %+v", out)
	}

	split := []ledger.Grant{
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"execute"}, Effect: "irreversible"},
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"write"}, Effect: "irreversible"},
	}
	h2 := newT2Host(t, t.TempDir())
	adoptWithGrants(t, h2, "op-split", pinned)
	if out := h2.machine.Dispatch("op-split", split); out.Status != "dispatched" {
		t.Fatalf("equivalent split rejected at dispatch while admission accepted it: %+v", out)
	}
}

// TestDispatchRefusesRegroupingThatDropsTheCeiling pins the other side: the
// same split shape is refused when the action that needs the irreversible
// ceiling is carried by a weaker grant. A naive union over the live grants
// would pass this, which is why the coverage check is per action.
func TestDispatchRefusesRegroupingThatDropsTheCeiling(t *testing.T) {
	pinned := []ledger.Grant{twoActionToolGrant("irreversible")}
	h := newT2Host(t, t.TempDir())
	adoptWithGrants(t, h, "op-mixed", pinned)

	mixed := []ledger.Grant{
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"execute"}, Effect: "none"},
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"write"}, Effect: "irreversible"},
	}
	if out := h.machine.Dispatch("op-mixed", mixed); out.Status != "rejected" {
		t.Fatalf("split that lowers the effect ceiling must be refused, got %+v", out)
	}
}

// TestDispatchRefusesDroppedAction pins the revocation behaviour through the
// regrouped path: removing one of the pinned actions still fails.
func TestDispatchRefusesDroppedAction(t *testing.T) {
	pinned := []ledger.Grant{twoActionToolGrant("irreversible")}
	h := newT2Host(t, t.TempDir())
	adoptWithGrants(t, h, "op-drop", pinned)

	live := []ledger.Grant{
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"execute"}, Effect: "irreversible"},
	}
	if out := h.machine.Dispatch("op-drop", live); out.Status != "rejected" {
		t.Fatalf("a dropped action must keep the op refused, got %+v", out)
	}
}

// TestFixtureJudgeDelegatesToMachineRule pins that the two fixture predicates
// ARE the machine rule: the split regrouping passes through them, and the
// effect comparison is rank-based (a stronger live grant covers a weaker
// requirement). Mutation anchor: restoring the old fixture-local loops
// (single-grant rule / strict equality on effect) makes this red while the
// machine-level tests above stay green.
func TestFixtureJudgeDelegatesToMachineRule(t *testing.T) {
	split := []ledger.Grant{
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"execute"}, Effect: "irreversible"},
		{Kind: "tool", Selector: "tool://credit-transfer", Actions: []string{"write"}, Effect: "irreversible"},
	}
	if !grantsCoverEffect(split, []ledger.Grant{twoActionToolGrant("irreversible")}) {
		t.Fatal("dispatch fixture must accept the equivalent split")
	}
	if !grantsCoverRead(split, []ledger.Grant{twoActionToolGrant("irreversible")}) {
		t.Fatal("read fixture must accept the equivalent split")
	}
	// rank semantics reached through the fixture: irreversible covers reversible
	if !grantsCoverEffect([]ledger.Grant{twoActionToolGrant("irreversible")}, []ledger.Grant{twoActionToolGrant("reversible")}) {
		t.Fatal("dispatch fixture must compare effect by rank, not equality")
	}
}
