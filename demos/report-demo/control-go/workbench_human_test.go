package main

import (
	"encoding/json"
	"testing"
)

func TestWorkbenchHumanReviewUsesTrustedTaskPageAndCurrentCheck(t *testing.T) {
	s := &Server{Config: Config{PublicURL: "https://control.example"}}
	task := &Task{ID: "T-fixture", Graph: Graph{Nodes: []Node{{ID: "check"}}}, Nodes: map[string]*NodeState{
		"check": {Result: &Result{ArtifactID: "checker-artifact", Valid: true, Output: json.RawMessage(`{"outcome":"HUMAN","object_digest":"sha256:report"}`)}},
	}}
	review := s.workbenchHumanReview(task)
	if review == nil || review["url"] != "https://control.example/ui/tasks/T-fixture" {
		t.Fatal("missing trusted, credential-free review link")
	}
	checks := review["checks"].([]any)
	if len(checks) != 1 || checks[0].(map[string]string)["check_artifact"] != "checker-artifact" {
		t.Fatal("review link lost checker identity")
	}
	task.Nodes["check"].Result.Valid = false
	if s.workbenchHumanReview(task) != nil {
		t.Fatal("withdrawn check offered review")
	}
	task.Nodes["check"].Result.Valid = true
	task.Nodes["check"].Result.Output = json.RawMessage(`{"outcome":"PASS","object_digest":"sha256:report"}`)
	if s.workbenchHumanReview(task) != nil {
		t.Fatal("PASS offered unnecessary manual review")
	}
}

func TestWorkbenchBlockedReasonHidesUnrecognizedWorkerText(t *testing.T) {
	task := &Task{Status: "BLOCKED", Graph: Graph{Nodes: []Node{{ID: "member"}}}, Nodes: map[string]*NodeState{
		"member": {Status: "BLOCKED", Reason: "private provider error"},
	}}
	if reason := workbenchBlockedReason(task); reason["code"] != "BLOCKED" || reason["message"] == "" || reason["message"] == task.Nodes["member"].Reason {
		t.Fatalf("expected a generic explanation, got %v", reason)
	}
	task.Status = "NEEDS_REVIEW"
	if reason := workbenchBlockedReason(task); reason["code"] != "REVIEW_REQUIRED" {
		t.Fatalf("review reason lost: %v", reason)
	}
	task.Status = "RUNNING"
	if workbenchBlockedReason(task) != nil {
		t.Fatal("active task projected as blocked")
	}
}
