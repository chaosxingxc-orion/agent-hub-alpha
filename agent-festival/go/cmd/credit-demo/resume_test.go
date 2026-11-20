package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func resumeRequest(s *Service, method, body string) *httptest.ResponseRecorder {
	mux := http.NewServeMux()
	s.routes(mux)
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(method, "/api/resume", strings.NewReader(body)))
	return w
}

func TestResumeAPIRejectsWithoutMutation(t *testing.T) {
	t.Setenv("CREDIT_DEMO_PASSWORD", "fixture-password")
	for _, tc := range []struct {
		name, method, body string
		code               int
	}{
		{"method", "GET", "", 405},
		{"malformed", "POST", "{", 400},
		{"oversized", "POST", strings.Repeat("x", 4097), 400},
		{"missing-password", "POST", `{"run":"R1","user":"alice"}`, 401},
		{"wrong-password", "POST", `{"run":"R1","user":"alice","password":"wrong"}`, 401},
		{"unknown-user", "POST", `{"run":"R1","user":"other","password":"fixture-password"}`, 401},
		{"running", "POST", `{"run":"R1","user":"alice","password":"fixture-password"}`, 409},
		{"unknown-run", "POST", `{"run":"R999","user":"alice","password":"fixture-password"}`, 409},
	} {
		t.Run(tc.name, func(t *testing.T) {
			s, _, _ := busFixture(t)
			before, _ := json.Marshal(s.st)
			w := resumeRequest(s, tc.method, tc.body)
			after, _ := json.Marshal(s.st)
			if w.Code != tc.code || string(before) != string(after) {
				t.Fatalf("code=%d, state changed=%v", w.Code, string(before) != string(after))
			}
		})
	}
}

func TestResumeAPIQueriesOriginalTaskAndPreservesResult(t *testing.T) {
	t.Setenv("CREDIT_DEMO_PASSWORD", "fixture-password")
	s, in, _ := busFixture(t)
	c := beginFixture(t, s, in)
	update := callIdentity(in, c)
	update.Status, update.Sequence = "unknown", 1
	if err := s.UpdateCall(update); err != nil {
		t.Fatal(err)
	}
	s.st.Runs[in.RunID].Steps["completed"] = &StepRec{ID: "completed", Status: "done", Output: "persisted result"}
	if err := s.Done(in.RunID, in.Token, nil, "", "connection lost"); err != nil {
		t.Fatal(err)
	}
	body := `{"run":"R1","user":"alice","password":"fixture-password"}`
	w := resumeRequest(s, "POST", body)
	if w.Code != 200 || s.st.Runs["R1"].Status != "queued" {
		t.Fatal(w.Code, w.Body.String())
	}
	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil || response["run"] != "R1" || response["status"] != "queued" {
		t.Fatal(response, err)
	}
	rev := s.st.Rev
	if w := resumeRequest(s, "POST", body); w.Code != 409 || s.st.Rev != rev {
		t.Fatal("duplicate resume changed queued task")
	}
	job, err := s.Lease("resumed-host")
	if err != nil || job == nil {
		t.Fatal(job, err)
	}
	in.Token = job.Token
	out, err := s.BeginCall(in)
	if err != nil || out["action"] != "query" || out["call"].(*RemoteCall).RequestID != c.RequestID {
		t.Fatal("resume dispatched a new request", out, err)
	}
	if st := s.st.Runs["R1"].Steps["completed"]; st.Status != "done" || st.Output != "persisted result" {
		t.Fatal("resume lost completed result")
	}
}

func TestResumeAPIStorageFailureDoesNotAcknowledgeSuccess(t *testing.T) {
	t.Setenv("CREDIT_DEMO_PASSWORD", "fixture-password")
	s, _, _ := busFixture(t)
	s.st.Runs["R1"].Status = "waiting"
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
	rev := s.st.Rev
	s.writeState = func(string, any) error { return errors.New("injected write failure") }
	w := resumeRequest(s, "POST", `{"run":"R1","user":"alice","password":"fixture-password"}`)
	if w.Code != 503 || s.st.Runs["R1"].Status != "waiting" || s.st.Rev != rev {
		t.Fatal("failed write acknowledged resume", w.Code, w.Body.String())
	}
}
