package main

import (
	"errors"
	"testing"
)

func TestPreviewBusProjectionCarriesDurableRevision(t *testing.T) {
	s, in, _ := busFixture(t)
	c := beginFixture(t, s, in)
	p := s.st.Plans["P1"][0]
	p.RunID = "R1"
	p.Steps = []Step{{ID: "report", Member: "report"}}
	g, err := s.graph(p, false)
	if err != nil || g.Rev != s.st.Rev || g.Bus == nil || g.Bus.PlannedNodes != 1 {
		t.Fatalf("revision and Bus summary must coexist: %#v, %v", g, err)
	}
	if len(g.Nodes) != 1 || g.Nodes[0].Execution.Kind != "agent_bus" || g.Nodes[0].Remote["node_execution_id"] != c.NodeExecutionID {
		t.Fatalf("missing real call association: %#v", g.Nodes)
	}
}

func TestPreviewFailedSaveDoesNotPublishNewRevision(t *testing.T) {
	s, _, _ := busFixture(t)
	before := s.st.Rev
	s.writeState = func(string, any) error { return errors.New("injected storage failure") }
	s.st.Runs["R1"].Status = "done"
	if err := s.save(); err == nil {
		t.Fatal("write failure must be returned")
	}
	if s.st.Rev != before || s.st.Runs["R1"].Status == "done" {
		t.Fatal("failed write published a new revision or completed state")
	}
	if err := s.save(); err == nil || s.st.Rev != before {
		t.Fatal("frozen storage must not advance revision")
	}
}
