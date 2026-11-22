package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestPGHumanReviewDoesNotRewriteCheckerAndRevocationStillWins(t *testing.T) {
	ctx, s, cfg, cat := integration(t)
	_, g := reportSample(t)
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
	for _, node := range g.Nodes[:len(g.Nodes)-1] {
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
	_, err = s.ResolveNode(ctx, task.ID, "select", lease.Owner, lease.Epoch, "", cat, "fixed-source")
	requireCode(t, err, "WAIT")
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
	v, err = s.ResolveNode(ctx, task.ID, "select", lease.Owner, lease.Epoch, "", cat, "fixed-source")
	if err != nil {
		t.Fatal(err)
	}
	permit := v.(map[string]any)
	selected, _ := json.Marshal(map[string]string{"artifact_digest": digest})
	selection, err := files.Put(selected)
	if err != nil {
		t.Fatal(err)
	}
	request := CommitRequest{lease.Owner, lease.Epoch, permit["input_digest"].(string), permit["attempt"].(int), selection}
	if _, err = s.Commit(ctx, task.ID, "select", request, selected, cat, "fixed-source"); err != nil {
		t.Fatal(err)
	}
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
	for _, decision := range []string{"ACCEPT", "REVISE", "REJECT"} {
		t.Run(decision, func(t *testing.T) {
			verifyHumanBranch(t, decision)
		})
	}
}

func verifyHumanBranch(t *testing.T, decision string) {
	t.Helper()
	ctx, s, _, cat := integration(t)
	_, graph := sample(t)
	graph.Nodes = append(graph.Nodes,
		Node{ID: "check", Expert: "checker@1", Task: "Initial review", DependsOn: []string{"draft"}},
		Node{ID: "revision", Expert: "writer@1", Task: "Revise only when requested", DependsOn: []string{"draft", "check"}, Condition: &Condition{Step: "check", Outcome: "REVISE"}},
		Node{ID: "check-revision", Expert: "checker@1", Task: "Check revised report", DependsOn: []string{"revision", "check"}, Condition: &Condition{Step: "check", Outcome: "REVISE"}},
		Node{ID: "select", Expert: "select@1", Task: "Select approved report", DependsOn: []string{"check", "check-revision"}},
	)
	if err := ValidateReportGraph(graph, cat); err != nil {
		t.Fatal(err)
	}
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
	_, err = s.ResolveNode(ctx, task.ID, "revision", lease.Owner, lease.Epoch, "", cat, "fixed-source")
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
	// REJECT also covers a reviewer acting before the worker closes its lease.
	if decision != "REJECT" {
		if _, err = s.Complete(ctx, task.ID, lease.Owner, lease.Epoch); err != nil {
			t.Fatal(err)
		}
		_, err = s.Heartbeat(ctx, task.ID, lease.Owner, lease.Epoch, "graph")
		requireCode(t, err, "LEASE_EXPIRED")
		_, err = s.ClaimRun(ctx, task.ID, "early-worker", "graph")
		requireCode(t, err, "POLICY_REJECTED")
	}
	if _, err = s.ApproveReport(ctx, task.ID, report, "reviewer"); err == nil {
		t.Fatal("unreviewed draft was approved")
	}
	in := CheckReview{CheckArtifact: check, ObjectDigest: digest, Decision: decision, Reason: "Fixture reviewer verified the cited evidence"}
	wrong := in
	wrong.ObjectDigest = "sha256:another-report"
	_, err = s.ReviewCheck(ctx, task.ID, "check", wrong, "reviewer")
	requireCode(t, err, "INPUT_CONFLICT")
	if _, err = s.ReviewCheck(ctx, task.ID, "check", in, "reviewer"); err != nil {
		t.Fatal(err)
	}
	if decision == "REJECT" {
		_, err = s.Heartbeat(ctx, task.ID, lease.Owner, lease.Epoch, "graph")
		requireCode(t, err, "LEASE_EXPIRED")
		if _, err = s.SetPause(ctx, task.ID, false); err != nil {
			t.Fatal(err)
		}
		_, err = s.ClaimRun(ctx, task.ID, "resumed-worker", "graph")
		requireCode(t, err, "POLICY_REJECTED")
		state, err := s.GetTask(ctx, task.ID)
		if err != nil || state.Status != "NEEDS_REVIEW" || state.Nodes["check"].Review == nil || state.Nodes["check"].Review.Reason != in.Reason {
			t.Fatal("rejection did not block the original graph or preserve the review", err)
		}
		if _, err = s.ApproveReport(ctx, task.ID, report, "reviewer"); err == nil {
			t.Fatal("rejected report was approved")
		}
		return
	}
	if decision == "ACCEPT" {
		// Preserve the historical local UNKNOWN selector case.
		if _, err = s.Pool.Exec(ctx, "UPDATE tasks SET document=jsonb_set(document,'{nodes,select,status}',to_jsonb('UNKNOWN'::text)) WHERE scope=$1 AND id=$2", s.Scope, task.ID); err != nil {
			t.Fatal(err)
		}
	}
	if _, err = s.SetPause(ctx, task.ID, false); err != nil {
		t.Fatal(err)
	}
	claim, err = s.ClaimRun(ctx, task.ID, "resumed-worker", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease = leaseFrom(t, claim)
	if decision == "REVISE" {
		_, err = s.AuthorizeSkip(ctx, task.ID, "revision", lease.Owner, lease.Epoch)
		requireCode(t, err, "POLICY_REJECTED")
		state, err := s.GetTask(ctx, task.ID)
		if err != nil {
			t.Fatal(err)
		}
		binding, err := s.Binding(ctx, state, "revision", cat, "fixed-source")
		if err != nil || len(binding.Reviews) != 1 || binding.Reviews[0].Decision != "REVISE" || binding.Reviews[0].ObjectDigest != digest {
			t.Fatal("revision did not consume the authorized review of the original draft", err)
		}
		revised := commit("revision", []byte(`{"fixture":"revised draft"}`))
		rechecked, _ := json.Marshal(map[string]any{"checker_version": "control-fixture/1", "object_digest": "sha256:" + revised, "outcome": "PASS", "failures": []string{}})
		checkArtifact := commit("check-revision", rechecked)
		stale := in
		stale.CheckArtifact = checkArtifact
		_, err = s.ReviewCheck(ctx, task.ID, "check-revision", stale, "reviewer")
		requireCode(t, err, "INPUT_CONFLICT")
		report = revised
	} else {
		_, err = s.ResolveNode(ctx, task.ID, "revision", lease.Owner, lease.Epoch, "", cat, "fixed-source")
		requireCode(t, err, "POLICY_REJECTED")
		for _, node := range []string{"revision", "check-revision"} {
			if _, err = s.AuthorizeSkip(ctx, task.ID, node, lease.Owner, lease.Epoch); err != nil {
				t.Fatal(err)
			}
		}
	}
	commit("select", []byte(`{"fixture":"selected"}`))
	finished, err := s.Complete(ctx, task.ID, lease.Owner, lease.Epoch)
	if err != nil || finished.(map[string]string)["status"] != "AWAITING_ACCEPTANCE" {
		t.Fatalf("resumed run did not finish: %v %v", finished, err)
	}
	if _, err = s.ApproveReport(ctx, task.ID, report, "reviewer"); err != nil {
		t.Fatal(err)
	}
	calls, err := s.Calls(ctx, task.ID, "select")
	if err != nil || len(calls) != 0 {
		t.Fatalf("local selector unexpectedly acquired a remote call: %d %v", len(calls), err)
	}
}
