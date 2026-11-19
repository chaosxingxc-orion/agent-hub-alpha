package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/openjiuwen/agent-festival/go/internal/rfc8785"
)

const h03ProfileSchema = "hub-node-profile/v1"
const h03ContractSchema = "hub-node-contract/v1"
const h03ProfileName = "two-role-identity/v1"

var h03ID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,127}$`)
var h03Hex = regexp.MustCompile(`^[0-9a-f]{64}$`)
var h03Decimal = regexp.MustCompile(`^-?(0|[1-9][0-9]{0,9})(\.[0-9]{1,6})?$`)

// h03RawHash deliberately does NOT call sha/canonicalHash: those original
// framework digests append NUL. This is the new private document byte pin.
func h03RawHash(value any) string {
	raw, err := rfc8785.Canonicalize(jsonDoc(value))
	if err != nil {
		return ""
	}
	sum := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(sum[:])
}

func h03BytesHash(raw []byte) string {
	sum := sha256.Sum256(raw)
	return hex.EncodeToString(sum[:])
}

func h03Strict(raw []byte, target any, limit int) error {
	if len(raw) == 0 || len(raw) > limit {
		return fmt.Errorf("bounded H03 JSON required")
	}
	// Reuse syntax/duplicate-key rejection only, never Builder authority.
	if _, err := builderJSON(raw); err != nil {
		return fmt.Errorf("invalid H03 JSON syntax")
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if err := d.Decode(target); err != nil {
		return err
	}
	if err := d.Decode(new(any)); err != io.EOF {
		return fmt.Errorf("trailing H03 document")
	}
	return nil
}

func h03Read(path, pin string, limit int) ([]byte, error) {
	if !h03Hex.MatchString(pin) {
		return nil, fmt.Errorf("pinned material required")
	}
	i, err := os.Lstat(path)
	if err != nil || !i.Mode().IsRegular() || i.Size() > int64(limit) {
		return nil, fmt.Errorf("regular bounded H03 source required")
	}
	raw, err := os.ReadFile(path)
	if err != nil || len(raw) > limit || h03BytesHash(raw) != pin {
		return nil, fmt.Errorf("H03 source digest mismatch")
	}
	return raw, nil
}

func loadH03Profile(path, pin string) (*H03Profile, error) {
	raw, err := h03Read(path, pin, 262144)
	if err != nil {
		return nil, err
	}
	var p H03Profile
	if err = h03Strict(raw, &p, 262144); err != nil {
		return nil, err
	}
	if p.SchemaVersion != h03ProfileSchema || !h03ID.MatchString(p.Version) || len(p.Materials) == 0 || len(p.Materials) > 8 || len(p.CandidateContracts) == 0 || len(p.CandidateContracts) > 64 {
		return nil, fmt.Errorf("invalid server H03 profile")
	}
	seen := map[string]bool{}
	for i := range p.Materials {
		m := &p.Materials[i]
		if !h03ID.MatchString(m.ID) || seen[m.ID] {
			return nil, fmt.Errorf("invalid material identity")
		}
		seen[m.ID] = true
		m.raw, err = h03Read(m.Path, m.SHA256, 65536)
		if err != nil {
			return nil, err
		}
	}
	return &p, nil
}

func h03ProfileDigest(p *H03Profile) string {
	// Machine-local paths are not business semantics and never reach the UI.
	materials := make([]MaterialRef, 0, len(p.Materials))
	for _, m := range p.Materials {
		materials = append(materials, MaterialRef{ID: m.ID, SHA256: m.SHA256})
	}
	return h03RawHash(map[string]any{"schema_version": p.SchemaVersion, "version": p.Version, "package": p.Package, "field_semantics": p.FieldSemantics, "candidate_contracts": p.CandidateContracts, "materials": materials})
}

func h03SpecMap(fields []H03FieldSpec) (map[string]H03FieldSpec, bool) {
	if len(fields) == 0 || len(fields) > 32 {
		return nil, false
	}
	m := map[string]H03FieldSpec{}
	for _, f := range fields {
		if !h03ID.MatchString(f.Name) || f.Type == "" || f.Unit == "" || f.Period == "" || f.Basis == "" {
			return nil, false
		}
		if _, ok := m[f.Name]; ok {
			return nil, false
		}
		m[f.Name] = f
	}
	return m, true
}

func h03ExactNames(names []string, fields map[string]H03FieldSpec) bool {
	if len(names) != len(fields) {
		return false
	}
	seen := map[string]bool{}
	for _, n := range names {
		if _, ok := fields[n]; !ok || seen[n] {
			return false
		}
		seen[n] = true
	}
	return true
}

func h03DecimalValid(value string) bool {
	if len(value) > 24 || !h03Decimal.MatchString(value) || value == "-0" || strings.Contains(value, ".") && strings.HasSuffix(value, "0") {
		return false
	}
	v := strings.TrimPrefix(value, "-")
	parts := strings.SplitN(v, ".", 2)
	whole, err := strconv.ParseInt(parts[0], 10, 64)
	return err == nil && whole <= 1000000000 && (whole < 1000000000 || len(parts) == 1)
}

// Only the explicit deployment profile activates additional checks. Unknown
// facts never pass; original recall, score, ranking and graph code stay intact.
func h03Eligibility(p *H03Profile, pkg *ScenarioPackage, card *CapabilityCard, q *SelectionQuery, el Eligibility) Eligibility {
	unknown := func(key, why string) {
		el.Checks[key] = "unknown"
		el.Reasons = append(el.Reasons, why)
		if el.Result == "pass" {
			el.Result = "unknown"
		}
	}
	exclude := func(key, why string) {
		el.Checks[key] = "excluded"
		el.Reasons = append(el.Reasons, why)
		el.Result = "excluded"
	}
	el.Checks["h03.profile"] = p.Version + "@" + h03ProfileDigest(p)
	if p.SchemaVersion != h03ProfileSchema || p.Package.ID != pkg.Package || p.Package.Version != pkg.Version || p.Package.Digest != pkg.Digest() {
		unknown("h03.profile", "server profile/package identity is unknown or stale")
		return el
	}
	semantics, ok := h03SpecMap(p.FieldSemantics)
	if !ok {
		unknown("h03.fields", "required field semantics are incomplete")
		return el
	}
	for name, required := range semantics {
		if required.Period != q.Period && (q.ComparePeriod == "" || required.Period != q.ComparePeriod) {
			exclude("h03.query-period:"+name, "profile field period is outside the requested periods")
		}
		if required.Basis != q.Basis {
			exclude("h03.query-basis:"+name, "profile field basis differs from the requested basis")
		}
	}
	contract, ok := p.CandidateContracts[card.ID+"@"+card.Version]
	if !ok || contract.SchemaVersion != h03ContractSchema {
		unknown("h03.contract", "approved capability-version field contract missing")
		return el
	}
	in, iok := h03SpecMap(contract.InputFields)
	out, ook := h03SpecMap(contract.OutputFields)
	if !iok || !ook || card.Input == nil || card.Output == nil {
		unknown("h03.contract", "field contract is incomplete")
		return el
	}
	if !h03ExactNames(card.Input.Requires, in) || !h03ExactNames(card.Output.Provides, out) {
		exclude("h03.fields", "card I/O and field contract sets differ")
		return el
	}
	for _, fields := range []map[string]H03FieldSpec{in, out} {
		for name, actual := range fields {
			want, known := semantics[name]
			if !known {
				unknown("h03.field:"+name, "required semantics unknown for "+name)
				continue
			}
			for _, d := range []struct{ key, actual, want string }{{"type", actual.Type, want.Type}, {"unit", actual.Unit, want.Unit}, {"period", actual.Period, want.Period}, {"basis", actual.Basis, want.Basis}} {
				key := "h03." + d.key + ":" + name
				if d.actual != d.want || d.key == "type" && d.actual != "decimal" {
					exclude(key, fmt.Sprintf("%s: expected=%s declared=%s", key, d.want, d.actual))
				} else {
					el.Checks[key] = "compatible"
				}
			}
		}
	}
	if card.Input.Schema == pkg.Selection.SourceSchema {
		available := map[string]H03FieldSpec{}
		for _, id := range q.Materials {
			var material *H03ProfileMaterial
			for i := range p.Materials {
				if p.Materials[i].ID == id {
					material = &p.Materials[i]
					break
				}
			}
			if material == nil || !h03Hex.MatchString(material.SHA256) || h03BytesHash(material.raw) != material.SHA256 {
				unknown("h03.material:"+id, "fixed material bytes are missing or changed")
				continue
			}
			pinned := false
			for _, m := range pkg.Materials {
				pinned = pinned || m.ID == id && m.SHA256 == material.SHA256
			}
			if !pinned {
				exclude("h03.material:"+id, "profile material does not match scenario pin")
				continue
			}
			var doc H03MaterialDocument
			if h03Strict(material.raw, &doc, 65536) != nil || doc.SchemaVersion != "h03-engineering-material/v1" || len(doc.Fields) == 0 || len(doc.Fields) > 32 {
				unknown("h03.material:"+id, "material field document is unknown")
				continue
			}
			for _, f := range doc.Fields {
				if _, duplicate := available[f.Name]; duplicate {
					exclude("h03.material:"+f.Name, "ambiguous material field")
					continue
				}
				if !h03DecimalValid(f.Value) {
					exclude("h03.material:"+f.Name, "noncanonical material decimal")
					continue
				}
				available[f.Name] = f.H03FieldSpec
			}
		}
		for name, want := range in {
			actual, found := available[name]
			if !found {
				unknown("h03.material:"+name, "required material field is absent")
				continue
			}
			if actual != want {
				exclude("h03.material:"+name, "material type/unit/period/basis does not match contract")
			}
		}
	}
	return el
}
