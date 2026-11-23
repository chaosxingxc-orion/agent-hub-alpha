package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"slices"
	"strings"

	"github.com/openjiuwen/agent-festival/go/internal/rfc8785"
	"github.com/openjiuwen/agent-festival/go/internal/validate"
)

func (pkg *ScenarioPackage) expert(id string) *Expert {
	if pkg == nil {
		return nil
	}
	c := pkg.card(id)
	if c == nil {
		return nil
	}
	e := &Expert{ID: c.ID, Name: c.Name, Kind: c.Kind, Summary: c.Summary, Tools: c.Tools, Prompt: c.Instruction, MaxIter: c.MaxIter, Package: pkg.Package, effectLevel: c.Effects.Level, compensate: c.Effects.Compensation, costCents: c.Bounds.CostCents, calls: c.Bounds.Calls, seconds: c.Bounds.Seconds}
	for _, g := range c.Requires {
		e.requires = append(e.requires, g.grant())
	}
	return e
}

func (p *Plan) expert(id string) *Expert {
	if p.ExecutionBinding == nil {
		return expertByID(id)
	}
	return p.scenario().expert(id)
}

func (p *Plan) busAgent(id string) string {
	if p.ExecutionBinding == nil {
		return busAgents[id]
	}
	pkg := p.scenario()
	if pkg == nil {
		return ""
	}
	c := pkg.card(id)
	if c == nil {
		return ""
	}
	return c.BusAgent
}

func (p *Plan) contractHash(id string, current map[string]string) string {
	if p.ExecutionBinding == nil {
		return current[id]
	}
	e := p.expert(id)
	if e == nil {
		return ""
	}
	d, _ := validate.ContractHash(e.contract())
	return d
}

func (s *Service) h03ProfileFor(pkg *ScenarioPackage) *H03Profile {
	p := s.cfg.H03Profile
	if p != nil && pkg != nil && p.Package.ID == pkg.Package {
		return p
	}
	return nil
}

func (s *Service) h03MaterialPath(pin string) string {
	return filepath.Join(s.cfg.DataDir, "h03-materials", pin+".json")
}

func (s *Service) h03ArchiveMaterial(m H03ProfileMaterial) error {
	raw, err := h03Read(m.Path, m.SHA256, 65536)
	if err != nil {
		return err
	}
	path := s.h03MaterialPath(m.SHA256)
	if _, err = os.Stat(path); err == nil {
		_, err = s.h03Material(m.SHA256)
		return err
	} else if !os.IsNotExist(err) {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	// Original durable JSON writer stores an opaque original-byte snapshot;
	// it is evidence referenced by Plan, never an independent task ledger.
	return writeJSON(path, map[string]string{"bytes": base64.StdEncoding.EncodeToString(raw)})
}

func (s *Service) h03Material(pin string) ([]byte, error) {
	if !h03Hex.MatchString(pin) {
		return nil, fmt.Errorf("invalid material pin")
	}
	info, err := os.Lstat(s.h03MaterialPath(pin))
	if err != nil || !info.Mode().IsRegular() || info.Size() > 100000 {
		return nil, fmt.Errorf("material snapshot unavailable")
	}
	raw, err := os.ReadFile(s.h03MaterialPath(pin))
	if err != nil {
		return nil, err
	}
	var doc struct {
		Bytes string `json:"bytes"`
	}
	if h03Strict(raw, &doc, 100000) != nil {
		return nil, fmt.Errorf("invalid material snapshot")
	}
	original, err := base64.StdEncoding.DecodeString(doc.Bytes)
	if err != nil || len(original) > 65536 || h03BytesHash(original) != pin {
		return nil, fmt.Errorf("material snapshot digest mismatch")
	}
	return original, nil
}

func (s *Service) h03Bind(p *Plan, rec *SelectionRecord) error {
	pkg := p.scenario()
	profile := s.h03ProfileFor(pkg)
	if profile == nil {
		return nil
	}
	if rec == nil || rec.Outcome != "graph" || rec.Digest != rec.digest() {
		return fmt.Errorf("profile requires intact original selection")
	}
	if len(p.Agents) > 16 || len(p.Agents) < 2 || len(p.Agents) > 16 || len(p.Agents) != len(rec.Selection) {
		return fmt.Errorf("finite profile requires the complete selected team")
	}
	snapshot, err := rfc8785.Canonicalize(jsonDoc(pkg))
	if err != nil {
		return err
	}
	b := &H03ExecutionBinding{SchemaVersion: "hub-node-binding/v1", BusTenant: s.cfg.BusTenant, Profile: h03ProfileName, ScenarioSnapshot: snapshot, NodeContracts: map[string]H03NodeContract{}}
	for _, id := range rec.Query.Materials {
		found := false
		for _, m := range profile.Materials {
			if m.ID == id {
				if err = s.h03ArchiveMaterial(m); err != nil {
					return err
				}
				b.Materials = append(b.Materials, MaterialRef{ID: m.ID, SHA256: m.SHA256})
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("selected material unavailable")
		}
	}
	for _, a := range p.Agents {
		c := pkg.card(a.Expert)
		if c == nil || !rec.selected(c.ID) || c.BusAgent == "" {
			return fmt.Errorf("every profile member requires its original remote role binding")
		}
		old := rec.eligibility(c.ID)
		if old == nil || old.Result != "pass" || old.Checks["h03.profile"] != profile.Version+"@"+h03ProfileDigest(profile) {
			return fmt.Errorf("selection profile changed; select again")
		}
		el := h03Eligibility(profile, pkg, c, &rec.Query, eligible(pkg, c, &rec.Query))
		if el.Result != "pass" {
			return fmt.Errorf("field eligibility changed: %v", el.Reasons)
		}
		b.NodeContracts[a.ID] = profile.CandidateContracts[c.ID+"@"+c.Version]
	}
	for _, a := range p.Agents {
		c := pkg.card(a.Expert)
		for _, f := range b.NodeContracts[a.ID].InputFields {
			matches := []H03InputBinding{}
			if c.Input.Schema == pkg.Selection.SourceSchema {
				for _, m := range b.Materials {
					raw, e := s.h03Material(m.SHA256)
					if e != nil {
						return e
					}
					var doc H03MaterialDocument
					if h03Strict(raw, &doc, 65536) != nil {
						return fmt.Errorf("invalid material fields")
					}
					for _, mf := range doc.Fields {
						if mf.H03FieldSpec == f {
							matches = append(matches, H03InputBinding{a.ID, f.Name, "material", m.ID, f.Name})
						}
					}
				}
			} else {
				for _, dep := range a.DependsOn {
					for _, of := range b.NodeContracts[dep].OutputFields {
						if of == f {
							matches = append(matches, H03InputBinding{a.ID, f.Name, "step", dep, f.Name})
						}
					}
				}
			}
			if len(matches) != 1 {
				return fmt.Errorf("field %s requires one unambiguous identity source", f.Name)
			}
			b.InputBindings = append(b.InputBindings, matches[0])
		}
	}
	// Deep copy deployment-owned contracts so later configuration cannot mutate
	// the proposed snapshot through shared maps/slices.
	raw, _ := json.Marshal(b)
	if len(raw) > 262144 {
		return fmt.Errorf("execution closure size limit")
	}
	var frozen H03ExecutionBinding
	if err = json.Unmarshal(raw, &frozen); err != nil {
		return err
	}
	p.ExecutionBinding = &frozen
	if errs := h03BindingErrors(p); len(errs) > 0 {
		return fmt.Errorf("unsupported field closure: %v", errs)
	}
	return nil
}

func h03BindingErrors(p *Plan) []string {
	b := p.ExecutionBinding
	if b == nil {
		return nil
	}
	bad := func(why string) []string { return []string{why} }
	if b.SchemaVersion != "hub-node-binding/v1" || b.BusTenant == "" || b.Profile != h03ProfileName || len(b.Materials) == 0 || len(b.Materials) > 8 || len(b.NodeContracts) != len(p.Agents) {
		return bad("invalid finite execution binding")
	}
	pkg, err := parseScenario(b.ScenarioSnapshot)
	if err != nil || pkg.Package != p.Scenario || p.Selection == nil {
		return bad("invalid adopted scenario snapshot")
	}
	if len(p.Agents) < 2 || len(p.Steps) != len(p.Agents) || len(p.Layers) != len(p.Agents) {
		return bad("only finite unconditional chains are supported")
	}
	roleIDs := map[string]bool{}
	for _, mapping := range b.InputBindings {
		if p.agent(mapping.ToNode) == nil {
			return bad("mapping targets undeclared node")
		}
	}
	for i, layer := range p.Layers {
		if len(layer) != 1 {
			return bad("parallel/fork nodes are unsupported")
		}
		node := layer[0]
		a := p.agent(node)
		if a == nil || a.RunIf != nil {
			return bad("conditional or missing member")
		}
		var step *Step
		for k := range p.Steps {
			if p.Steps[k].ID == node {
				step = &p.Steps[k]
				break
			}
		}
		if step == nil || step.Member != a.ID || step.When != nil || !slices.Equal(step.DependsOn, a.DependsOn) {
			return bad("compiled steps differ from adopted identity chain")
		}
		if i == 0 && len(a.DependsOn) != 0 || i > 0 && (len(a.DependsOn) != 1 || a.DependsOn[0] != p.Layers[i-1][0]) {
			return bad("nonchain dependency")
		}
		c := pkg.card(a.Expert)
		if c == nil || c.BusAgent == "" || (c.Kind != "investigate" && c.Kind != "write") || c.Effects.Level != "none" || len(c.Tools) != 0 {
			return bad("unsupported profile member")
		}
		if roleIDs[c.BusAgent] {
			return bad("engineering roles require distinct logical Runtime identities")
		}
		roleIDs[c.BusAgent] = true
		contract, ok := b.NodeContracts[node]
		if !ok || contract.SchemaVersion != h03ContractSchema {
			return bad("missing node contract")
		}
		in, iok := h03SpecMap(contract.InputFields)
		out, ook := h03SpecMap(contract.OutputFields)
		if !iok || !ook || c.Input == nil || c.Output == nil || !h03ExactNames(c.Input.Requires, in) || !h03ExactNames(c.Output.Provides, out) {
			return bad("card/field contract sets differ")
		}
		for _, fields := range [][]H03FieldSpec{contract.InputFields, contract.OutputFields} {
			for _, f := range fields {
				if f.Type != "decimal" {
					return bad("only decimal fields are supported")
				}
			}
		}
		seen := map[string]bool{}
		for _, binding := range b.InputBindings {
			if binding.ToNode != node {
				continue
			}
			f, exists := in[binding.ToField]
			if !exists || seen[binding.ToField] || binding.ToField != binding.SourceField {
				return bad("unsupported or duplicate input mapping")
			}
			seen[binding.ToField] = true
			switch binding.SourceKind {
			case "material":
				found := false
				for _, m := range b.Materials {
					found = found || m.ID == binding.SourceID && h03Hex.MatchString(m.SHA256)
				}
				if !found || i != 0 {
					return bad("invalid material source")
				}
			case "step":
				if !contains(a.DependsOn, binding.SourceID) {
					return bad("undeclared upstream source")
				}
				found := false
				for _, up := range b.NodeContracts[binding.SourceID].OutputFields {
					found = found || up == f
				}
				if !found {
					return bad("incompatible upstream field")
				}
			default:
				return bad("unsupported mapping source")
			}
		}
		if len(seen) != len(in) {
			return bad("input field mapping incomplete")
		}
	}
	return nil
}

func (s *Service) h03Ready(p *Plan, adopting bool) error {
	if p.ExecutionBinding == nil {
		return nil
	}
	if errs := h03BindingErrors(p); len(errs) > 0 {
		return fmt.Errorf("unqualified profile: %v", errs)
	}
	if s.cfg.H03Profile == nil || len(s.cfg.H03RuntimeCredential) < 32 {
		return fmt.Errorf("profile runtime identity is disabled")
	}
	if p.ExecutionBinding.BusTenant != s.cfg.BusTenant {
		return fmt.Errorf("adopted Bus tenant changed; original task query/dispatch refused")
	}
	profile := s.cfg.H03Profile
	pkg := p.scenario()
	for _, a := range p.Agents {
		c := pkg.card(a.Expert)
		current, ok := profile.CandidateContracts[c.ID+"@"+c.Version]
		if !ok || !reflect.DeepEqual(current, p.ExecutionBinding.NodeContracts[a.ID]) {
			return fmt.Errorf("current role/field permission no longer covers this snapshot")
		}
	}
	for _, m := range p.ExecutionBinding.Materials {
		allowed := false
		for _, current := range profile.Materials {
			allowed = allowed || current.ID == m.ID && current.SHA256 == m.SHA256
		}
		if !allowed {
			return fmt.Errorf("current material permission missing")
		}
		if _, err := s.h03Material(m.SHA256); err != nil {
			return err
		}
	}
	if adopting {
		if pkg.Digest() != profile.Package.Digest {
			return fmt.Errorf("profile changed before adoption")
		}
		rec := s.st.Selections[p.Selection.ID]
		if rec == nil || rec.Digest != p.Selection.Digest || rec.digest() != rec.Digest {
			return fmt.Errorf("adoption selection changed")
		}
		for _, a := range p.Agents {
			el := rec.eligibility(a.Expert)
			if el == nil || el.Result != "pass" || el.Checks["h03.profile"] != profile.Version+"@"+h03ProfileDigest(profile) {
				return fmt.Errorf("selection profile version drift")
			}
		}
	}
	return nil
}

func h03FromEnv() (*H03Profile, string, error) {
	flag := os.Getenv("H03_PROFILE_ENABLED")
	if flag == "" {
		return nil, "", nil
	}
	if flag != "1" {
		return nil, "", fmt.Errorf("invalid profile flag")
	}
	credential := os.Getenv("H03_RUNTIME_CREDENTIAL")
	if len(credential) < 32 || strings.ContainsAny(credential, "\r\n") {
		return nil, "", fmt.Errorf("independent runtime identity required")
	}
	p, err := loadH03Profile(os.Getenv("H03_PROFILE_CONFIG"), os.Getenv("H03_PROFILE_SHA256"))
	return p, credential, err
}
