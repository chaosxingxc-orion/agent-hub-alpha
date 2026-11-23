package main

import "testing"

func TestH03ExecutionRequiredNodesCannotBeSkippedOrFaked(t *testing.T) {
	s, p, r, _ := h03ExecutionFixture(t)
	rev := s.st.Rev
	if err := s.Done(r.ID, r.Token, nil, "fake complete", ""); err == nil || s.st.Rev != rev {
		t.Fatal("pending nodes became complete")
	}
	if err := s.Done(r.ID, r.Token, []string{"sum", "reconcile"}, "fake complete", ""); err == nil || s.st.Rev != rev {
		t.Fatal("skipped required nodes became complete")
	}
	h03CompleteComponent(t, s, p, r, "sum", map[string]string{"reference": "95", "total": "100"})
	if err := s.Done(r.ID, r.Token, nil, "only one node", ""); err == nil {
		t.Fatal("missing required second result")
	}
	if err := s.StepPut(r.ID, r.Token, "reconcile", "legacy bypass"); err == nil {
		t.Fatal("StepPut bypassed remote formal commit")
	}
}

func TestH03ExecutionOutputContractRejectsWrongFactsShape(t *testing.T) {
	for _, mode := range []string{"unit", "period", "basis", "type", "missing", "input-digest", "closure", "empty"} {
		t.Run(mode, func(t *testing.T) {
			s, p, r, _ := h03ExecutionFixture(t)
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
			fields := []H03Field{}
			for _, f := range p.ExecutionBinding.NodeContracts["sum"].OutputFields {
				value := "100"
				if f.Name == "reference" {
					value = "95"
				}
				fields = append(fields, H03Field{f, value})
			}
			out := H03NodeOutput{SchemaVersion: "hub-node-output/v1", ExecutionBindingSHA256: h03RawHash(p.ExecutionBinding), NodeContractSHA256: h03RawHash(p.ExecutionBinding.NodeContracts["sum"]), InputDigest: c.InputDigest, Fields: fields}
			switch mode {
			case "unit":
				out.Fields[0].Unit = "EUR-million"
			case "period":
				out.Fields[0].Period = "FY2024"
			case "basis":
				out.Fields[0].Basis = "parent"
			case "type":
				out.Fields[0].Type = "text"
			case "missing":
				out.Fields = out.Fields[:1]
			case "input-digest":
				out.InputDigest = "wrong"
			case "closure":
				out.ExecutionBindingSHA256 = "wrong"
			}
			in.Output = h03Canonical(out)
			if mode == "empty" {
				in.Output = ""
			}
			rev := s.st.Rev
			if _, err = s.CommitCall(in); err == nil || s.st.Rev != rev || r.Steps["sum"].Status == "done" {
				t.Fatal("invalid output was formally committed")
			}
		})
	}
}
