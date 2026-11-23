package ledger

// One-JSON-value-per-line pins (external audit GC#25 / GH#54).
//
// The loader's two decoders accept a first value and ignore whatever follows,
// so a durable-form line that once was TWO records squeezed together reads as
// the first one only: the second record vanishes, and a restart can re-issue a
// dispatch permit that the missing record was meant to block. The gate is the
// stdlib whole-text judgment (json.Valid) placed before both decoders.
//
// Position semantics follow the existing torn-tail doctrine: a bad segment at
// EOF *without* a terminating newline is a torn tail (the machine opens
// read-only); anything newline-terminated, or bad anywhere before the end, is
// a hard error.
//
// Mutation anchor: deleting the json.Valid gate makes the squeezed-line arms
// green again (they parse as the first record).

import (
	"strconv"
	"strings"
	"testing"
)

// eventLine returns a minimal line that the typed decode accepts (the shape
// judgment only needs event_type / arbitration_seq / identity.operation_id).
func eventLine(seq int, eventType, op string) string {
	return `{"event_type":"` + eventType + `","arbitration_seq":` + strconv.Itoa(seq) + `,"identity":{"operation_id":"` + op + `"}}`
}

// TestParseLines_RejectsSqueezedRecords pins the audit's exact shape: two event
// objects on one line must not be read as one.
func TestParseLines_RejectsSqueezedRecords(t *testing.T) {
	first := eventLine(0, "adopted", "op-1")
	second := eventLine(1, "dispatched", "op-1")
	raw := []byte(first + second + "\n")

	lines, torn, err := ParseLines(raw)
	if err == nil {
		t.Fatalf("a line holding two records must be refused, got %d parsed lines (torn=%v)", len(lines), torn)
	}
	if !strings.Contains(err.Error(), "not exactly one JSON value") {
		t.Errorf("error should name the actual problem, got: %v", err)
	}
	if !strings.Contains(err.Error(), "re-Open") {
		t.Errorf("a refusal on a repairable file must name the next step, got: %v", err)
	}
	if torn {
		t.Error("a newline-terminated bad line is corruption, not a torn tail")
	}
}

// TestParseLines_TrailingGarbageFamily pins the wider family the same gate
// closes: any trailing data after the first value (not just a second object).
func TestParseLines_TrailingGarbageFamily(t *testing.T) {
	base := eventLine(0, "adopted", "op-1")
	// "}]" and "}}" are the two forms where a decoder-scoped check (More())
	// reports no more elements while the text still carries trailing data —
	// they pin the choice of a whole-text judgment over a stream-scoped one.
	for _, tail := range []string{"x", "[]", "{} {}", "null", `{"b":1}`, "}]", "}}"} {
		if _, torn, err := ParseLines([]byte(base + tail + "\n")); err == nil {
			t.Errorf("trailing %q must be refused, got err=nil (torn=%v)", tail, torn)
		}
	}
}

// TestParseLines_SqueezedLastLineWithoutNewlineIsTorn pins the position
// semantics: the same damage at EOF without a terminating newline is the torn
// tail the doctrine already describes (open read-only), not a hard error.
func TestParseLines_SqueezedLastLineWithoutNewlineIsTorn(t *testing.T) {
	raw := []byte(eventLine(0, "adopted", "op-1") + eventLine(1, "dispatched", "op-1"))
	lines, torn, err := ParseLines(raw)
	if err != nil {
		t.Fatalf("a torn final segment must be reported via the flag, not an error: %v", err)
	}
	if !torn {
		t.Error("the squeezed final segment must set torn")
	}
	if len(lines) != 0 {
		t.Errorf("the damaged segment must not contribute a line, got %d", len(lines))
	}
}

// TestParseLines_NormalFormsStillPass pins the must-pass group: legitimate
// durable forms (including whitespace and CRLF) are untouched by the gate.
func TestParseLines_NormalFormsStillPass(t *testing.T) {
	l0 := eventLine(0, "adopted", "op-1")
	l1 := eventLine(1, "dispatched", "op-1")
	cases := map[string]struct {
		raw  string
		want int
	}{
		"plain":            {l0 + "\n" + l1 + "\n", 2},
		"no final newline": {l0 + "\n" + l1, 2},
		"trailing spaces":  {l0 + " \n" + l1 + "  \n", 2},
		"CRLF":             {l0 + "\r\n" + l1 + "\r\n", 2},
		"blank lines":      {l0 + "\n\n" + l1 + "\n", 2},
		"single line":      {l0 + "\n", 1},
	}
	for name, c := range cases {
		lines, torn, err := ParseLines([]byte(c.raw))
		if err != nil || torn {
			t.Errorf("%s: a legitimate durable form must pass (err=%v torn=%v)", name, err, torn)
			continue
		}
		// Exact count: a regression that drops a legitimate line would
		// otherwise stay green behind a len>0 assertion.
		if len(lines) != c.want {
			t.Errorf("%s: parsed %d lines, want %d", name, len(lines), c.want)
		}
	}
}
