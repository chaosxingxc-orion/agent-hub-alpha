package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestPGHumanReviewDoesNotRewriteCheckerAndRevocationStillWins(t *testing.T) {
	ctx, s, cfg, cat := integration(t)
	_, g := sample(t)
	g.Nodes = append(g.Nodes, Node{ID: "check", Expert: "checker@1", Role: "program", Task: "isolated checker fixture", DependsOn: []string{"draft"}})
	task, err := s.Propose(ctx, "planner", "human-path", g, "compiler-fixture")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Admit(ctx, task.ID, task.GraphDigest, "test-reviewer"); err != nil {
		t.Fatal(err)
	}
	v, err := s.ClaimRun(ctx, task.ID, "test-worker", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease := leaseFrom(t, v)
	files := Files{Root: t.TempDir()}
	var report, check, digest string
	for _, node := range g.Nodes {
		v, err = s.ResolveNode(ctx, task.ID, node.ID, lease.Owner, lease.Epoch, "", cat, "fixed-source")
		if err != nil {
			t.Fatal(err)
		}
		permit := v.(map[string]any)
		var raw []byte
		if node.ID == "check" {
			raw, _ = json.Marshal(map[string]any{"checker_version": "control-fixture/1", "object_digest": digest, "outcome": "HUMAN", "failures": []string{}, "human_review": []string{"fixture; not business acceptance"}})
		} else {
			raw, _ = json.Marshal(map[string]string{"fixture": node.ID})
		}
		artifact, err := files.Put(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.Commit(ctx, task.ID, node.ID, CommitRequest{lease.Owner, lease.Epoch, permit["input_digest"].(string), permit["attempt"].(int), artifact}, raw, cat, "fixed-source"); err != nil {
			t.Fatal(err)
		}
		if node.ID == "draft" {
			report = artifact
			digest = "sha256:" + artifact
		}
		if node.ID == "check" {
			check = artifact
		}
	}
	completion, err := s.Complete(ctx, task.ID, lease.Owner, lease.Epoch)
	if err != nil {
		t.Fatal(err)
	}
	if completion.(map[string]any)["status"] != "WAIT" {
		t.Fatal("unresolved HUMAN completed")
	}
	if _, err = s.ApproveReport(ctx, task.ID, report, "test-reviewer"); err == nil {
		t.Fatal("HUMAN bypassed")
	}
	in := CheckReview{CheckArtifact: check, ObjectDigest: digest, Decision: "ACCEPT", Reason: "CONTROL TEST ONLY: simulate a qualified reviewer decision; not a real business signature"}
	server := &Server{Store: s, Config: cfg, Catalog: cat, Files: files}
	readReview := func() map[string]any {
		r := httptest.NewRequest("GET", "/v1/workbench/tasks/"+task.ID, nil)
		r.Header.Set("Authorization", "Bearer "+cfg.Tokens["operator"])
		w := httptest.NewRecorder()
		server.Routes().ServeHTTP(w, r)
		if w.Code != 200 {
			t.Fatalf("workbench projection rejected: %d", w.Code)
		}
		var out map[string]any
		if err := json.Unmarshal(w.Body.Bytes(), &out); err != nil {
			t.Fatal(err)
		}
		return out
	}
	if readReview()["human_review"] == nil {
		t.Fatal("HUMAN task has no trusted fallback in the HTTP projection")
	}
	if _, err = s.ReviewCheck(ctx, task.ID, "check", in, "test-reviewer"); err != nil {
		t.Fatal(err)
	}
	if readReview()["human_review"] != nil {
		t.Fatal("resolved check still offers pending manual review")
	}
	in.Reason = "changed explanation is a new fact and cannot replace the original review"
	_, err = s.ReviewCheck(ctx, task.ID, "check", in, "test-reviewer")
	requireCode(t, err, "INPUT_CONFLICT")
	state, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	var original map[string]any
	json.Unmarshal(state.Nodes["check"].Result.Output, &original)
	if original["outcome"] != "HUMAN" || effectiveCheck(state.Nodes["check"]) != "PASS" {
		t.Fatal("checker fact and human authority conflated")
	}
	if _, err = s.SetPause(ctx, task.ID, false); err != nil {
		t.Fatal(err)
	}
	v, err = s.ClaimRun(ctx, task.ID, "test-worker-after-review", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease = leaseFrom(t, v)
	if _, err = s.Complete(ctx, task.ID, lease.Owner, lease.Epoch); err != nil {
		t.Fatal(err)
	}
	if _, err = s.ApproveReport(ctx, task.ID, report, "test-reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Deliver(ctx, task.ID, report, "human-delivery-fixture", "test-reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Invalidate(ctx, check, "withdraw checker basis", "test-reviewer"); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Deliver(ctx, task.ID, report, "new-delivery-after-withdrawal", "test-reviewer"); err == nil {
		t.Fatal("withdrawn checker allowed delivery")
	}
}

// Simulates the waiting update sent by the worker immediately after the
// selector discovers that its conditional predecessor has not settled.
func TestPGHumanSelectorResumesAfterEarlyWaitingUpdate(t *testing.T) {
	ctx, s, _, cat := integration(t)
	_, graph := sample(t)
	graph.Nodes = append(graph.Nodes,
		Node{ID: "check", Expert: "checker@1", Task: "Initial review", DependsOn: []string{"draft"}},
		Node{ID: "check-revision", Expert: "checker@1", Task: "Conditional review", DependsOn: []string{"draft"}, Condition: &Condition{Step: "check", Outcome: "REVISE"}},
		Node{ID: "select", Expert: "select@1", Task: "Select approved report", DependsOn: []string{"check", "check-revision"}},
	)
	task, err := s.Propose(ctx, "planner", "human-early-select", graph, "compiler-fixture")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Admit(ctx, task.ID, task.GraphDigest, "reviewer"); err != nil {
		t.Fatal(err)
	}
	claim, err := s.ClaimRun(ctx, task.ID, "original-worker", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease := leaseFrom(t, claim)
	files := Files{Root: t.TempDir()}
	commit := func(node string, raw []byte) string {
		t.Helper()
		resolved, err := s.ResolveNode(ctx, task.ID, node, lease.Owner, lease.Epoch, "", cat, "fixed-source")
		if err != nil {
			t.Fatal(err)
		}
		permit := resolved.(map[string]any)
		if permit["status"] != "GRANTED" {
			t.Fatalf("%s: expected a new permit, got %v", node, permit["status"])
		}
		artifact, err := files.Put(raw)
		if err != nil {
			t.Fatal(err)
		}
		request := CommitRequest{lease.Owner, lease.Epoch, permit["input_digest"].(string), permit["attempt"].(int), artifact}
		if _, err = s.Commit(ctx, task.ID, node, request, raw, cat, "fixed-source"); err != nil {
			t.Fatal(err)
		}
		return artifact
	}
	for _, node := range []string{"extract", "business"} {
		raw, _ := json.Marshal(map[string]string{"fixture": node})
		commit(node, raw)
	}
	report := commit("draft", []byte(`{"fixture":"draft"}`))
	digest := "sha256:" + report
	raw, _ := json.Marshal(map[string]any{"checker_version": "control-fixture/1", "object_digest": digest, "outcome": "HUMAN", "failures": []string{}})
	check := commit("check", raw)
	_, err = s.ResolveNode(ctx, task.ID, "select", lease.Owner, lease.Epoch, "", cat, "fixed-source")
	requireCode(t, err, "WAIT")
	if _, err = s.MarkWaiting(ctx, task.ID, "select", lease.Owner, lease.Epoch, "UNKNOWN", "WAIT"); err != nil {
		t.Fatal(err)
	}
	paused, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	if paused.Status != "WAITING_INPUT" {
		t.Fatalf("early selector update replaced the human wait: %s", paused.Status)
	}
	if _, err = s.Complete(ctx, task.ID, lease.Owner, lease.Epoch); err != nil {
		t.Fatal(err)
	}
	_, err = s.Heartbeat(ctx, task.ID, lease.Owner, lease.Epoch, "graph")
	requireCode(t, err, "LEASE_EXPIRED")
	_, err = s.ClaimRun(ctx, task.ID, "early-worker", "graph")
	requireCode(t, err, "POLICY_REJECTED")
	if _, err = s.ApproveReport(ctx, task.ID, report, "reviewer"); err == nil {
		t.Fatal("unreviewed draft was approved")
	}
	in := CheckReview{CheckArtifact: check, ObjectDigest: digest, Decision: "ACCEPT", Reason: "Fixture reviewer verified the cited evidence"}
	if _, err = s.ReviewCheck(ctx, task.ID, "check", in, "reviewer"); err != nil {
		t.Fatal(err)
	}
	// Persist the exact node projection left by older workers. No remote
	// call was created for this local selector and the human fact is intact.
	if _, err = s.Pool.Exec(ctx, "UPDATE tasks SET document=jsonb_set(document,'{nodes,select,status}',to_jsonb('UNKNOWN'::text)) WHERE scope=$1 AND id=$2", s.Scope, task.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = s.SetPause(ctx, task.ID, false); err != nil {
		t.Fatal(err)
	}
	claim, err = s.ClaimRun(ctx, task.ID, "resumed-worker", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease = leaseFrom(t, claim)
	if _, err = s.AuthorizeSkip(ctx, task.ID, "check-revision", lease.Owner, lease.Epoch); err != nil {
		t.Fatal(err)
	}
	commit("select", []byte(`{"fixture":"selected"}`))
	finished, err := s.Complete(ctx, task.ID, lease.Owner, lease.Epoch)
	if err != nil || finished.(map[string]string)["status"] != "AWAITING_ACCEPTANCE" {
		t.Fatalf("resumed run did not finish: %v %v", finished, err)
	}
	calls, err := s.Calls(ctx, task.ID, "select")
	if err != nil || len(calls) != 0 {
		t.Fatalf("local selector unexpectedly acquired a remote call: %d %v", len(calls), err)
	}
}
