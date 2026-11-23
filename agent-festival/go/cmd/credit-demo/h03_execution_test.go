package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

// Component fixtures only. They do not substitute for T313's original Java,
// Runtime and Bus engineering acceptance.
func h03ExecutionFixture(t *testing.T) (*Service, *Plan, *Run, *atomic.Int64) {
	t.Helper()
	s, p, _ := h03BoundFixture(t)
	reads := &atomic.Int64{}
	rdc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reads.Add(1)
		if r.Header.Get("Authorization") != "Bearer component-registry" || r.Header.Get("X-Tenant-ID") != "component" {
			w.WriteHeader(401)
			return
		}
		id := strings.Split(r.URL.Path, "/")[3]
		if strings.HasSuffix(r.URL.Path, "/instances") {
			_ = json.NewEncoder(w).Encode(map[string]any{"instances": []any{map[string]any{"instanceId": "component-instance", "health": "ONLINE", "contractVersion": "1.0", "capabilityVersion": "0.1.0"}}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"name": id, "version": "0.1.0", "skills": []string{"engineering-component"}})
	}))
	t.Cleanup(rdc.Close)
	s.cfg.BusRegistryURL = rdc.URL
	s.cfg.BusRegistryToken = "component-registry"
	s.cfg.BusTenant = "component"
	if err := s.bindPlan(p); err != nil {
		t.Fatal(err)
	}
	p.Hash = planHash(p, s.cat.Hashes)
	p.RunID = "R1"
	p.Precheck = precheck(p)
	if !p.Precheck.OK {
		t.Fatalf("fixture precheck %+v", p.Precheck)
	}
	binding := *p.Binding
	r := &Run{ID: "R1", PlanID: p.ID, PlanVersion: p.Version, PlanHash: p.Hash, Binding: &binding, Status: "running", Token: "lease-component", Epoch: 1, Steps: map[string]*StepRec{}, ApprovedBy: "alice"}
	for _, l := range p.Layers {
		for _, id := range l {
			r.Order = append(r.Order, id)
			r.Steps[id] = &StepRec{ID: id, Expert: p.agent(id).Expert, Status: "pending"}
		}
	}
	s.st.Runs[r.ID] = r
	if err := s.save(); err != nil {
		t.Fatal(err)
	}
	return s, p, r, reads
}

func h03CompleteComponent(t *testing.T, s *Service, p *Plan, r *Run, node string, values map[string]string) CallCommand {
	t.Helper()
	in := CallCommand{RunID: r.ID, Token: r.Token, Step: node}
	decision, err := s.BeginCall(in)
	if err != nil {
		t.Fatal(err)
	}
	if decision["action"] != "send" {
		t.Fatalf("expected one first intent, got %+v", decision)
	}
	c := decision["call"].(*RemoteCall)
	if _, err = s.AgentContext(CallCommand{RunID: r.ID, Step: node, Token: c.CapabilityToken, ServiceID: c.Binding.AgentID, ServiceVersion: c.Binding.Version, ContextID: c.ContextID}); err != nil {
		t.Fatal(err)
	}
	in = callIdentity(in, c)
	in.RemoteTaskID = "component-task-" + node
	in.Status = "waiting"
	in.Sequence = 1
	if err = s.UpdateCall(in); err != nil {
		t.Fatal(err)
	}
	out := H03NodeOutput{SchemaVersion: "hub-node-output/v1", ExecutionBindingSHA256: h03RawHash(p.ExecutionBinding), NodeContractSHA256: h03RawHash(p.ExecutionBinding.NodeContracts[node]), InputDigest: c.InputDigest, Fields: []H03Field{}}
	for _, f := range p.ExecutionBinding.NodeContracts[node].OutputFields {
		out.Fields = append(out.Fields, H03Field{f, values[f.Name]})
	}
	in.Output = h03Canonical(out)
	if _, err = s.CommitCall(in); err != nil {
		t.Fatal(err)
	}
	return in
}

func TestH03ExecutionAuthoritativeInputsFormalChainAndProjection(t *testing.T) {
	s, p, r, _ := h03ExecutionFixture(t)
	if _, err := s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "reconcile"}); err == nil {
		t.Fatal("uncommitted upstream consumed")
	}
	first := h03CompleteComponent(t, s, p, r, "sum", map[string]string{"reference": "95", "total": "100"})
	input, _, err := s.h03Input(r, p, "reconcile")
	if err != nil {
		t.Fatal(err)
	}
	if input.Fields[0].Value != "95" || input.Fields[1].Value != "100" || len(input.Dependencies) != 1 || input.Dependencies[0].ArtifactDigest != sha(first.Output) {
		t.Fatalf("formal fields/provenance not consumed: %+v", input)
	}
	h03CompleteComponent(t, s, p, r, "reconcile", map[string]string{"difference": "5"})
	if err = s.Done(r.ID, r.Token, nil, "component completed", ""); err != nil {
		t.Fatal(err)
	}
	rep, err := s.Verify(r.ID, false)
	if err != nil || !rep.OK || !strings.Contains(rep.Summary, "D2 are NOT_RUN") {
		t.Fatalf("engineering verification %+v %v", rep, err)
	}
	g, err := s.graph(p, true)
	if err != nil {
		t.Fatal(err)
	}
	raw, _ := json.Marshal(g)
	if strings.Contains(string(raw), "capability_token") || strings.Contains(string(raw), "\"value\"") || strings.Contains(string(raw), s.cfg.H03RuntimeCredential) {
		t.Fatalf("private payload leaked: %s", raw)
	}
	for _, n := range g.Nodes {
		if n.Consumption["status"] != "committed" || n.Output != "" {
			t.Fatalf("false projection %+v", n)
		}
	}
}

func TestH03LeasePreservesBoundModelAndStructuredProfile(t *testing.T) {
	s, _, r, reads := h03ExecutionFixture(t)
	r.LeaseUntil = time.Now().Add(-time.Minute)
	r.Status = "queued"
	bound := s.runModel(r)
	before := reads.Load()
	job, err := s.LeaseAs("engineering", bound+"-different")
	if err != nil || job != nil || reads.Load() != before {
		t.Fatalf("mismatched executor must not receive a lease or discover capabilities: job=%v err=%v", job, err)
	}
	job, err = s.LeaseAs("engineering", bound)
	if err != nil || job == nil {
		t.Fatalf("bound executor did not receive a job: %v", err)
	}
	if job.Model != bound || job.ExecutionProfile == "" || job.Members["sum"]["node_contract"] == nil {
		t.Fatalf("lease lost its model or structured profile: %+v", job)
	}
}

func TestH03ExecutionUnknownRetainsOriginalIdentity(t *testing.T) {
	s, _, r, reads := h03ExecutionFixture(t)
	in := CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}
	a, err := s.BeginCall(in)
	if err != nil {
		t.Fatal(err)
	}
	c := a["call"].(*RemoteCall)
	before := reads.Load()
	b, err := s.BeginCall(in)
	if err != nil || b["action"] != "unknown" || b["call"].(*RemoteCall).RequestID != c.RequestID || reads.Load() != before {
		t.Fatalf("unknown redispatched: %+v %v", b, err)
	}
	id := callIdentity(in, c)
	id.Status = "unknown"
	id.Sequence = 1
	if err = s.UpdateCall(id); err != nil {
		t.Fatal(err)
	}
	b, err = s.BeginCall(in)
	if err != nil || b["action"] != "query" || b["call"].(*RemoteCall).RemoteTaskID != id.RemoteTaskID {
		t.Fatal("original task query lost", b, err)
	}
	if err = s.Done(r.ID, r.Token, nil, "pretend done", ""); err != nil {
		t.Fatal(err)
	}
	if s.st.Runs[r.ID].Status != "waiting" {
		t.Fatal("unknown was marked successful")
	}
}

func TestH03ExecutionRejectsWorkerInputAndResultMutation(t *testing.T) {
	s, p, r, reads := h03ExecutionFixture(t)
	before := reads.Load()
	if _, err := s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "sum", Input: "forged"}); err == nil || reads.Load() != before {
		t.Fatal("caller forged input dispatched")
	}
	in := h03CompleteComponent(t, s, p, r, "sum", map[string]string{"reference": "95", "total": "100"})
	before = reads.Load()
	if _, err := s.CommitCall(in); err != nil {
		t.Fatal("same result must reuse", err)
	}
	in.Output = strings.Replace(in.Output, "\"100\"", "\"101\"", 1)
	if _, err := s.CommitCall(in); err == nil {
		t.Fatal("committed result overwritten")
	}
	b, err := s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"})
	if err != nil || b["action"] != "reuse" || reads.Load() != before {
		t.Fatal("formal result was resent", b, err)
	}
}

func TestH03ExecutionHistoricalCommitSurvivesDisabledExecution(t *testing.T) {
	s, p, r, _ := h03ExecutionFixture(t)
	h03CompleteComponent(t, s, p, r, "sum", map[string]string{"reference": "95", "total": "100"})
	s.cfg.H03Profile = nil
	view := s.h03Consumption(r, p, "sum")
	if view["status"] != "committed" || view["artifact_digest"] == "" {
		t.Fatal("disabled permission erased historical commit projection")
	}
	if _, err := s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}); err == nil {
		t.Fatal("disabled profile remained executable")
	}
}
