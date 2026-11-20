package main

import (
	"encoding/json"
	"testing"
)

func TestPGDelayedDeliveryUsesOwnPermitAndRevocationStopsReuse(t *testing.T) {
	ctx, s, cfg, cat := integration(t)
	_, g := sample(t)
	g.Nodes = append(g.Nodes, Node{ID: "check", Expert: "checker@1", Role: "程序检查", Task: "检查实际报告", DependsOn: []string{"draft"}})
	task, err := s.Propose(ctx, "planner", "delivery-test", g, "runtime-digest")
	if err != nil {
		t.Fatal(err)
	}
	if _, err = s.Admit(ctx, task.ID, task.GraphDigest, "reviewer"); err != nil {
		t.Fatal(err)
	}
	v, err := s.ClaimRun(ctx, task.ID, "worker", "graph")
	if err != nil {
		t.Fatal(err)
	}
	lease := leaseFrom(t, v)
	files := Files{t.TempDir()}
	var reportArtifact, reportDigest string
	for _, node := range g.Nodes {
		v, err = s.ResolveNode(ctx, task.ID, node.ID, lease.Owner, lease.Epoch, "", cat, "fixed-material")
		if err != nil {
			t.Fatal(err)
		}
		permit := v.(map[string]any)
		var raw []byte
		if node.ID == "check" {
			raw, _ = json.Marshal(map[string]any{"outcome": "PASS", "object_digest": reportDigest, "checker_version": "isolated-control-test/1", "scope": "not-a-business-verification"})
		} else {
			raw, _ = json.Marshal(map[string]string{"node": node.ID, "fixture": "control-protocol-only"})
		}
		artifact, err := files.Put(raw)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = s.Commit(ctx, task.ID, node.ID, CommitRequest{lease.Owner, lease.Epoch, permit["input_digest"].(string), permit["attempt"].(int), artifact}, raw, cat, "fixed-material"); err != nil {
			t.Fatal(err)
		}
		if node.ID == "draft" {
			reportArtifact = artifact
			reportDigest = "sha256:" + artifact
		}
	}
	if _, err = s.Complete(ctx, task.ID, lease.Owner, lease.Epoch); err != nil {
		t.Fatal(err)
	}
	if _, err = s.Deliver(ctx, task.ID, reportArtifact, "one-delivery", "reviewer"); err == nil {
		t.Fatal("delivery without user approval accepted")
	}
	approval, err := s.ApproveReport(ctx, task.ID, reportArtifact, "reviewer")
	if err != nil {
		t.Fatal(err)
	}
	approved, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	repeated, err := s.ApproveReport(ctx, task.ID, reportArtifact, "reviewer")
	if err != nil {
		t.Fatal(err)
	}
	if approval.(map[string]string)["approval_id"] != repeated.(map[string]string)["approval_id"] {
		t.Fatal("repeated approval changed receipt")
	}
	afterApprovalRetry, err := s.GetTask(ctx, task.ID)
	if err != nil || afterApprovalRetry.Generation != approved.Generation || len(afterApprovalRetry.ReportApprovals) != 1 {
		t.Fatal("repeated approval changed task or added an approval", err)
	}
	first, err := s.Deliver(ctx, task.ID, reportArtifact, "one-delivery", "reviewer")
	if err != nil {
		t.Fatal(err)
	}
	delivered, err := s.GetTask(ctx, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	restarted, err := OpenStore(ctx, cfg.DSN, s.Scope, 30)
	if err != nil {
		t.Fatal(err)
	}
	defer restarted.Pool.Close()
	for _, retry := range []struct {
		store   *Store
		command string
	}{{s, "one-delivery"}, {s, "another-tab"}, {restarted, "another-tab"}} {
		receipt, err := retry.store.Deliver(ctx, task.ID, reportArtifact, retry.command, "reviewer")
		if err != nil {
			t.Fatal(err)
		}
		var previous map[string]any
		if err := json.Unmarshal(receipt.(json.RawMessage), &previous); err != nil {
			t.Fatal(err)
		}
		if previous["receipt_id"] != first.(map[string]any)["receipt_id"] || previous["epoch"] != "1" {
			t.Fatal("delivery retry changed original receipt or epoch", retry.command)
		}
	}
	afterDeliveryRetries, err := s.GetTask(ctx, task.ID)
	if err != nil || afterDeliveryRetries.Generation != delivered.Generation || len(afterDeliveryRetries.DeliveryReceipts) != 1 {
		t.Fatal("delivery retries changed task or added a publication", err)
	}
	_, err = s.getTestEffect(ctx, nil, "another-tab")
	requireCode(t, err, "INPUT_CONFLICT")
	_, err = s.Deliver(ctx, task.ID, reportArtifact, "another-tab", "different-user")
	requireCode(t, err, "APPROVAL_REQUIRED")
	var domain string
	if err = s.Pool.QueryRow(ctx, "SELECT domain FROM run_leases WHERE scope=$1 AND task_id=$2 AND owner='go-publisher'", s.Scope, task.ID).Scan(&domain); err != nil || domain != "delivery" {
		t.Fatal("publisher reused graph permit")
	}
	if _, err = s.Invalidate(ctx, reportArtifact, "isolated corruption test", "operator"); err != nil {
		t.Fatal(err)
	}
	for _, command := range []string{"one-delivery", "another-tab", "new-after-withdrawal"} {
		if _, err = s.Deliver(ctx, task.ID, reportArtifact, command, "reviewer"); err == nil {
			t.Fatal("revoked report delivered through cached receipt", command)
		}
	}
}

func TestTaskReadGrantCannotAuthorizeAnotherTaskOrSurviveTampering(t *testing.T) {
	s := &Server{Config: Config{Tokens: map[string]string{"planner": "test-only-key-with-sufficient-length"}}}
	grant := s.viewGrant("T-one")
	if !s.allowView("T-one", grant) {
		t.Fatal("valid one-task grant rejected")
	}
	if s.allowView("T-two", grant) || s.allowView("T-one", grant+"tamper") {
		t.Fatal("read grant scope/signature not enforced")
	}
}
