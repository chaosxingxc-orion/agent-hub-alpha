package e2e

// W8 (GC#24 / GH#55): a completed knowledge-method result must carry the
// method identity and declared version durably — a fresh host (empty
// in-memory registry) must be able to read them from the result file.
//
// RED-BEFORE on main: KnowledgeUsed is declared but never written, and the
// version is absent from both durable artifacts.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
)

// readPersistedResults decodes the durable result file. It is the
// discriminating leg: the audit's bytes.Contains check proves the version
// string exists somewhere in the artifacts, while this decode proves it is
// attached to the operation's result record.
func readPersistedResults(t *testing.T, dir string) map[string]storedResult {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "t2-results.json"))
	if err != nil {
		t.Fatalf("read t2-results.json: %v", err)
	}
	var out map[string]storedResult
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decode t2-results.json: %v", err)
	}
	return out
}

// TestT2KnowledgeProvenance_AuditRepro is the audit's reproduction: publish
// history-check@1.7.3, run the knowledge fast path (zero model calls), then
// reconstruct the host from the durable files only and retry.
func TestT2KnowledgeProvenance_AuditRepro(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	method := knowledgeMethod{
		ID:          "history-check",
		Version:     "1.7.3",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true,
	}
	h.registry.mustPublish(t, method)

	cm := newCountingModel(nil)
	first := &firstExecInput{req: requestForT2("op-history", "er-1", "ev-1", t2Digest('c'), "br-history"), evidence: creditEvidence{}}
	out, err := h.nodeEntryExec("op-history", chainGrants(), "history-verification",
		map[string]string{"overdue_months": "0", "business_request_id": "br-history"}, first, Ports{Model: cm})
	if err != nil {
		t.Fatalf("first execution: %v", err)
	}
	if out.Branch != "B-method" {
		t.Fatalf("branch = %q, want B-method", out.Branch)
	}
	if cm.calls != 0 {
		t.Fatalf("model calls = %d, want 0", cm.calls)
	}
	if s := h.machine.State("op-history"); s != adopt.StateCompleted {
		t.Fatalf("state = %q, want completed", s)
	}

	// A fresh host reads only the durable files; the registry is empty.
	restarted := newT2Host(t, dir)
	if len(restarted.registry.methods) != 0 {
		t.Fatal("fresh host unexpectedly has a method registry")
	}
	retry := &firstExecInput{req: requestForT2("op-history", "er-1", "ev-1", t2Digest('c'), "br-history"), evidence: creditEvidence{}}
	got, err := restarted.nodeEntryExec("op-history", fullGrantsN3(), "history-verification", nil, retry, Ports{})
	if err != nil {
		t.Fatalf("retry after fresh construction: %v", err)
	}
	if got.Branch != "A-retry" {
		t.Fatalf("retry branch = %q, want A-retry", got.Branch)
	}
	if got.Result.Receipt != out.Result.Receipt {
		t.Fatalf("retry receipt = %q, want %q", got.Result.Receipt, out.Result.Receipt)
	}
	if got.Result.BusinessReqID != "br-history" || got.Result.EvaluationRev != "er-1" || got.Result.CmdDigest != t2Digest('c') {
		t.Fatalf("retry lost business identity/material: %+v", got.Result)
	}
	if got.Result.KnowledgeUsed != method.ID {
		t.Errorf("persisted result lost the actual method identity: want %q got %q", method.ID, got.Result.KnowledgeUsed)
	}
	if got.Result.KnowledgeVersion != method.Version {
		t.Errorf("persisted result lost the declared method version: want %q got %q", method.Version, got.Result.KnowledgeVersion)
	}

	// The version must be attached to THIS operation's record, not merely
	// present somewhere in the artifacts (the audit's Contains leg alone
	// would go green on an empty or unrelated needle).
	persistedRows := readPersistedResults(t, dir)
	row, ok := persistedRows["op-history"]
	if !ok {
		t.Fatalf("op-history has no persisted result row")
	}
	if row.KnowledgeUsed != method.ID || row.KnowledgeVersion != method.Version {
		t.Fatalf("persisted row provenance = (%q,%q), want (%q,%q)", row.KnowledgeUsed, row.KnowledgeVersion, method.ID, method.Version)
	}

	persisted := []byte{}
	for _, name := range []string{"adopt.jsonl", "t2-results.json"} {
		raw, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			t.Fatalf("read %s: %v", name, err)
		}
		persisted = append(persisted, raw...)
	}
	if !bytes.Contains(persisted, []byte(method.Version)) {
		t.Errorf("method version %q is absent from both durable artifacts; only the in-memory registry knew it", method.Version)
	}
}

// TestT2KnowledgeProvenance_HistoryPinnedToServedVersion pins the other
// direction: publishing a newer method (and invalidating the older one)
// must not rewrite what an already-completed result says. The retry path
// must serve the stored provenance, not re-look-up the live registry.
//
// Mutation anchors: (a) stamping in branch A from findForNode turns op1
// into v2 and reddens the history leg; (b) hard-coding a version turns the
// new-task leg red (both legs are needed — either one alone is satisfiable
// by a constant).
func TestT2KnowledgeProvenance_HistoryPinnedToServedVersion(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	h.registry.mustPublish(t, knowledgeMethod{
		ID: "method-overdue-v1", Version: "1.0.0",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true,
	})
	first := &firstExecInput{req: requestForT2("op-hist-v1", "er-1", "ev-1", t2Digest('d'), "br-hist-v1"), evidence: creditEvidence{}}
	out, err := h.nodeEntryExec("op-hist-v1", chainGrants(), "history-verification",
		map[string]string{"overdue_months": "0", "business_request_id": "br-hist-v1"}, first, Ports{Model: newCountingModel(nil)})
	if err != nil || out.Branch != "B-method" {
		t.Fatalf("v1 execution = (%v, %q), want B-method", err, out.Branch)
	}
	if out.Result.KnowledgeVersion != "1.0.0" {
		t.Fatalf("v1 result version = %q, want 1.0.0", out.Result.KnowledgeVersion)
	}

	// A correction supersedes v1 and v1 is invalidated (the single flip path).
	h.registry.mustPublish(t, knowledgeMethod{
		ID: "method-overdue-v2", Version: "2.0.0",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true, Replaces: "method-overdue-v1",
	})
	if err := h.registry.invalidate("method-overdue-v1"); err != nil {
		t.Fatal(err)
	}

	// Fresh host, registry re-published to the CURRENT state (v1 historical,
	// v2 live and superseding). The retry must still read v1 from the
	// durable result: a branch-A re-lookup of findForNode would return v2
	// here and turn this leg red (the empty-registry audit repro cannot
	// distinguish "stored" from "looked up and found nothing").
	restarted := newT2Host(t, dir)
	restarted.registry.mustPublish(t, knowledgeMethod{
		ID: "method-overdue-v1", Version: "1.0.0",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true,
	})
	restarted.registry.mustPublish(t, knowledgeMethod{
		ID: "method-overdue-v2", Version: "2.0.0",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true, Replaces: "method-overdue-v1",
	})
	retry := &firstExecInput{req: requestForT2("op-hist-v1", "er-1", "ev-1", t2Digest('d'), "br-hist-v1"), evidence: creditEvidence{}}
	got, err := restarted.nodeEntryExec("op-hist-v1", fullGrantsN3(), "history-verification", nil, retry, Ports{})
	if err != nil || got.Branch != "A-retry" {
		t.Fatalf("retry = (%v, %q), want A-retry", err, got.Branch)
	}
	if got.Result.KnowledgeVersion != "1.0.0" || got.Result.KnowledgeUsed != "method-overdue-v1" {
		t.Fatalf("history result provenance = (%q,%q), want (method-overdue-v1,1.0.0)", got.Result.KnowledgeUsed, got.Result.KnowledgeVersion)
	}

	// A NEW task in the same host consumes the correction, v2 — a frozen
	// v1 stamp cannot pass both legs.
	newTask := &firstExecInput{req: requestForT2("op-hist-v2", "er-1", "ev-1", t2Digest('e'), "br-hist-v2"), evidence: creditEvidence{}}
	out2, err := restarted.nodeEntryExec("op-hist-v2", chainGrants(), "history-verification",
		map[string]string{"overdue_months": "1", "business_request_id": "br-hist-v2"}, newTask, Ports{Model: newCountingModel(nil)})
	if err != nil || out2.Branch != "B-method" {
		t.Fatalf("new task = (%v, %q), want B-method", err, out2.Branch)
	}
	if out2.Result.KnowledgeVersion != "2.0.0" || out2.Result.KnowledgeUsed != "method-overdue-v2" {
		t.Fatalf("new task provenance = (%q,%q), want (method-overdue-v2,2.0.0)", out2.Result.KnowledgeUsed, out2.Result.KnowledgeVersion)
	}

	rows := readPersistedResults(t, dir)
	if rows["op-hist-v1"].KnowledgeVersion != "1.0.0" || rows["op-hist-v2"].KnowledgeVersion != "2.0.0" {
		t.Fatalf("persisted versions = (v1:%q, v2:%q), want (1.0.0, 2.0.0)", rows["op-hist-v1"].KnowledgeVersion, rows["op-hist-v2"].KnowledgeVersion)
	}
}
func TestT2KnowledgeProvenance_VetoResultCarriesProvenance(t *testing.T) {
	dir := t.TempDir()
	h := newT2Host(t, dir)
	method := knowledgeMethod{
		ID: "veto-check", Version: "9.9.9",
		ServesNode:  "history-verification",
		InputFields: []string{"overdue_months", "business_request_id"},
		Valid:       true,
	}
	h.registry.mustPublish(t, method)

	cm := newCountingModel(nil)
	first := &firstExecInput{req: requestForT2("op-veto", "er-1", "ev-1", t2Digest('9'), "br-veto"), evidence: creditEvidence{}}
	out, err := h.nodeEntryExec("op-veto", chainGrants(), "history-verification",
		map[string]string{"overdue_months": "4", "business_request_id": "br-veto"}, first, Ports{Model: cm})
	if err != nil || out.Branch != "B-method" {
		t.Fatalf("veto execution = (%v, %q), want B-method", err, out.Branch)
	}
	if cm.calls != 0 {
		t.Fatalf("model calls = %d, want 0", cm.calls)
	}
	if !strings.Contains(out.Result.Receipt, "veto-fired") {
		t.Fatalf("receipt = %q, want the veto shape", out.Result.Receipt)
	}
	if out.Result.KnowledgeUsed != method.ID || out.Result.KnowledgeVersion != method.Version {
		t.Fatalf("veto provenance = (%q,%q), want (%q,%q)", out.Result.KnowledgeUsed, out.Result.KnowledgeVersion, method.ID, method.Version)
	}

	restarted := newT2Host(t, dir) // empty registry: read the durable row only
	retry := &firstExecInput{req: requestForT2("op-veto", "er-1", "ev-1", t2Digest('9'), "br-veto"), evidence: creditEvidence{}}
	got, err := restarted.nodeEntryExec("op-veto", fullGrantsN3(), "history-verification", nil, retry, Ports{})
	if err != nil || got.Branch != "A-retry" {
		t.Fatalf("veto retry = (%v, %q), want A-retry", err, got.Branch)
	}
	if got.Result.KnowledgeUsed != method.ID || got.Result.KnowledgeVersion != method.Version {
		t.Fatalf("retried veto provenance = (%q,%q), want (%q,%q)", got.Result.KnowledgeUsed, got.Result.KnowledgeVersion, method.ID, method.Version)
	}
	rows := readPersistedResults(t, dir)
	if rows["op-veto"].KnowledgeUsed != method.ID || rows["op-veto"].KnowledgeVersion != method.Version {
		t.Fatalf("persisted veto provenance = (%q,%q), want (%q,%q)", rows["op-veto"].KnowledgeUsed, rows["op-veto"].KnowledgeVersion, method.ID, method.Version)
	}
}
