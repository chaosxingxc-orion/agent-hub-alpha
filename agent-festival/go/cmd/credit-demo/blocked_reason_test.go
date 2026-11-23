package main

import (
	"encoding/json"
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

// olderBuildBlocks does what the candidate before this change did to a
// finished run at startup: keep the pre-reconciliation snapshot, then block
// the run and replace its final output with the reason. withSnapshot false
// drops the snapshot, as if it was lost.
func olderBuildBlocks(t *testing.T, e *ar04Env, s *Service, r *Run, withSnapshot bool) string {
	t.Helper()
	name := "reconciliation-" + strings.TrimPrefix(sha(string(s.lastSaved)), "sha256:") + ".json"
	if withSnapshot {
		if err := os.WriteFile(filepath.Join(e.dir, name), s.lastSaved, 0600); err != nil {
			t.Fatal(err)
		}
	}
	run := s.st.Runs[r.ID]
	run.Status, run.Token, run.Blocked, run.Reconciliation = "blocked", "", nil, nil
	run.Final = "RECONCILIATION_REQUIRED: 采纳账本与快照操作序列不一致；保留原账本与快照，需人工核对，不能重新生成决策。"
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
	return name
}

func dispositions(t *testing.T, e *ar04Env) []reconcileRecord {
	t.Helper()
	paths, _ := filepath.Glob(filepath.Join(e.dir, "disposition-*.json"))
	var out []reconcileRecord
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		var doc struct {
			PreState string            `json:"pre_state"`
			Records  []reconcileRecord `json:"records"`
		}
		if err := json.Unmarshal(raw, &doc); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(filepath.Join(e.dir, doc.PreState)); err != nil {
			t.Fatalf("disposition without its pre-reconciliation snapshot: %v", err)
		}
		out = append(out, doc.Records...)
	}
	return out
}

func hasRecord(recs []reconcileRecord, run, from, to, disposition, source string) bool {
	for _, x := range recs {
		if x.Run == run && x.From == from && x.To == to && x.Disposition == disposition && strings.HasPrefix(x.Source, source) && x.Note != "" {
			return true
		}
	}
	return false
}

// First upgrade: the finished run is flagged, and the flag is also kept as a
// disposition record beside the pre-reconciliation snapshot.
func TestReconciliationRecordsDisposition(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _ := adoptedRun(t, e, s, true)
	makeLegacy(t, s, r)
	s = e.open()
	if rc := s.st.Runs[r.ID].Reconciliation; rc.Disposition != "FLAGGED" || rc.Source != "runs/"+r.ID+"/adopt.jsonl" {
		t.Fatalf("flag: %+v", rc)
	}
	if recs := dispositions(t, e); len(recs) != 1 || !hasRecord(recs, r.ID, "done", "done", "FLAGGED", "runs/") || recs[0].Cause != "LEGACY_ADOPTION_PROTOCOL" {
		t.Fatalf("records: %+v", recs)
	}
}

// Second upgrade: a finished run the older build blocked gets its original
// status and final output back from the pre-reconciliation snapshot.
func TestReconciliationRestoresRunBlockedByOlderBuild(t *testing.T) {
	e := newAR04(t, true)
	s := e.open()
	r, _ := adoptedRun(t, e, s, true)
	makeLegacy(t, s, r)
	snapshot := olderBuildBlocks(t, e, s, r, true)
	s = e.open()
	got := s.st.Runs[r.ID]
	if got.Status != "done" || got.Final != "final credit report" || got.Blocked != nil {
		t.Fatalf("not restored: status=%s final=%q blocked=%+v", got.Status, got.Final, got.Blocked)
	}
	rc := got.Reconciliation
	if rc == nil || rc.Cause != "LEGACY_ADOPTION_PROTOCOL" || rc.Disposition != "RESTORED" || rc.Source != snapshot {
		t.Fatalf("reconciliation: %+v", rc)
	}
	if g := graphRunOf(t, s, got); g.Blocked != nil || g.Reconciliation == nil || g.Reconciliation.Disposition != "RESTORED" {
		t.Fatalf("graph: blocked=%+v reconciliation=%+v", g.Blocked, g.Reconciliation)
	}
	if !hasRecord(dispositions(t, e), r.ID, "blocked", "done", "RESTORED", snapshot) {
		t.Fatalf("records: %+v", dispositions(t, e))
	}
	if next, err := s.LeaseAs("worker", "model-a"); err != nil || next != nil {
		t.Fatalf("restored history leased: %v %v", next, err)
	}
	if err := s.ResumeRun(r.ID); err == nil {
		t.Fatal("restored history resumed")
	}
	rev := s.st.Rev
	if again := e.open(); again.st.Rev != rev {
		t.Fatal("repeated restart rewrote the restored run")
	}
}

// Without original evidence the run stays blocked and is marked for review.
func TestReconciliationKeepsOlderBlockWithoutEvidence(t *testing.T) {
	for _, tc := range []struct {
		name, want string
		snapshot   bool
		change     func(*Run)
	}{
		{"no snapshot", "未找到对账前快照", false, nil},
		{"changed after snapshot", "之后任务记录有其他变化", true, func(r *Run) { r.Steps["related"].Output = "edited" }},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newAR04(t, true)
			s := e.open()
			r, _ := adoptedRun(t, e, s, true)
			makeLegacy(t, s, r)
			olderBuildBlocks(t, e, s, r, tc.snapshot)
			if tc.change != nil {
				tc.change(s.st.Runs[r.ID])
				if err := s.save(); err != nil {
					t.Fatal(err)
				}
			}
			s = e.open()
			got := s.st.Runs[r.ID]
			if got.Status != "blocked" || !strings.HasPrefix(got.Final, "RECONCILIATION_REQUIRED") {
				t.Fatalf("restored without evidence: %s %q", got.Status, got.Final)
			}
			expectBlocked(t, s, got, "RECONCILIATION_REQUIRED", "LEGACY_ADOPTION_PROTOCOL")
			if b := graphRunOf(t, s, got).Blocked; b.Disposition != "PENDING_REVIEW" || !strings.Contains(b.Note, tc.want) {
				t.Fatalf("blocked: %+v", b)
			}
			if !hasRecord(dispositions(t, e), r.ID, "blocked", "blocked", "PENDING_REVIEW", "") {
				t.Fatalf("records: %+v", dispositions(t, e))
			}
			rev := s.st.Rev
			if again := e.open(); again.st.Rev != rev {
				t.Fatal("repeated restart rewrote the pending run")
			}
		})
	}
}

// Two pre-reconciliation snapshots of the same run, one where it had finished
// and one where it was still running: the restore cites the finished one
// whichever is scanned first (the order follows the file names).
func TestReconciliationRestoreCitesFinishedSnapshotInEitherOrder(t *testing.T) {
	for _, tc := range []struct{ finished, running string }{
		{"reconciliation-0-finished.json", "reconciliation-1-running.json"},
		{"reconciliation-1-finished.json", "reconciliation-0-running.json"},
	} {
		t.Run(tc.finished, func(t *testing.T) {
			e := newAR04(t, true)
			s := e.open()
			r, job := adoptedRun(t, e, s, false)
			running := append([]byte(nil), s.lastSaved...)
			if err := s.Done(r.ID, job.Token, nil, "final credit report", ""); err != nil {
				t.Fatal(err)
			}
			makeLegacy(t, s, r)
			finished := append([]byte(nil), s.lastSaved...)
			olderBuildBlocks(t, e, s, r, false)
			for name, raw := range map[string][]byte{tc.finished: finished, tc.running: running} {
				if err := os.WriteFile(filepath.Join(e.dir, name), raw, 0600); err != nil {
					t.Fatal(err)
				}
			}
			s = e.open()
			got := s.st.Runs[r.ID]
			if got.Status != "done" || got.Final != "final credit report" {
				t.Fatalf("not restored: %s %q", got.Status, got.Final)
			}
			if rc := got.Reconciliation; rc == nil || rc.Disposition != "RESTORED" || rc.Source != tc.finished {
				t.Fatalf("restore cites %+v, want source %s", rc, tc.finished)
			}
			if !hasRecord(dispositions(t, e), r.ID, "blocked", "done", "RESTORED", tc.finished) {
				t.Fatalf("records: %+v", dispositions(t, e))
			}
		})
	}
}
