package main

import (
	"fmt"
	"sort"
	"strings"

	"github.com/openjiuwen/agent-festival/go/internal/rfc8785"
)

func h03Canonical(value any) string {
	raw, err := rfc8785.Canonicalize(jsonDoc(value))
	if err != nil {
		return ""
	}
	return string(raw)
}

func (s *Service) h03RunPlan(r *Run) *Plan {
	if r == nil || r.PlanVersion < 1 {
		return nil
	}
	versions := s.st.Plans[r.PlanID]
	if r.PlanVersion > len(versions) {
		return nil
	}
	return versions[r.PlanVersion-1]
}

func h03FieldsMatch(fields []H03Field, specs []H03FieldSpec) bool {
	if len(fields) != len(specs) {
		return false
	}
	expected, ok := h03SpecMap(specs)
	if !ok {
		return false
	}
	last := ""
	for _, f := range fields {
		spec, exists := expected[f.Name]
		if !exists || spec != f.H03FieldSpec || f.Type != "decimal" || !h03DecimalValid(f.Value) || f.Name <= last {
			return false
		}
		last = f.Name
	}
	return true
}

func (s *Service) h03Input(r *Run, p *Plan, node string) (H03BoundInput, string, error) {
	var empty H03BoundInput
	if err := s.h03Ready(p, false); err != nil {
		return empty, "", err
	}
	if r.PlanHash != p.Hash || planHash(p, s.cat.Hashes) != p.Hash {
		return empty, "", errCallConflict
	}
	a := p.agent(node)
	contract, ok := p.ExecutionBinding.NodeContracts[node]
	if a == nil || !ok {
		return empty, "", errCallConflict
	}
	input := H03BoundInput{SchemaVersion: "hub-bound-input/v1", ExecutionBindingSHA256: h03RawHash(p.ExecutionBinding), NodeContractSHA256: h03RawHash(contract), PlanHash: p.Hash, RunID: r.ID, NodeExecutionID: sha(r.ID, p.Key(), node, "iteration:0"), NodePath: node, Attempt: 1, Fields: []H03Field{}, Dependencies: []H03Dependency{}}
	deps := map[string]H03Dependency{}
	for _, binding := range p.ExecutionBinding.InputBindings {
		if binding.ToNode != node {
			continue
		}
		var field *H03Field
		switch binding.SourceKind {
		case "material":
			var pin string
			for _, m := range p.ExecutionBinding.Materials {
				if m.ID == binding.SourceID {
					pin = m.SHA256
					break
				}
			}
			raw, err := s.h03Material(pin)
			if err != nil {
				return empty, "", err
			}
			var doc H03MaterialDocument
			if h03Strict(raw, &doc, 65536) != nil {
				return empty, "", errCallConflict
			}
			for _, f := range doc.Fields {
				if f.Name == binding.SourceField {
					copy := f
					field = &copy
					break
				}
			}
			deps["material:"+binding.SourceID] = H03Dependency{Kind: "material", ID: binding.SourceID, SHA256: "sha256:" + pin}
		case "step":
			up := r.Steps[binding.SourceID]
			if up == nil {
				return empty, "", errCallConflict
			}
			out, err := s.h03FormalOutput(r, p, up)
			if err != nil {
				return empty, "", err
			}
			for _, f := range out.Fields {
				if f.Name == binding.SourceField {
					copy := f
					field = &copy
					break
				}
			}
			deps["step:"+binding.SourceID] = H03Dependency{Kind: "step", ID: binding.SourceID, NodeExecutionID: up.Call.NodeExecutionID, ArtifactDigest: up.Call.ArtifactDigest}
		default:
			return empty, "", errCallConflict
		}
		if field == nil {
			return empty, "", fmt.Errorf("required source field is not committed")
		}
		field.Name = binding.ToField
		input.Fields = append(input.Fields, *field)
	}
	sort.Slice(input.Fields, func(i, j int) bool { return input.Fields[i].Name < input.Fields[j].Name })
	keys := []string{}
	for k := range deps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		input.Dependencies = append(input.Dependencies, deps[k])
	}
	if !h03FieldsMatch(input.Fields, contract.InputFields) {
		return empty, "", fmt.Errorf("source semantics do not match the adopted node contract")
	}
	if len(h03Canonical(input)) > 65536 {
		return empty, "", fmt.Errorf("input size limit")
	}
	return input, a.Task, nil
}

func (s *Service) h03ValidateOutput(p *Plan, step *StepRec, raw string) (H03NodeOutput, error) {
	var out H03NodeOutput
	c := step.Call
	if c == nil || h03Strict([]byte(raw), &out, 65536) != nil || raw != h03Canonical(out) {
		return out, fmt.Errorf("canonical typed formal output required")
	}
	contract, ok := p.ExecutionBinding.NodeContracts[step.ID]
	if !ok || out.SchemaVersion != "hub-node-output/v1" || out.InputDigest != c.InputDigest || out.ExecutionBindingSHA256 != h03RawHash(p.ExecutionBinding) || out.NodeContractSHA256 != h03RawHash(contract) || !h03FieldsMatch(out.Fields, contract.OutputFields) {
		return out, fmt.Errorf("formal output contract/input mismatch")
	}
	return out, nil
}

func (s *Service) h03CallInputValid(r *Run, p *Plan, st *StepRec) error {
	input, goal, err := s.h03Input(r, p, st.ID)
	if err != nil {
		return err
	}
	c := st.Call
	if c == nil || c.Input != h03Canonical(input) || c.Goal != goal || c.NodeExecutionID != input.NodeExecutionID || c.Attempt != 1 {
		return errCallConflict
	}
	d, err := canonicalHash(map[string]any{"input": c.Input, "goal": goal, "binding": c.Binding, "model": s.runModel(r)})
	if err != nil || d != c.InputDigest {
		return errCallConflict
	}
	return nil
}

func (s *Service) h03FormalOutput(r *Run, p *Plan, st *StepRec) (H03NodeOutput, error) {
	var empty H03NodeOutput
	c := st.Call
	if c == nil || st.Status != "done" || c.Status != "completed" || c.AgentStartedAt == "" || c.RemoteTaskID == "" || c.ContextID == "" {
		return empty, fmt.Errorf("upstream formal output is not committed")
	}
	if err := s.h03CallInputValid(r, p, st); err != nil {
		return empty, err
	}
	prefix := "[step:" + st.ID + "]\n"
	if !strings.HasPrefix(st.Output, prefix) {
		return empty, errCallConflict
	}
	raw := strings.TrimPrefix(st.Output, prefix)
	if sha(raw) != c.ArtifactDigest {
		return empty, errCallConflict
	}
	return s.h03ValidateOutput(p, st, raw)
}

func (s *Service) h03CommitCheck(r *Run, p *Plan, st *StepRec, in CallCommand) error {
	if err := s.h03CallInputValid(r, p, st); err != nil {
		return err
	}
	if st.Call.AgentStartedAt == "" {
		return fmt.Errorf("role never accepted its scoped context")
	}
	_, err := s.h03ValidateOutput(p, st, in.Output)
	return err
}

func (s *Service) h03AgentView(p *Plan, st *StepRec) map[string]any {
	c := st.Call
	return map[string]any{"expert": st.Expert, "tools": []string{}, "node_contract": p.ExecutionBinding.NodeContracts[st.ID], "call": map[string]any{"node_execution_id": c.NodeExecutionID, "graph_revision": c.GraphRevision, "node_path": c.NodePath, "attempt": c.Attempt, "binding": c.Binding, "input_digest": c.InputDigest, "input": c.Input, "goal": c.Goal, "request_id": c.RequestID, "context_id": c.ContextID}}
}

func (s *Service) h03DoneCheck(r *Run, p *Plan, skipped []string) error {
	if len(skipped) != 0 {
		return fmt.Errorf("all profile nodes are required; skipped is forbidden")
	}
	if len(r.Steps) != len(p.Agents) {
		return errCallConflict
	}
	for _, a := range p.Agents {
		st := r.Steps[a.ID]
		if st == nil {
			return errCallConflict
		}
		if _, err := s.h03FormalOutput(r, p, st); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) h03Consumption(r *Run, p *Plan, node string) map[string]any {
	contract, ok := p.ExecutionBinding.NodeContracts[node]
	if !ok {
		return nil
	}
	view := map[string]any{"status": "pending", "execution_binding_sha256": h03RawHash(p.ExecutionBinding), "node_contract_sha256": h03RawHash(contract), "input_fields": []string{}, "output_fields": []string{}}
	in, out := []string{}, []string{}
	for _, f := range contract.InputFields {
		in = append(in, f.Name)
	}
	for _, f := range contract.OutputFields {
		out = append(out, f.Name)
	}
	view["input_fields"], view["output_fields"] = in, out
	if r != nil {
		if st := r.Steps[node]; st != nil && st.Call != nil {
			c := st.Call
			view["status"] = "input_bound"
			view["input_digest"] = c.InputDigest
			var input H03BoundInput
			if h03Strict([]byte(c.Input), &input, 65536) == nil {
				view["dependencies"] = input.Dependencies
			}
			// A revoked/disabled execution permission does not erase an already
			// persisted formal result. Project historical commit integrity,
			// separately from whether another execution is currently allowed.
			raw := strings.TrimPrefix(st.Output, "[step:"+st.ID+"]\n")
			_, outputErr := s.h03ValidateOutput(p, st, raw)
			if st.Status == "done" && c.Status == "completed" && c.AgentStartedAt != "" && sha(raw) == c.ArtifactDigest && outputErr == nil {
				view["status"] = "committed"
				view["artifact_digest"] = c.ArtifactDigest
			}
		}
	}
	return view
}

func (s *Service) h03Verify(runID string, tamper bool) (*Report, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r := s.st.Runs[runID]
	p := s.h03RunPlan(r)
	if r == nil || p == nil || p.ExecutionBinding == nil {
		return nil, fmt.Errorf("unknown profile run")
	}
	rep := &Report{RunID: runID, OK: true, Summary: "Engineering field-consumption verification only; business quality, independent Gold and D2 are NOT_RUN."}
	add := func(c Check) { rep.Checks = append(rep.Checks, c); rep.OK = rep.OK && c.OK }
	err := s.h03Ready(p, false)
	add(Check{Name: "adopted field closure", OK: err == nil, Detail: "fixed scenario/material/field identity and current permission"})
	add(Check{Name: "original approved Plan hash", OK: !tamper && planHash(p, s.cat.Hashes) == p.Hash && r.PlanHash == p.Hash, Detail: p.Hash})
	err = s.h03DoneCheck(r, p, nil)
	add(Check{Name: "all required formal node results", OK: err == nil && r.Status == "done", Detail: "Original Call/StepRec identities and NUL artifact digests; not a financial correctness claim."})
	return rep, nil
}
