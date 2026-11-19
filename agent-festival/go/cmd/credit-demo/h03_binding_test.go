package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/compose"
)

func h03BoundFixture(t *testing.T) (*Service, *Plan, *SelectionRecord) {
	t.Helper()
	compose.RegisterGraph()
	pkg, q, profile := h03TestFixture(t)
	pkg.Governance.State = "approved"
	pkg.Governance.ApprovedBy = "engineering-fixture"
	pkg.Governance.ApprovedAt = "2026-11-17"
	pkg.ParentBounds = map[string]any{"max_cost": map[string]any{"amount": 100, "unit": "USD-cents"}, "max_downstream_calls": 4, "max_seconds": 120}
	for i := range pkg.Capabilities {
		pkg.Capabilities[i].Effects.Level = "none"
		pkg.Capabilities[i].Bounds.Seconds = 30
		pkg.Capabilities[i].BusAgent = "h03-" + pkg.Capabilities[i].Node
	}
	pkg.digest = h03RawHash(pkg)
	profile.Package.Digest = pkg.Digest()
	body, _ := json.Marshal(pkg)
	if err := withScenarioDir(t, map[string]string{"engineering.json": string(body)}); err != nil {
		t.Fatal(err)
	}
	profile.Materials[0].Path = filepath.Join(t.TempDir(), "material.json")
	if err := os.WriteFile(profile.Materials[0].Path, profile.Materials[0].raw, 0600); err != nil {
		t.Fatal(err)
	}
	s := &Service{cfg: Config{DataDir: t.TempDir(), H03Profile: profile, H03RuntimeCredential: strings.Repeat("r", 32), BusTenant: "component", Model: "engineering-no-model", LeaseSeconds: 30}, cat: &Catalog{Hashes: map[string]string{}}, writeState: writeJSON, st: &State{Plans: map[string][]*Plan{}, Runs: map[string]*Run{}, Selections: map[string]*SelectionRecord{}}}
	rec, err := selectCapabilities(scenarios.get(pkg.Package), q, profile)
	if err != nil || rec.Outcome != "graph" {
		t.Fatalf("selection failed %+v %v", rec, err)
	}
	rec.ID = "S1"
	rec.Digest = rec.digest()
	s.st.Selections[rec.ID] = rec
	p := &Plan{ID: "P1", Version: 1, CaseID: "material-a", Scenario: pkg.Package, Agents: rec.Team, Selection: &SelectionRef{ID: rec.ID, Digest: rec.Digest}, Binding: &ExecBinding{JudgmentSpec: "sha256:" + strings.Repeat("a", 64), Model: s.cfg.Model}}
	for _, g := range rec.Graph {
		p.Steps = append(p.Steps, Step{ID: g.Node, Member: g.Node, DependsOn: g.DependsOn})
	}
	p.Layers, err = layers(p.Steps)
	if err != nil {
		t.Fatal(err)
	}
	if err = s.h03Bind(p, rec); err != nil {
		t.Fatal(err)
	}
	p.Hash = planHash(p, s.cat.Hashes)
	s.st.Plans[p.ID] = []*Plan{p}
	return s, p, rec
}

func TestH03BindingFreezesScenarioAndMaterialBytes(t *testing.T) {
	s, p, _ := h03BoundFixture(t)
	if errs := h03BindingErrors(p); len(errs) > 0 {
		t.Fatal(errs)
	}
	before := p.Hash
	name := p.expert("eng.sum").Name
	pkg := scenarios.get(p.Scenario)
	pkg.Capabilities[1].Name = "new catalog name"
	rebuildCatalogData()
	if p.expert("eng.sum").Name != name || planHash(p, map[string]string{"eng.sum": "new-contract"}) != before {
		t.Fatal("old task silently used current catalog")
	}
	if err := os.WriteFile(s.cfg.H03Profile.Materials[0].Path, []byte("new source bytes"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := s.h03Material(p.ExecutionBinding.Materials[0].SHA256); err != nil {
		t.Fatal("original archived bytes lost", err)
	}
	raw, _ := json.Marshal(p)
	var restored Plan
	if err := json.Unmarshal(raw, &restored); err != nil {
		t.Fatal(err)
	}
	if restored.expert("eng.sum").Name != name || planHash(&restored, nil) != before {
		t.Fatal("persisted snapshot did not restore its own semantics")
	}
	if err := s.h03Ready(&restored, false); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(s.h03MaterialPath(p.ExecutionBinding.Materials[0].SHA256)); err != nil {
		t.Fatal(err)
	}
	if err := s.h03Ready(&restored, false); err == nil {
		t.Fatal("missing material snapshot must stop")
	}
}

func TestH03BindingRejectsAdoptionAndShapeDrift(t *testing.T) {
	for _, mode := range []string{"profile-version", "selection", "conditional", "missing-mapping", "contract-set", "upstream-unit", "material-permission", "role-permission"} {
		t.Run(mode, func(t *testing.T) {
			s, p, rec := h03BoundFixture(t)
			switch mode {
			case "profile-version":
				s.cfg.H03Profile.Version = "next"
			case "selection":
				rec.Eligibility[0].Reasons = append(rec.Eligibility[0].Reasons, "tampered")
			case "conditional":
				p.Agents[0].RunIf = &RunIf{Step: "sum", Contains: "yes"}
			case "missing-mapping":
				p.ExecutionBinding.InputBindings = p.ExecutionBinding.InputBindings[1:]
			case "contract-set":
				c := p.ExecutionBinding.NodeContracts["sum"]
				c.InputFields = c.InputFields[:1]
				p.ExecutionBinding.NodeContracts["sum"] = c
			case "upstream-unit":
				c := p.ExecutionBinding.NodeContracts["sum"]
				c.OutputFields[1].Unit = "EUR-million"
				p.ExecutionBinding.NodeContracts["sum"] = c
			case "material-permission":
				s.cfg.H03Profile.Materials = nil
			case "role-permission":
				delete(s.cfg.H03Profile.CandidateContracts, "eng.sum@0.1.0")
			}
			if err := s.h03Ready(p, true); err == nil {
				t.Fatalf("drift accepted: %s", mode)
			}
		})
	}
}

func TestH03PrecheckUsesFrozenFieldAndCardSets(t *testing.T) {
	_, p, _ := h03BoundFixture(t)
	if pc := precheck(p); !pc.OK {
		t.Fatalf("legal finite profile rejected %+v", pc)
	}
	p.ExecutionBinding.InputBindings = nil
	if pc := precheck(p); pc.OK {
		t.Fatal("UI precheck certified missing field mappings")
	}
}

func TestH03BindingDependencySequenceSemantics(t *testing.T) {
	for _, mode := range []string{"nil-to-empty", "empty-to-nil", "nonempty-chain", "root-nonempty", "wrong-predecessor", "different-count", "different-order"} {
		t.Run(mode, func(t *testing.T) {
			_, p, _ := h03BoundFixture(t)
			root := p.agent("sum")
			var first, next *Step
			for i := range p.Steps {
				switch p.Steps[i].ID {
				case "sum":
					first = &p.Steps[i]
				case "reconcile":
					next = &p.Steps[i]
				}
			}
			if root == nil || first == nil || next == nil {
				t.Fatal("expected original two-node fixture")
			}
			switch mode {
			case "nil-to-empty":
				root.DependsOn, first.DependsOn = nil, []string{}
			case "empty-to-nil":
				root.DependsOn, first.DependsOn = []string{}, nil
			case "nonempty-chain":
				// The valid reconcile -> sum sequence remains exact.
			case "root-nonempty":
				first.DependsOn = []string{"reconcile"}
			case "wrong-predecessor":
				next.DependsOn = []string{"other"}
			case "different-count":
				next.DependsOn = []string{"sum", "sum"}
			case "different-order":
				p.agent("reconcile").DependsOn = []string{"sum", "other"}
				next.DependsOn = []string{"other", "sum"}
			}
			errs := h03BindingErrors(p)
			valid := mode == "nil-to-empty" || mode == "empty-to-nil" || mode == "nonempty-chain"
			if valid && len(errs) != 0 {
				t.Fatalf("equivalent dependency sequence rejected: %v", errs)
			}
			if !valid && (len(errs) != 1 || errs[0] != "compiled steps differ from adopted identity chain") {
				t.Fatalf("dependency mismatch bypassed compiled identity gate: %v", errs)
			}
		})
	}
}
