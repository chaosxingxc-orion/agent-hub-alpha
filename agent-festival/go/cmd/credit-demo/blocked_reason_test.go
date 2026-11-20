package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Blocked-reason projection (W02 PR #59 request) and reconciliation of runs
// that have nothing left to execute (W02 R1-R13 after the e1961fe8 upgrade).

func graphRunOf(t *testing.T, s *Service, r *Run) *GraphRun {
	t.Helper()
	_, p, ok := s.RunByID(r.ID)
	if !ok || p == nil {
		t.Fatalf("run %s has no plan", r.ID)
	}
	g, err := s.graph(p, false) // takes s.mu itself
	if err != nil || g.Run == nil {
		t.Fatalf("graph: %v", err)
	}
	return g.Run
}

func summaryOf(t *testing.T, s *Service, r *Run) PlanSummary {
	t.Helper()
	for _, row := range s.RecentPlans(100) {
		if row.Run == r.ID {
			return row
		}
	}
	t.Fatalf("run %s not listed", r.ID)
	return PlanSummary{}
}

func expectBlocked(t *testing.T, s *Service, r *Run, code, cause string) {
	t.Helper()
	g, row := graphRunOf(t, s, r), summaryOf(t, s, r)
	for where, b := range map[string]*BlockReason{"graph": g.Blocked, "plan list": row.Blocked} {
		if b == nil || b.Code != code || b.Cause != cause || b.Note == "" || b.At == "" {
			t.Fatalf("%s: blocked = %+v, want code %s cause %q with a note", where, b, code, cause)
		}
	}
}

func TestBlockedReasonSpecUnresolvable(t *testing.T) {
	e := newAR04(t, false)
	s := e.open()
	_, r := e.proposeAndApprove(s)
	e.spec = ar04Spec(t, t.TempDir(), "2.0.0", "RULE-V2")
	if err := os.RemoveAll(filepath.Join(e.dir, "specs")); err != nil {
		t.Fatal(err)
	}
	s = e.open()
	if _, err := s.Lease("worker-after-restart"); err != nil {
		t.Fatal(err)
	}
	expectBlocked(t, s, s.st.Runs[r.ID], "SPEC_UNRESOLVABLE", "")
	if note := graphRunOf(t, s, s.st.Runs[r.ID]).Blocked.Note; strings.Contains(note, e.dir) || strings.Contains(note, "no such file") {
		t.Fatalf("note leaks the raw error: %q", note)
	}
}

func TestBlockedReasonModelBindingMismatch(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _, c := m1RemoteCall(t, s, e)
	if _, err := s.AgentContext(agentCmd(r, c)); err != nil {
		t.Fatal(err)
	}
	cmd := agentCmd(r, c)
	cmd.EventID, cmd.Model = "ev-1", "model-b"
	if err := s.ModelCall(cmd); err == nil {
		t.Fatal("mismatch accepted")
	}
	expectBlocked(t, s, s.st.Runs[r.ID], "MODEL_BINDING_MISMATCH", "")
	if note := graphRunOf(t, s, s.st.Runs[r.ID]).Blocked.Note; !strings.Contains(note, "model-a") || !strings.Contains(note, "model-b") {
		t.Fatalf("note does not name both models: %q", note)
	}
}

// adoptedRun drives a run through an adopted judgment; finish also reports
// it done with a final report.
func adoptedRun(t *testing.T, e *ar04Env, s *Service, finish bool) (*Run, *Job) {
	t.Helper()
	_, r := e.proposeAndApprove(s)
	job, err := s.LeaseAs("worker", "model-a")
	if err != nil || job == nil {
		t.Fatalf("lease: %v", err)
	}
	if err := s.StepPut(r.ID, job.Token, "related", "verified evidence"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Judgment(r.ID, job.Token, "judge", `{"level":"grant-50000","evidence_steps":["related"],"rationale":"original"}`); err != nil {
		t.Fatal(err)
	}
	if finish {
		if err := s.Done(r.ID, job.Token, nil, "final credit report", ""); err != nil {
			t.Fatal(err)
		}
	}
	return r, job
}

func makeLegacy(t *testing.T, s *Service, r *Run) {
	t.Helper()
	s.st.Runs[r.ID].ReplaySteps[0]["request"].(map[string]any)["adoption_version"] = "0.1.0"
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
}

// A running run on the old adoption protocol is blocked, and the reason says
// it is the protocol, not a sequence mismatch.
func TestReconciliationNamesLegacyProtocol(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _ := adoptedRun(t, e, s, false)
	makeLegacy(t, s, r)
	s = e.open()
	got := s.st.Runs[r.ID]
	if got.Status != "blocked" || !strings.Contains(got.Final, "RECONCILIATION_REQUIRED") || !strings.Contains(got.Final, "旧协议 0.1.0") {
		t.Fatalf("running legacy run: %s %q", got.Status, got.Final)
	}
	expectBlocked(t, s, got, "RECONCILIATION_REQUIRED", "LEGACY_ADOPTION_PROTOCOL")
}

// A finished run whose records fail the startup check stays done with its
// final report; the failed check is recorded beside it and projected.
// Repeated restarts do not rewrite it.
func TestReconciliationKeepsFinishedRunVisible(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _ := adoptedRun(t, e, s, true)
	makeLegacy(t, s, r)
	s = e.open()
	got := s.st.Runs[r.ID]
	if got.Status != "done" || got.Final != "final credit report" || got.Blocked != nil {
		t.Fatalf("finished run changed: status=%s final=%q blocked=%+v", got.Status, got.Final, got.Blocked)
	}
	rc := got.Reconciliation
	if rc == nil || rc.Code != "RECONCILIATION_REQUIRED" || rc.Cause != "LEGACY_ADOPTION_PROTOCOL" || !strings.Contains(rc.Note, "旧协议 0.1.0") {
		t.Fatalf("reconciliation flag: %+v", rc)
	}
	if g := graphRunOf(t, s, got); g.Blocked != nil || g.Reconciliation == nil || g.Reconciliation.Cause != "LEGACY_ADOPTION_PROTOCOL" {
		t.Fatalf("graph: blocked=%+v reconciliation=%+v", g.Blocked, g.Reconciliation)
	}
	if row := summaryOf(t, s, got); row.RunStatus != "done" || row.Reconciliation == nil {
		t.Fatalf("plan list: %+v", row)
	}
	rev, at := s.st.Rev, rc.At
	s = e.open()
	if s.st.Rev != rev || s.st.Runs[r.ID].Reconciliation.At != at {
		t.Fatalf("second restart rewrote the state (rev %d -> %d)", rev, s.st.Rev)
	}
}

// A consistent finished run gets no flag.
func TestReconciliationLeavesConsistentFinishedRunAlone(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _ := adoptedRun(t, e, s, true)
	s = e.open()
	got := s.st.Runs[r.ID]
	if got.Status != "done" || got.Reconciliation != nil || got.Blocked != nil {
		t.Fatalf("consistent finished run: %s %+v %+v", got.Status, got.Reconciliation, got.Blocked)
	}
}
