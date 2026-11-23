package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func h03TestFixture(t *testing.T) (*ScenarioPackage, SelectionQuery, *H03Profile) {
	t.Helper()
	field := func(n string) H03FieldSpec {
		return H03FieldSpec{Name: n, Type: "decimal", Unit: "USD-million", Period: "FY2025", Basis: "consolidated"}
	}
	doc := H03MaterialDocument{SchemaVersion: "h03-engineering-material/v1", Fields: []H03Field{{field("component_a"), "40"}, {field("component_b"), "60"}, {field("reference"), "95"}}}
	raw, _ := json.Marshal(doc)
	pkg := &ScenarioPackage{Package: "engineering", Version: "0.1.0", Materials: []MaterialRef{{ID: "material-a", SHA256: h03BytesHash(raw)}}, Selection: &SelectionPolicy{SourceSchema: "material-ref", MinScore: 0.5, MustDeclare: map[string][]string{"material-ref": {"entities", "basis", "periods", "materials"}, "indicator-set": {"basis", "periods", "templates"}}}}
	pkg.Selection.Weights.Focus = 0.3
	pkg.Selection.Weights.IO = 0.6
	pkg.Selection.Weights.Current = 0.1
	app := &Applicability{Entities: []string{"engineering-entity"}, Basis: []string{"consolidated"}, Periods: []string{"FY2025"}, Materials: []string{"material-a"}, Templates: []string{"engineering-report"}}
	sum := CapabilityCard{ID: "eng.sum", Version: "0.1.0", Node: "sum", Name: "sum", Kind: "investigate", Focus: []string{"other"}, Applicability: app, Input: &CapabilityIO{Schema: "material-ref", Requires: []string{"component_a", "component_b", "reference"}}, Output: &CapabilityIO{Schema: "indicator-set", Provides: []string{"reference", "total"}}}
	bad := sum
	bad.ID = "eng.wrong-unit"
	bad.Node = "wrong"
	bad.Focus = []string{"reconcile"}
	reconcile := CapabilityCard{ID: "eng.reconcile", Version: "0.1.0", Node: "reconcile", Name: "reconcile", Kind: "write", Applicability: app, Input: &CapabilityIO{Schema: "indicator-set", Requires: []string{"reference", "total"}}, Output: &CapabilityIO{Schema: "engineering-report", Ref: "engineering-report", Provides: []string{"difference"}}}
	pkg.Capabilities = []CapabilityCard{bad, sum, reconcile}
	pkg.digest = h03RawHash(pkg)
	contract := H03NodeContract{SchemaVersion: h03ContractSchema, InputFields: []H03FieldSpec{field("component_a"), field("component_b"), field("reference")}, OutputFields: []H03FieldSpec{field("reference"), field("total")}}
	wrong := contract
	wrong.OutputFields = append([]H03FieldSpec{}, contract.OutputFields...)
	wrong.OutputFields[1].Unit = "EUR-million"
	p := &H03Profile{SchemaVersion: h03ProfileSchema, Version: "1.0.0", Package: PackageRef{ID: pkg.Package, Version: pkg.Version, Digest: pkg.Digest()}, FieldSemantics: []H03FieldSpec{field("component_a"), field("component_b"), field("reference"), field("total"), field("difference")}, CandidateContracts: map[string]H03NodeContract{"eng.sum@0.1.0": contract, "eng.wrong-unit@0.1.0": wrong, "eng.reconcile@0.1.0": {SchemaVersion: h03ContractSchema, InputFields: []H03FieldSpec{field("reference"), field("total")}, OutputFields: []H03FieldSpec{field("difference")}}}, Materials: []H03ProfileMaterial{{ID: "material-a", SHA256: h03BytesHash(raw), raw: raw}}}
	q := SelectionQuery{Package: pkg.Package, Entity: "engineering-entity", Basis: "consolidated", Period: "FY2025", Focus: []string{"reconcile"}, Deliverable: "engineering-report", Materials: []string{"material-a"}}
	return pkg, q, p
}

func TestH03EligibilityLowerScoreCompatibleSelected(t *testing.T) {
	pkg, q, p := h03TestFixture(t)
	rec, err := selectCapabilities(pkg, q, p)
	if err != nil {
		t.Fatal(err)
	}
	bad, good := rec.eligibility("eng.wrong-unit"), rec.eligibility("eng.sum")
	if len(rec.Recall.Candidates) != 3 || len(rec.Eligibility) != 3 || bad.Result != "excluded" || bad.Score != 1 || good.Result != "pass" || good.Score != 0.7 || rec.Outcome != "graph" || !rec.selected("eng.sum") || rec.selected("eng.wrong-unit") || len(rec.Graph) != 2 {
		t.Fatalf("wrong selection: %+v", rec)
	}
	if bad.Checks["h03.unit:total"] != "excluded" || !strings.Contains(strings.Join(bad.Reasons, " "), "USD-million") {
		t.Fatalf("missing precise reason: %+v", bad)
	}
	if !strings.Contains(good.Checks["h03.profile"], h03ProfileDigest(p)) {
		t.Fatal("profile digest missing")
	}
	before := rec.digest()
	p.Version = "1.0.1"
	after, _ := selectCapabilities(pkg, q, p)
	after.CreatedAt = rec.CreatedAt
	if after.digest() == before {
		t.Fatal("selection digest must bind actual profile version")
	}
}

func TestH03EligibilityUnknownDoesNotPass(t *testing.T) {
	for _, mode := range []string{"missing-contract", "missing-semantics", "missing-material", "bad-profile", "material-pin", "material-field"} {
		t.Run(mode, func(t *testing.T) {
			pkg, q, p := h03TestFixture(t)
			switch mode {
			case "missing-contract":
				p.CandidateContracts = nil
			case "missing-semantics":
				p.FieldSemantics = nil
			case "missing-material":
				p.Materials = nil
			case "bad-profile":
				p.Package.Digest = "sha256:" + strings.Repeat("0", 64)
			case "material-pin":
				p.Materials[0].raw = []byte("{}")
			case "material-field":
				var d H03MaterialDocument
				_ = json.Unmarshal(p.Materials[0].raw, &d)
				d.Fields = d.Fields[:2]
				raw, _ := json.Marshal(d)
				p.Materials[0].raw = raw
				p.Materials[0].SHA256 = h03BytesHash(raw)
				pkg.Materials[0].SHA256 = h03BytesHash(raw)
			}
			r, err := selectCapabilities(pkg, q, p)
			if err != nil {
				t.Fatal(err)
			}
			if r.Outcome != "no_graph" || r.selected("eng.sum") {
				t.Fatalf("unknown must not generate qualified graph: %+v", r)
			}
			if r.eligibility("eng.sum").Result != "unknown" {
				t.Fatalf("unknown reason lost: %+v", r.eligibility("eng.sum"))
			}
		})
	}
}

func TestH03EligibilityDimensionsAndCardSets(t *testing.T) {
	for _, dimension := range []string{"type", "unit", "period", "basis", "input-set", "output-set", "material-unit", "decimal"} {
		t.Run(dimension, func(t *testing.T) {
			pkg, q, p := h03TestFixture(t)
			c := p.CandidateContracts["eng.sum@0.1.0"]
			switch dimension {
			case "type":
				c.OutputFields[1].Type = "string"
			case "unit":
				c.OutputFields[1].Unit = "EUR-million"
			case "period":
				c.OutputFields[1].Period = "FY2024"
			case "basis":
				c.OutputFields[1].Basis = "parent"
			case "input-set":
				c.InputFields = c.InputFields[:2]
			case "output-set":
				c.OutputFields = c.OutputFields[:1]
			case "material-unit", "decimal":
				var d H03MaterialDocument
				_ = json.Unmarshal(p.Materials[0].raw, &d)
				if dimension == "decimal" {
					d.Fields[0].Value = "40.0"
				} else {
					d.Fields[0].Unit = "EUR-million"
				}
				raw, _ := json.Marshal(d)
				p.Materials[0].raw = raw
				p.Materials[0].SHA256 = h03BytesHash(raw)
				pkg.Materials[0].SHA256 = h03BytesHash(raw)
			}
			p.CandidateContracts["eng.sum@0.1.0"] = c
			r, err := selectCapabilities(pkg, q, p)
			if err != nil {
				t.Fatal(err)
			}
			if r.Outcome != "no_graph" || r.eligibility("eng.sum").Result != "excluded" {
				t.Fatalf("incompatible candidate passed: %+v", r)
			}
		})
	}
}

func TestH03EligibilityLegacyNilProfileUnchanged(t *testing.T) {
	pkg, q, _ := h03TestFixture(t)
	a, e := selectCapabilities(pkg, q)
	if e != nil {
		t.Fatal(e)
	}
	b, e := selectCapabilities(pkg, q, nil)
	if e != nil {
		t.Fatal(e)
	}
	a.CreatedAt = "fixed"
	b.CreatedAt = "fixed"
	if a.digest() != b.digest() || !a.selected("eng.wrong-unit") {
		t.Fatal("legacy selection algorithm or digest changed")
	}
}

func TestH03EligibilityQueryCannotChangeFixedFieldScope(t *testing.T) {
	for _, dim := range []string{"period", "basis"} {
		t.Run(dim, func(t *testing.T) {
			pkg, q, p := h03TestFixture(t)
			// Even when cards declare the additional task scope, fixed material
			// fields cannot silently be interpreted as belonging to that scope.
			if dim == "period" {
				q.Period = "FY2024"
				for i := range pkg.Capabilities {
					pkg.Capabilities[i].Applicability.Periods = []string{"FY2024", "FY2025"}
				}
			} else {
				q.Basis = "parent"
				for i := range pkg.Capabilities {
					pkg.Capabilities[i].Applicability.Basis = []string{"parent", "consolidated"}
				}
			}
			r, err := selectCapabilities(pkg, q, p)
			if err != nil || r.Outcome != "no_graph" || r.eligibility("eng.sum").Result != "excluded" {
				t.Fatalf("scope drift accepted: %+v %v", r, err)
			}
		})
	}
}

func TestH03EligibilityPinnedServerProfile(t *testing.T) {
	_, _, p := h03TestFixture(t)
	dir := t.TempDir()
	p.Materials[0].Path = filepath.Join(dir, "material.json")
	if e := os.WriteFile(p.Materials[0].Path, p.Materials[0].raw, 0600); e != nil {
		t.Fatal(e)
	}
	raw, _ := json.Marshal(p)
	path := filepath.Join(dir, "profile.json")
	if e := os.WriteFile(path, raw, 0600); e != nil {
		t.Fatal(e)
	}
	loaded, e := loadH03Profile(path, h03BytesHash(raw))
	if e != nil {
		t.Fatal(e)
	}
	if h03ProfileDigest(loaded) != h03ProfileDigest(p) {
		t.Fatal("loaded semantics differ")
	}
	if e = os.WriteFile(p.Materials[0].Path, []byte("{}"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e = loadH03Profile(path, h03BytesHash(raw)); e == nil {
		t.Fatal("drifted material accepted")
	}
	for _, v := range []string{"0", "-1", "1.25", "1000000000"} {
		if !h03DecimalValid(v) {
			t.Fatalf("valid decimal rejected %s", v)
		}
	}
	for _, v := range []string{"-0", "01", "1.0", "1e3", "1000000000.1", "NaN"} {
		if h03DecimalValid(v) {
			t.Fatalf("invalid decimal accepted %s", v)
		}
	}
	var target H03MaterialDocument
	if h03Strict([]byte(`{"schema_version":"a","schema_version":"b","fields":[]}`), &target, 65536) == nil {
		t.Fatal("duplicate JSON keys accepted")
	}
}
