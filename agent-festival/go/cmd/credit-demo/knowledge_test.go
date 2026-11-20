package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A10 / ALK-08 (#16): cross-task reuse of an annual-report indicator node.

type taskRun struct {
	run        *Run
	plan       *Plan
	modelCalls int             // steps the executor had to compute
	recorded   map[string]bool // steps StepGet answered from a record
}

// runTask selects, proposes, approves and executes one task the way the
// Java executor drives the service: StepGet first, compute and StepPut only
// when nothing is recorded.
func runTask(t *testing.T, s *Service, q SelectionQuery, edit func([]PlanAgent)) taskRun {
	t.Helper()
	rec := mustSelect(t, s, q)
	in := ProposeInput{SelectionID: rec.ID}
	if edit != nil {
		in.Agents = append([]PlanAgent(nil), rec.Team...)
		edit(in.Agents)
	}
	p, err := s.Propose(in)
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Lease("fixture-executor")
	if err != nil || job == nil || job.RunID != r.ID {
		t.Fatalf("lease: %+v %v", job, err)
	}
	tr := taskRun{run: r, plan: p, recorded: map[string]bool{}}
	for _, id := range r.Order {
		recorded, _, err := s.StepGet(r.ID, job.Token, id)
		if err != nil {
			t.Fatal(err)
		}
		if recorded {
			tr.recorded[id] = true
			continue
		}
		tr.modelCalls++
		if err := s.StepPut(r.ID, job.Token, id, "[step:"+id+"]\n"+id+" computed in "+r.ID); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Done(r.ID, job.Token, nil, "fixture report", ""); err != nil {
		t.Fatal(err)
	}
	return tr
}

func graphNode(t *testing.T, s *Service, p *Plan, id string) GraphNode {
	t.Helper()
	g, err := s.graph(p, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range g.Nodes {
		if n.ID == id {
			return n
		}
	}
	t.Fatalf("no node %s", id)
	return GraphNode{}
}

// A new task with the same indicator node, inputs and snapshot reuses the
// first task's result: no model call for that node, the source is recorded
// and shown; nodes that are not reusable run as before.
func TestKnowledgeCrossTaskHit(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")

	t1 := runTask(t, s, q, nil)
	ext1 := t1.run.Steps["extract"]
	if ext1.Lookup == nil || ext1.Lookup.Outcome != "miss" || ext1.Lookup.Reason != "没有其他任务消费过这些来源" || ext1.Published == "" {
		t.Fatalf("first consumer: lookup=%+v published=%q", ext1.Lookup, ext1.Published)
	}
	if t1.modelCalls != 4 || t1.run.Steps["operating"].Lookup != nil || t1.run.Steps["operating"].Published != "" {
		t.Fatalf("first task: %d model calls; non-reusable node looked up or published", t1.modelCalls)
	}

	t2 := runTask(t, s, q, nil)
	ext2 := t2.run.Steps["extract"]
	if !t2.recorded["extract"] || t2.modelCalls != 3 {
		t.Fatalf("second task: extract recorded=%v, %d model calls", t2.recorded["extract"], t2.modelCalls)
	}
	if ext2.ReusedFrom != ext1.Published || ext2.Lookup.Outcome != "hit" || ext2.Lookup.Source != t1.run.ID+"/extract" {
		t.Fatalf("second task extract: reused_from=%q lookup=%+v", ext2.ReusedFrom, ext2.Lookup)
	}
	if ext2.Output != ext1.Output || ext2.Published != "" {
		t.Fatalf("reused output %q (published %q), want %q", ext2.Output, ext2.Published, ext1.Output)
	}
	n := graphNode(t, s, t2.plan, "extract")
	if n.Status != "reused" || n.Knowledge == nil || n.Knowledge.Outcome != "hit" || n.Knowledge.Source != t1.run.ID+"/extract" || n.Knowledge.Reason == "" {
		t.Fatalf("projection of the hit: status=%s knowledge=%+v", n.Status, n.Knowledge)
	}
	items := s.Knowledge()
	if len(items) != 1 || items[0].ID != ext1.Published || strings.Join(items[0].ReusedBy, ",") != t2.run.ID+"/extract" {
		t.Fatalf("knowledge list: %+v", items)
	}
}

// A corrected snapshot (same material id, different digest) is a different
// source: the earlier result is not reused, and the reason says why.
func TestKnowledgeSnapshotVersionMiss(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")
	t1 := runTask(t, s, q, nil)

	corrected := strings.Replace(strings.Replace(annualReportPackage,
		"52c49dcdc00d4f5436adabc6561044d7a539e9bb0a8ebbdb692a1da6237459da",
		"abababababababababababababababababababababababababababababababab", 1),
		`"version": "0.1.0"`, `"version": "0.1.1"`, 1)
	pkg, err := parseScenario([]byte(corrected))
	if err != nil {
		t.Fatal(err)
	}
	scenarios.packages[arPackage] = pkg // the operator re-imports the package with the corrected snapshot

	t3 := runTask(t, s, q, nil)
	ext := t3.run.Steps["extract"]
	if t3.recorded["extract"] || ext.ReusedFrom != "" || t3.modelCalls != 4 {
		t.Fatalf("result of the old snapshot reused: recorded=%v reused_from=%q", t3.recorded["extract"], ext.ReusedFrom)
	}
	if ext.Lookup.Outcome != "miss" || ext.Lookup.Item != t1.run.Steps["extract"].Published || !strings.Contains(ext.Lookup.Reason, "同一材料的其他版本") {
		t.Fatalf("version miss: %+v", ext.Lookup)
	}
	if n := graphNode(t, s, t3.plan, "extract"); n.Knowledge == nil || !strings.Contains(n.Knowledge.Reason, "版本不同不复用") {
		t.Fatalf("projection does not show the reason: %+v", n.Knowledge)
	}
	if ext.Published == "" || ext.Published == t1.run.Steps["extract"].Published {
		t.Fatalf("the corrected snapshot's result is not published as its own item: %q", ext.Published)
	}
}

// A revoked result is not reused; runs that reused it before the revocation
// show it; a later result of the same node is reused instead.
func TestKnowledgeRevocation(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")
	t1 := runTask(t, s, q, nil)
	k1 := t1.run.Steps["extract"].Published
	t2 := runTask(t, s, q, nil)
	if t2.run.Steps["extract"].ReusedFrom != k1 {
		t.Fatal("fixture: the second task should reuse the first result")
	}

	if _, err := s.RevokeKnowledge(k1, "alice", ""); err == nil {
		t.Fatal("revocation without a reason accepted")
	}
	if _, err := s.RevokeKnowledge(k1, "alice", "抽取用了母公司口径"); err != nil {
		t.Fatal(err)
	}
	if n := graphNode(t, s, t2.plan, "extract"); n.Knowledge == nil || !strings.Contains(n.Knowledge.Revoked, "抽取用了母公司口径") {
		t.Fatalf("run that reused the revoked result does not show it: %+v", n.Knowledge)
	}

	t4 := runTask(t, s, q, nil)
	ext := t4.run.Steps["extract"]
	if t4.recorded["extract"] || ext.Lookup.Outcome != "miss" || !strings.Contains(ext.Lookup.Reason, k1+" 已撤回") || !strings.Contains(ext.Lookup.Reason, "抽取用了母公司口径") {
		t.Fatalf("revoked result: recorded=%v lookup=%+v", t4.recorded["extract"], ext.Lookup)
	}
	t5 := runTask(t, s, q, nil)
	if got := t5.run.Steps["extract"].ReusedFrom; got != ext.Published || got == k1 {
		t.Fatalf("after revocation the fresh result %s should be reused, got %q", ext.Published, got)
	}

	mux := http.NewServeMux()
	s.knowledgeRoutes(mux)
	post := func(body map[string]string) int {
		raw, _ := json.Marshal(body)
		w := httptest.NewRecorder()
		mux.ServeHTTP(w, httptest.NewRequest(http.MethodPost, "/api/knowledge/revoke", bytes.NewReader(raw)))
		return w.Code
	}
	if code := post(map[string]string{"item": ext.Published, "reason": "x", "user": "alice", "password": "wrong"}); code != http.StatusUnauthorized {
		t.Fatalf("revoke without the password: %d", code)
	}
	if code := post(map[string]string{"item": ext.Published, "reason": "复核发现错页", "user": "alice", "password": envOr("CREDIT_DEMO_PASSWORD", "demo")}); code != http.StatusOK {
		t.Fatalf("revoke with the password: %d", code)
	}
}

// Anything the node's model call reads that differs (here the sub-task the
// planner wrote) means a miss with the reason, never a wrong hit.
func TestKnowledgeDifferentInputMisses(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")
	runTask(t, s, q, nil)
	t2 := runTask(t, s, q, func(agents []PlanAgent) {
		for i := range agents {
			if agents[i].ID == "extract" {
				agents[i].Task += "；另抽取分红指标"
			}
		}
	})
	ext := t2.run.Steps["extract"]
	if t2.recorded["extract"] || ext.Lookup.Outcome != "miss" || !strings.Contains(ext.Lookup.Reason, "来源相同，但指令、子任务、任务输入或上游结果不同") {
		t.Fatalf("different sub-task: recorded=%v lookup=%+v", t2.recorded["extract"], ext.Lookup)
	}
}

// Publishing is part of the step's commit: if the commit cannot be saved,
// no knowledge item survives either.
func TestKnowledgePublishFailsWithTheCommit(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	rec := mustSelect(t, s, arQuery("经营", "资产质量"))
	p, err := s.Propose(ProposeInput{SelectionID: rec.ID})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	job, _ := s.Lease("fixture-executor")
	if _, _, err := s.StepGet(r.ID, job.Token, "extract"); err != nil {
		t.Fatal(err)
	}
	s.writeState = func(string, any) error { return errors.New("disk full") }
	if err := s.StepPut(r.ID, job.Token, "extract", "[step:extract]\nindicators"); err == nil {
		t.Fatal("commit reported success without a durable record")
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.st.Knowledge) != 0 || s.st.Runs[r.ID].Steps["extract"].Status == "done" {
		t.Fatalf("knowledge %v / step %s survived a failed commit", s.st.Knowledge, s.st.Runs[r.ID].Steps["extract"].Status)
	}
}

// TestKnowledgeSamples writes the extract node's lookup in four tasks (first
// consumer, cross-task hit, corrected snapshot, after revocation) and the
// knowledge list when K_SAMPLES_DIR is set (evidence for #16).
func TestKnowledgeSamples(t *testing.T) {
	dir := os.Getenv("K_SAMPLES_DIR")
	if dir == "" {
		t.Skip("K_SAMPLES_DIR not set")
	}
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")
	out := map[string]any{}
	node := func(name string, tr taskRun) {
		out[name] = map[string]any{"run": tr.run.ID, "model_calls": tr.modelCalls, "extract": graphNode(t, s, tr.plan, "extract")}
	}
	t1 := runTask(t, s, q, nil)
	node("1-first-consumer", t1)
	t2 := runTask(t, s, q, nil)
	node("2-cross-task-hit", t2)
	original := scenarios.packages[arPackage]
	corrected := strings.Replace(strings.Replace(annualReportPackage,
		"52c49dcdc00d4f5436adabc6561044d7a539e9bb0a8ebbdb692a1da6237459da",
		"abababababababababababababababababababababababababababababababab", 1),
		`"version": "0.1.0"`, `"version": "0.1.1"`, 1)
	pkg, err := parseScenario([]byte(corrected))
	if err != nil {
		t.Fatal(err)
	}
	scenarios.packages[arPackage] = pkg
	node("3-corrected-snapshot", runTask(t, s, q, nil))
	scenarios.packages[arPackage] = original
	if _, err := s.RevokeKnowledge(t1.run.Steps["extract"].Published, "alice", "抽取用了母公司口径（样例）"); err != nil {
		t.Fatal(err)
	}
	node("4-after-revocation", runTask(t, s, q, nil))
	out["2-cross-task-hit-after-revocation"] = graphNode(t, s, t2.plan, "extract").Knowledge
	out["knowledge"] = s.Knowledge()
	raw, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "k-reuse-extract.json"), append(raw, '\n'), 0o644); err != nil {
		t.Fatal(err)
	}
}
