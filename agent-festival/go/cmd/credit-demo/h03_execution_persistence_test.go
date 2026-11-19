package main

import (
	"errors"
	"testing"
)

func TestH03ExecutionIntentPersistenceFailureNeverReturnsSend(t *testing.T) {
	s, _, r, reads := h03ExecutionFixture(t)
	s.writeState = func(string, any) error { return errors.New("component injected storage failure") }
	in := CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}
	if decision, err := s.BeginCall(in); !errors.Is(err, errStorage) || decision != nil {
		t.Fatalf("storage failure gave send intent: %+v %v", decision, err)
	}
	if s.st.Runs[r.ID].Steps["sum"].Call != nil {
		t.Fatal("dirty in-memory call survived failed save")
	}
	before := reads.Load()
	if decision, err := s.BeginCall(in); !errors.Is(err, errStorage) || decision != nil || reads.Load() != before {
		t.Fatal("frozen service rediscovered/sent after failure")
	}
}

func TestH03ExecutionFormalPersistenceFailureCannotAdvance(t *testing.T) {
	s, p, r, reads := h03ExecutionFixture(t)
	in := CallCommand{RunID: r.ID, Token: r.Token, Step: "sum"}
	decision, err := s.BeginCall(in)
	if err != nil {
		t.Fatal(err)
	}
	c := decision["call"].(*RemoteCall)
	if _, err = s.AgentContext(CallCommand{RunID: r.ID, Step: "sum", Token: c.CapabilityToken, ServiceID: c.Binding.AgentID, ServiceVersion: c.Binding.Version, ContextID: c.ContextID}); err != nil {
		t.Fatal(err)
	}
	in = callIdentity(in, c)
	in.Status = "waiting"
	in.Sequence = 1
	if err = s.UpdateCall(in); err != nil {
		t.Fatal(err)
	}
	out := H03NodeOutput{SchemaVersion: "hub-node-output/v1", ExecutionBindingSHA256: h03RawHash(p.ExecutionBinding), NodeContractSHA256: h03RawHash(p.ExecutionBinding.NodeContracts["sum"]), InputDigest: c.InputDigest, Fields: []H03Field{}}
	for _, f := range p.ExecutionBinding.NodeContracts["sum"].OutputFields {
		value := "100"
		if f.Name == "reference" {
			value = "95"
		}
		out.Fields = append(out.Fields, H03Field{f, value})
	}
	in.Output = h03Canonical(out)
	s.writeState = func(string, any) error { return errors.New("component injected formal storage failure") }
	if _, err = s.CommitCall(in); !errors.Is(err, errStorage) {
		t.Fatalf("formal write failure not surfaced %v", err)
	}
	if s.st.Runs[r.ID].Steps["sum"].Status == "done" || s.st.Runs[r.ID].Steps["sum"].Output != "" {
		t.Fatal("uncommitted formal output survived")
	}
	before := reads.Load()
	if _, err = s.BeginCall(CallCommand{RunID: r.ID, Token: r.Token, Step: "reconcile"}); !errors.Is(err, errStorage) || reads.Load() != before {
		t.Fatal("successor advanced after formal storage failure")
	}
	if err = s.Done(r.ID, r.Token, nil, "fake complete", ""); !errors.Is(err, errStorage) {
		t.Fatal("storage failure completed run")
	}
}
