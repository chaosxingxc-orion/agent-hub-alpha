package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// Knowledge reuse provenance (#55 analog, A10 / ALK-08): the reuse key must
// use the model the run is bound to (AR-04), not the service's current
// default, and a reused step must say durably which result, capability,
// package, sources, model and output it took.

// stepsOn executes a run with an executor on the given model, the way the
// Java executor drives the service.
func stepsOn(t *testing.T, s *Service, r *Run, executorModel string) taskRun {
	t.Helper()
	job, err := s.LeaseAs("executor-"+executorModel, executorModel)
	if err != nil || job == nil || job.RunID != r.ID {
		t.Fatalf("lease %s on %s: %+v %v", r.ID, executorModel, job, err)
	}
	tr := taskRun{run: r, recorded: map[string]bool{}}
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
		if err := s.StepPut(r.ID, job.Token, id, "[step:"+id+"]\n"+id+" computed in "+r.ID+" on "+executorModel); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Done(r.ID, job.Token, nil, "fixture report", ""); err != nil {
		t.Fatal(err)
	}
	return tr
}

func proposeApprove(t *testing.T, s *Service, q SelectionQuery) *Run {
	t.Helper()
	rec := mustSelect(t, s, q)
	p, err := s.Propose(ProposeInput{SelectionID: rec.ID})
	if err != nil {
		t.Fatal(err)
	}
	r, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// A run bound to model-b must not reuse a result computed on model-a, even
// when the service's default model has meanwhile gone back to model-a; and
// what it computes itself must be published as a model-b result.
func TestKnowledgeReuseFollowsTheRunsBoundModel(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	q := arQuery("经营", "资产质量")

	s.cfg.Model = "model-a"
	t1 := runTask(t, s, q, nil)
	k1 := t1.run.Steps["extract"].Published
	if k1 == "" {
		t.Fatal("first task published nothing")
	}

	s.cfg.Model = "model-b" // the run is adopted under model-b
	r2 := proposeApprove(t, s, q)
	if r2.Binding == nil || r2.Binding.Model != "model-b" {
		t.Fatalf("run 2 binding: %+v", r2.Binding)
	}
	s.cfg.Model = "model-a" // the service default goes back before the run executes
	t2 := stepsOn(t, s, r2, "model-b")
	ext2 := t2.run.Steps["extract"]
	if t2.recorded["extract"] || ext2.ReusedFrom != "" {
		t.Fatalf("a model-b run reused %s, computed on model-a (lookup %+v)", ext2.ReusedFrom, ext2.Lookup)
	}
	k2 := ext2.Published
	if k2 == "" {
		t.Fatal("run 2 published nothing")
	}

	t3 := runTask(t, s, q, nil) // bound to model-a
	if got := t3.run.Steps["extract"].ReusedFrom; got != k1 {
		t.Fatalf("a model-a run reused %q, want the model-a result %q (not the model-b result %q)", got, k1, k2)
	}
}

// After a reuse, the persisted state alone says what was reused: the step
// carries the item, capability and package versions, source digests, model
// and output digest; the item carries its model and output digest.
func TestKnowledgeReuseProvenanceIsDurable(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	s.cfg.Model = "model-a"
	q := arQuery("经营", "资产质量")
	t1 := runTask(t, s, q, nil)
	t2 := runTask(t, s, q, nil)
	k1 := t1.run.Steps["extract"].Published
	if t2.run.Steps["extract"].ReusedFrom != k1 {
		t.Fatalf("setup: no reuse of %s", k1)
	}

	raw, err := os.ReadFile(filepath.Join(s.cfg.DataDir, "state.json"))
	if err != nil {
		t.Fatal(err)
	}
	var persisted struct {
		Runs      map[string]map[string]any `json:"runs"`
		Knowledge map[string]map[string]any `json:"knowledge"`
	}
	if err := json.Unmarshal(raw, &persisted); err != nil {
		t.Fatal(err)
	}
	item := persisted.Knowledge[k1]
	if item == nil {
		t.Fatalf("item %s not persisted", k1)
	}
	if item["model"] != "model-a" || item["output_digest"] != sha(t1.run.Steps["extract"].Output) {
		t.Fatalf("persisted item lacks model/output digest: model=%v output_digest=%v", item["model"], item["output_digest"])
	}
	steps, _ := persisted.Runs[t2.run.ID]["steps"].(map[string]any)
	step, _ := steps["extract"].(map[string]any)
	lookup, _ := step["knowledge_lookup"].(map[string]any)
	if lookup == nil {
		t.Fatal("reusing step has no persisted lookup")
	}
	want := map[string]any{"item": k1, "capability": item["capability"], "package": item["package"],
		"model": "model-a", "output_digest": item["output_digest"]}
	for k, v := range want {
		if lookup[k] != v || v == nil || v == "" {
			t.Errorf("persisted lookup %s = %v, want %v", k, lookup[k], v)
		}
	}
	srcs, _ := lookup["sources"].([]any)
	if len(srcs) == 0 {
		t.Errorf("persisted lookup has no source digests")
	} else if m, _ := srcs[0].(map[string]any); m["sha256"] == nil || m["sha256"] == "" {
		t.Errorf("persisted lookup source without digest: %v", srcs[0])
	}
}
