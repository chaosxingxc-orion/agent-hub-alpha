package adopt

// Consumer-level pin for the one-value-per-line gate (external audit GC#25 /
// GH#54). The loader-level arms live in internal/ledger; this file pins what
// the MACHINE does with such a file:
//   - newline-terminated trailing data → Open fails (corruption, not a torn
//     tail) — the machine must not resume from a file whose second record on a
//     line was silently dropped;
//   - the same damage at EOF without a newline → Open succeeds read-only and
//     the dispatch permit is NOT re-issued (the existing torn-tail doctrine,
//     already pinned by TestCrashRecoveryTornTail).
//
// Mutation anchor: removing the json.Valid gate makes the first arm open
// successfully with one event parsed (the trailing object is ignored).

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func squeezedFile(t *testing.T, newlineTerminated bool) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "adopt.jsonl")
	m := mustOpen(t, path)
	if o := m.Adopt(testRequest("op-sq", digestA, pinGrants())); o.Status != "adopted" {
		t.Fatalf("adopt = %q, want adopted", o.Status)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	// squeeze a second object onto the last record's line
	squeezed := append(bytes.TrimRight(raw, "\n"), []byte(`{"event_type":"dispatched","arbitration_seq":99}`)...)
	if newlineTerminated {
		squeezed = append(squeezed, '\n')
	}
	if err := os.WriteFile(path, squeezed, 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestOpenRefusesTrailingDataOnALine(t *testing.T) {
	path := squeezedFile(t, true)
	if _, err := Open(path); err == nil {
		t.Fatal("Open must refuse a newline-terminated line carrying trailing data — resuming would drop the second record")
	}
}

func TestOpenTornWhenTrailingDataIsTheLastUnterminatedSegment(t *testing.T) {
	path := squeezedFile(t, false)
	m, err := Open(path)
	if err != nil {
		t.Fatalf("a torn final segment must open read-only, got: %v", err)
	}
	if !m.TornTail() {
		t.Fatal("TornTail() = false, want true")
	}
	// The rejection must be the torn-tail gate's, not an incidental
	// "unknown operation" (the squeezed segment makes the projection start
	// empty, so a bare status check would pass for the wrong reason).
	o := m.Dispatch("op-sq", pinGrants())
	if o.Status != "rejected" {
		t.Fatalf("dispatch on a damaged form = %q, want rejected — a dropped record must not re-arm the permit", o.Status)
	}
	if !strings.Contains(o.Detail, "torn tail") {
		t.Fatalf("rejection detail = %q, want the torn-tail gate — otherwise this arm would pass on an unknown-op rejection", o.Detail)
	}
}
