package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
)

func TestEffectDispatchProjection(t *testing.T) {
	for _, tc := range []struct {
		name     string
		revoked  bool
		wantHTTP int
		wantApps int
		want     adopt.State
	}{
		{name: "granted", wantHTTP: http.StatusOK, wantApps: 1, want: adopt.StateCompleted},
		{name: "revoked", revoked: true, wantHTTP: http.StatusBadRequest, want: adopt.StateAdopted},
	} {
		t.Run(tc.name, func(t *testing.T) {
			e := newAR04(t, true)
			s := e.open()
			p, r := e.proposeAndApprove(s)
			job, err := s.LeaseAs("worker", "model-a")
			if err != nil || job == nil {
				t.Fatalf("lease: %v", err)
			}
			if err := s.StepPut(r.ID, job.Token, "related", "verified evidence"); err != nil {
				t.Fatal(err)
			}
			if _, err := s.Judgment(r.ID, job.Token, "judge", `{"level":"grant-50000","evidence_steps":["related"],"rationale":"grant before revocation"}`); err != nil {
				t.Fatal(err)
			}
			if _, _, err := s.StepGet(r.ID, job.Token, "submit"); err != nil {
				t.Fatal(err)
			}
			if tc.revoked {
				pkg := p.scenario()
				grants := pkg.CallerHolds
				t.Cleanup(func() { pkg.CallerHolds = grants })
				pkg.CallerHolds = nil
			}
			s.bank.AcceptDelay = time.Millisecond
			s.bank.TerminalDelay = time.Millisecond
			mux := http.NewServeMux()
			s.routes(mux)
			payload, err := json.Marshal(map[string]string{"run_id": r.ID, "token": job.Token, "step": "submit"})
			if err != nil {
				t.Fatal(err)
			}
			out := httptest.NewRecorder()
			mux.ServeHTTP(out, httptest.NewRequest(http.MethodPost, "/exec/effect", bytes.NewReader(payload)))
			s.bank.mu.Lock()
			apps := len(s.bank.apps)
			s.bank.mu.Unlock()
			state := s.machines[r.ID].State(r.DecisionOp)
			if out.Code != tc.wantHTTP || apps != tc.wantApps || state != tc.want {
				t.Fatalf("effect: HTTP %d, applications %d, state %s, response %s", out.Code, apps, state, out.Body)
			}
			if !tc.revoked {
				if !strings.Contains(out.Body.String(), "RECEIPT=APPROVED") {
					t.Fatalf("authorized effect has no approved receipt: %s", out.Body)
				}
				return
			}

			reason := "live grant set no longer covers"
			if !strings.Contains(out.Body.String(), reason) || r.Receipt != "" || r.ReceiptOf != "" {
				t.Fatalf("revoked effect: response %s, receipt %q/%q", out.Body, r.Receipt, r.ReceiptOf)
			}
			if err := s.Done(r.ID, job.Token, nil, "", out.Body.String()); err != nil {
				t.Fatal(err)
			}
			stored := e.open().st.Runs[r.ID]
			if submit := stored.Steps["submit"]; submit.Status != "failed" || submit.Activity != "派发被拒" || submit.FinishedAt == "" {
				t.Fatalf("persisted submit step: status=%q activity=%q finished_at=%q", submit.Status, submit.Activity, submit.FinishedAt)
			}
			graphResponse := httptest.NewRecorder()
			mux.ServeHTTP(graphResponse, httptest.NewRequest(http.MethodGet, "/api/graph?run="+r.ID, nil))
			var graph Graph
			if err := json.Unmarshal(graphResponse.Body.Bytes(), &graph); graphResponse.Code != http.StatusOK || err != nil || graph.Run == nil || graph.Run.Status != "failed" {
				t.Fatalf("graph: HTTP %d, error %v, response %s", graphResponse.Code, err, graphResponse.Body)
			}
			var submit GraphNode
			for _, node := range graph.Nodes {
				if node.ID == "submit" {
					submit = node
					break
				}
			}
			if submit.Status != "failed" || submit.Activity != "派发被拒" {
				t.Fatalf("projected submit node: status=%q, activity=%q", submit.Status, submit.Activity)
			}
			page := httptest.NewRecorder()
			mux.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/ui/run/"+r.ID, nil))
			if page.Code != http.StatusOK || !strings.Contains(page.Body.String(), "/api/graph?run="+r.ID) {
				t.Fatalf("run page does not use the graph projection: HTTP %d", page.Code)
			}
			audit, err := os.ReadFile(filepath.Join(e.dir, "runs", r.ID, "adopt.jsonl"))
			if err != nil || !strings.Contains(string(audit), reason) || !strings.Contains(string(audit), `"status":"rejected"`) || strings.Contains(string(audit), `"status":"dispatched"`) {
				t.Fatalf("revocation audit: %v, %s", err, audit)
			}
		})
	}
}
