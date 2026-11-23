package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestH03ExecutionRuntimeAndRoleIdentityBoundary(t *testing.T) {
	s, _, r, reads := h03ExecutionFixture(t)
	mux := http.NewServeMux()
	s.routes(mux)
	handler := s.h03ControlHandler(mux)
	for _, path := range []string{"/exec/lease", "/exec/heartbeat", "/exec/call/begin", "/exec/call/update", "/exec/call/commit", "/exec/done"} {
		for _, credential := range []string{"", "role-token", "wrong-runtime"} {
			req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(`{"run_id":"R1","step":"sum","token":"lease-component"}`))
			req.Header.Set(h03RuntimeHeader, credential)
			response := httptest.NewRecorder()
			before := reads.Load()
			rev := s.st.Rev
			handler.ServeHTTP(response, req)
			if response.Code != 401 || s.st.Rev != rev || reads.Load() != before {
				t.Fatalf("identity mutated state %s %d", path, response.Code)
			}
		}
	}
	for _, path := range []string{"/exec/effect", "/exec/judgment", "/exec/branch", "/member-mcp/intake", "/debug/state", "/agent/model"} {
		req := httptest.NewRequest(http.MethodPost, path, strings.NewReader("{}"))
		req.Header.Set(h03RuntimeHeader, s.cfg.H03RuntimeCredential)
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, req)
		if response.Code != 404 {
			t.Fatalf("disallowed route reachable %s %d", path, response.Code)
		}
	}
	decision, err := s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"})
	if err != nil {
		t.Fatal(err)
	}
	c := decision["call"].(*RemoteCall)
	if _, err = s.BeginCall(CallCommand{RunID: r.ID, Token: c.CapabilityToken, Step: "sum"}); err == nil {
		t.Fatal("role credential became runtime lease")
	}
	view, err := s.AgentContext(CallCommand{RunID: r.ID, Step: "sum", Token: c.CapabilityToken, ServiceID: c.Binding.AgentID, ServiceVersion: c.Binding.Version, ContextID: c.ContextID})
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(view)
	if bytes.Contains(raw, []byte(c.CapabilityToken)) || bytes.Contains(raw, []byte(s.cfg.H03RuntimeCredential)) || bytes.Contains(raw, []byte("capability_token")) {
		t.Fatalf("secret leaked into role context %s", raw)
	}
	if _, err = s.AgentContext(CallCommand{RunID: r.ID, Step: "reconcile", Token: c.CapabilityToken, ServiceID: c.Binding.AgentID, ServiceVersion: c.Binding.Version, ContextID: c.ContextID}); err == nil {
		t.Fatal("role accessed another node")
	}
}

func TestH03ExecutionTenantChangeCannotDispatchOrQuery(t *testing.T) {
	for _, started := range []bool{false, true} {
		t.Run(map[bool]string{false: "before-send", true: "old-task-query"}[started], func(t *testing.T) {
			s, _, r, reads := h03ExecutionFixture(t)
			in := CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}
			if started {
				got, err := s.BeginCall(in)
				if err != nil {
					t.Fatal(err)
				}
				c := got["call"].(*RemoteCall)
				update := callIdentity(in, c)
				update.Sequence = 1
				update.Status = "waiting"
				if err = s.UpdateCall(update); err != nil {
					t.Fatal(err)
				}
			}
			before := reads.Load()
			rev := s.st.Rev
			s.cfg.BusTenant = "same-card-in-another-tenant"
			if _, err := s.BeginCall(in); err == nil || reads.Load() != before || s.st.Rev != rev {
				t.Fatal("tenant drift discovered/dispatched/queried or mutated")
			}
		})
	}
}

func TestH03ExecutionControlledDeploymentDoesNotRunLegacyState(t *testing.T) {
	s, _, _, reads := h03ExecutionFixture(t)
	before := reads.Load()
	rev := s.st.Rev
	if _, err := s.Propose(ProposeInput{CaseID: "C006", Goal: "legacy credit plan"}); err == nil || reads.Load() != before || s.st.Rev != rev {
		t.Fatal("controlled MCP Propose reached legacy compile/adoption")
	}
	legacy := &Plan{ID: "legacy-plan", Version: 1, Hash: "legacy-hash"}
	run := &Run{ID: "legacy-run", PlanID: legacy.ID, PlanVersion: 1, PlanHash: legacy.Hash, Status: "queued", Token: "legacy-token", Steps: map[string]*StepRec{}}
	s.st.Plans[legacy.ID] = []*Plan{legacy}
	s.st.Runs = map[string]*Run{run.ID: run}
	if job, err := s.Lease("anonymous-legacy-worker"); err != nil || job != nil || run.Status != "queued" || s.st.Rev != rev || reads.Load() != before {
		t.Fatal("legacy state was leased or rewritten")
	}
	if _, err := s.Approve(legacy.ID, 1, legacy.Hash, "alice", nil); err == nil || s.st.Rev != rev {
		t.Fatal("legacy approval bypassed isolated profile")
	}
	if err := s.Heartbeat(run.ID, run.Token); err == nil || s.st.Rev != rev {
		t.Fatal("old lease mutated legacy state")
	}
}

func TestH03ExecutionRejectsDuplicateAndTrailingRequestJSON(t *testing.T) {
	s, _, _, reads := h03ExecutionFixture(t)
	mux := http.NewServeMux()
	s.routes(mux)
	handler := s.h03ControlHandler(mux)
	for _, body := range []string{`{"run_id":"R1","run_id":"other","token":"lease-component","step":"sum"}`, `{"run_id":"R1","token":"lease-component","step":"sum"} {}`} {
		req := httptest.NewRequest(http.MethodPost, "/exec/call/begin", strings.NewReader(body))
		req.Header.Set(h03RuntimeHeader, s.cfg.H03RuntimeCredential)
		response := httptest.NewRecorder()
		rev := s.st.Rev
		before := reads.Load()
		handler.ServeHTTP(response, req)
		if response.Code != 400 || s.st.Rev != rev || reads.Load() != before {
			t.Fatal("ambiguous request accepted or mutated state")
		}
	}
}
