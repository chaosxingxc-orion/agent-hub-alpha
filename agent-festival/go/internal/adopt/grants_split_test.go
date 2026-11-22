package adopt

// Grant-coverage atoms (external audit GC#23 / GH#56).
//
// The authority expressed by a grant is the set of (kind, selector, action)
// atoms it carries. A pinned authority may legitimately be written as one
// grant or split across equivalent grants — the provider regrouped its
// records without changing what is allowed — and both forms must be judged
// the same. What must NOT become transparent is a regroup that changes the
// atoms: dropping an action, lowering the effect level of the grant that
// carries it, or moving it to another kind/selector.
//
// Mutation anchors: reverting GrantsCover to "one live grant must cover all
// pinned actions" turns the split cases red; aggregating actions across
// grants WITHOUT the per-action effect check turns split-mixed-effect red
// (a naive union would let an irreversible action ride on a none-level grant).

import (
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/ledger"
)

const (
	gcKind   = "tool"
	gcSel    = "tool://credit-transfer"
	gcEffect = "irreversible"
)

func gcGrant(effect string, actions ...string) ledger.Grant {
	return ledger.Grant{Kind: gcKind, Selector: gcSel, Effect: effect, Actions: actions}
}

func TestGrantsCoverSplitEquivalence(t *testing.T) {
	pinned := []ledger.Grant{gcGrant(gcEffect, "execute", "write")}

	cases := []struct {
		name string
		live []ledger.Grant
		want bool
	}{
		{
			name: "merged (one grant carries both actions)",
			live: []ledger.Grant{gcGrant(gcEffect, "execute", "write")},
			want: true,
		},
		{
			name: "split-equivalent (same atoms across two grants)",
			live: []ledger.Grant{gcGrant(gcEffect, "execute"), gcGrant(gcEffect, "write")},
			want: true,
		},
		{
			name: "split-mixed-effect (action rides a weaker grant) must be refused",
			live: []ledger.Grant{gcGrant("none", "execute"), gcGrant(gcEffect, "write")},
			want: false,
		},
		{
			name: "dropped action",
			live: []ledger.Grant{gcGrant(gcEffect, "execute")},
			want: false,
		},
		{
			name: "lowered effect on the grant carrying all actions",
			live: []ledger.Grant{gcGrant("reversible", "execute", "write")},
			want: false,
		},
		{
			name: "different selector",
			live: []ledger.Grant{{Kind: gcKind, Selector: "tool://other", Effect: gcEffect, Actions: []string{"execute", "write"}}},
			want: false,
		},
		{
			name: "different kind",
			live: []ledger.Grant{{Kind: "data", Selector: gcSel, Effect: gcEffect, Actions: []string{"execute", "write"}}},
			want: false,
		},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := GrantsCover(c.live, pinned); got != c.want {
				t.Fatalf("GrantsCover = %v, want %v", got, c.want)
			}
		})
	}
}

// TestGrantsCoverHigherEffectStillCovers pins the other direction of the
// effect comparison: a stronger live grant covers a weaker requirement.
func TestGrantsCoverHigherEffectStillCovers(t *testing.T) {
	pinned := []ledger.Grant{gcGrant("reversible", "execute")}
	live := []ledger.Grant{gcGrant("irreversible", "execute")}
	if !GrantsCover(live, pinned) {
		t.Fatal("an irreversible live grant must cover a reversible requirement")
	}
}

// TestGrantsCoverEmptyPinnedActionsRefused pins that a pinned grant with no
// actions is a fixture bug, not a universal permit (the old loop returned true
// vacuously for it).
func TestGrantsCoverEmptyPinnedActionsRefused(t *testing.T) {
	pinned := []ledger.Grant{{Kind: gcKind, Selector: gcSel, Effect: gcEffect}}
	if GrantsCover([]ledger.Grant{gcGrant(gcEffect, "execute")}, pinned) {
		t.Fatal("a pinned grant with no actions must not be treated as covered")
	}
}

// TestGrantsCoverUnknownPinnedEffectFailsClosed pins the one-line fail-closed
// guard for an effect level the judge cannot rank: rank(-1) compares below
// nothing, so without the guard any live grant would "cover" it.
func TestGrantsCoverUnknownPinnedEffectFailsClosed(t *testing.T) {
	live := []ledger.Grant{gcGrant("none", "execute")}
	for _, bogus := range []string{"", "Irreversible", "super-effect"} {
		pinned := []ledger.Grant{{Kind: gcKind, Selector: gcSel, Effect: bogus, Actions: []string{"execute"}}}
		if GrantsCover(live, pinned) {
			t.Errorf("pinned effect %q is unrankable and must be refused", bogus)
		}
	}
}

// TestGrantsCoverEmptyPinnedIsVacuouslyCovered keeps the machine's existing
// semantics for an empty REQUIREMENT LIST (nothing to require): the schema
// refuses an empty grants array at the ledger boundary, so this is a defensive
// shape rather than a permit path.
func TestGrantsCoverEmptyPinnedIsVacuouslyCovered(t *testing.T) {
	if !GrantsCover([]ledger.Grant{gcGrant(gcEffect, "execute")}, nil) {
		t.Fatal("an empty pinned list has nothing to require")
	}
}
