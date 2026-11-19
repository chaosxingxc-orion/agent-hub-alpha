package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
	"github.com/openjiuwen/agent-festival/go/internal/compose"
	"github.com/openjiuwen/agent-festival/go/internal/evidlog"
)

// ALK-05 / AR-06 (#27): capability selection generates the graph.

// Fixed PR33 ee1bac5 fixture, retained here without importing legacy deployment.
// Original bytes SHA256: 97def791b1fbdfc936e0bce79de4083a5db9953674230d7d1209cf47c536f4d4
const annualReportPackage = `{
  "package": "annual-report-cmb-2025",
  "version": "0.1.0",
  "title": "公开年报经营与资产质量简报（#16 / W01）",
  "governance": {
    "state": "approved",
    "approved_by": "演示审批人（受治理导入，非生产身份）",
    "approved_at": "2026-11-17"
  },
  "rules": {
    "rule_set": "annual-report-selection-v1",
    "coverage": [
      {"name": "年报快照", "selectors": ["data:annual-report/cmb-2025/*"]}
    ]
  },
  "caller_holds": [
    {"kind": "data", "selector": "data:annual-report/cmb-2025/*", "actions": ["read"], "effect": "none"},
    {"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none"}
  ],
  "parent_bounds": {
    "max_cost": {"amount": 300, "unit": "USD-cents"},
    "max_downstream_calls": 20,
    "max_seconds": 1200
  },
  "materials": [
    {"id": "cmb-2025-annual-report-a", "sha256": "52c49dcdc00d4f5436adabc6561044d7a539e9bb0a8ebbdb692a1da6237459da", "source": "PR #21 source.json"}
  ],
  "selection": {
    "recall": "在获准场景包内按能力卡召回全部候选（有限配置，未做语义召回）",
    "source_schema": "material-ref",
    "must_declare": {
      "material-ref": ["entities", "basis", "periods", "materials"],
      "indicator-set": ["basis", "periods"],
      "findings": ["templates"]
    },
    "weights": {"focus": 0.6, "io": 0.3, "current": 0.1},
    "min_score": 0.5
  },
  "capabilities": [
    {
      "id": "ar.material-extractor", "version": "1.0.0", "node": "extract",
      "name": "年报指标抽取", "kind": "investigate",
      "summary": "从固定年报快照抽取指标及原文引用",
      "instruction": "你是年报资料员。只依据给定的固定年报快照抽取指标，每个数值都给出单位、期间、PDF 页、印刷页和原文行；快照里找不到的指标写明缺失，不要编造。正文印刷页 = PDF 页 - 1。",
      "requires": [{"kind": "data", "selector": "data:annual-report/cmb-2025/*", "actions": ["read"], "effect": "none", "purpose": "fixed annual report snapshot"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 40, "calls": 2, "seconds": 300},
      "applicability": {
        "entities": ["招商银行股份有限公司"], "basis": ["本集团"], "periods": ["2025", "2024"],
        "units": ["人民币百万元", "%"], "materials": ["cmb-2025-annual-report-a"]
      },
      "input": {"schema": "material-ref"},
      "output": {"schema": "indicator-set", "provides": ["operating_income", "net_profit_attributable", "net_interest_margin", "cost_income_ratio", "npl_balance", "npl_ratio", "provision_coverage", "cet1_advanced"]},
      "limits": ["只读快照，不联网", "页码按 PR #21 的规则：正文印刷页 = PDF 页 - 1"]
    },
    {
      "id": "ar.operating-analyst", "version": "1.0.0", "node": "operating",
      "name": "经营分析", "kind": "investigate",
      "summary": "分析营收、归母净利润、净息差和成本收入比的变化",
      "instruction": "你是经营分析员。只用上游抽取的指标和引用作分析；比率变化用百分点表述；每条结论都标明所依据的指标和原文页码。",
      "requires": [{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none", "purpose": "upstream indicators of this task"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 30, "calls": 0, "seconds": 180},
      "focus": ["经营"],
      "applicability": {"basis": ["本集团"], "periods": ["2025", "2024"]},
      "input": {"schema": "indicator-set", "requires": ["operating_income", "net_profit_attributable", "net_interest_margin", "cost_income_ratio"]},
      "output": {"schema": "findings", "section": "operating"},
      "limits": ["比率变化用百分点表述"]
    },
    {
      "id": "ar.asset-quality-analyst", "version": "1.0.0", "node": "asset_quality",
      "name": "资产质量分析", "kind": "investigate",
      "summary": "分析不良余额与不良率、拨备覆盖率和核心一级资本充足率",
      "instruction": "你是资产质量分析员。只用上游抽取的指标和引用作分析；核心一级资本充足率取高级法，与权重法区分；每条结论都标明所依据的指标和原文页码。",
      "requires": [{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none", "purpose": "upstream indicators of this task"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 30, "calls": 0, "seconds": 180},
      "focus": ["资产质量"],
      "applicability": {"basis": ["本集团"], "periods": ["2025", "2024"]},
      "input": {"schema": "indicator-set", "requires": ["npl_balance", "npl_ratio", "provision_coverage", "cet1_advanced"]},
      "output": {"schema": "findings", "section": "asset_quality"},
      "limits": ["核心一级资本充足率取高级法，与权重法区分"]
    },
    {
      "id": "ar.report-writer", "version": "1.0.0", "node": "write",
      "name": "年报简报写作", "kind": "write",
      "summary": "按 W01 模板把分析段落写成结构化简报",
      "instruction": "你是简报撰写人。按交付模板组织上游分析段落；每个数字和结论都要回到快照的页与原文行，没有出处的内容不写。",
      "requires": [{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none", "purpose": "upstream findings of this task"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 30, "calls": 0, "seconds": 180},
      "applicability": {"templates": ["demo/annual-report/report.schema.json"]},
      "input": {"schema": "findings", "requires_sections_any": ["operating", "asset_quality"]},
      "output": {"schema": "report", "ref": "demo/annual-report/report.schema.json"},
      "limits": ["每个数字和结论都要回到快照的页与原文行"]
    },
    {
      "id": "ar.parent-only-analyst", "version": "2.3.0", "node": "parent_analysis",
      "name": "母公司口径财务分析", "kind": "investigate",
      "summary": "母公司口径的经营与资产质量分析（演示：应被排除的候选）",
      "instruction": "你是母公司口径财务分析员。只分析母公司报表口径的指标。",
      "requires": [{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none", "purpose": "upstream indicators of this task"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 30, "calls": 0, "seconds": 180},
      "focus": ["经营", "资产质量"],
      "applicability": {"basis": ["母公司"], "periods": ["2025", "2024"]},
      "input": {"schema": "indicator-set", "requires": ["operating_income", "net_profit_attributable"]},
      "output": {"schema": "findings", "section": "operating"},
      "limits": ["只支持母公司口径"]
    },
    {
      "id": "ar.generic-finance-qa", "version": "0.9.0", "node": "generic_qa",
      "name": "通用财报问答", "kind": "investigate",
      "summary": "通用财报问答（演示：口径、期间、输入输出均未声明的候选）",
      "instruction": "你是通用财报问答助手。",
      "requires": [{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": ["read"], "effect": "none", "purpose": "upstream indicators of this task"}],
      "effects": {"level": "none"},
      "bounds": {"cost_cents": 20, "calls": 0, "seconds": 120}
    }
  ]
}
`

const arPackage = "annual-report-cmb-2025"

// withAnnualReport imports the annual-report package (optionally edited) the
// way an operator does, through a scenario directory.
func withAnnualReport(t *testing.T, edit func(doc map[string]any)) {
	t.Helper()
	body := annualReportPackage
	if edit != nil {
		var doc map[string]any
		if err := json.Unmarshal([]byte(body), &doc); err != nil {
			t.Fatal(err)
		}
		edit(doc)
		raw, err := json.Marshal(doc)
		if err != nil {
			t.Fatal(err)
		}
		body = string(raw)
	}
	if err := withScenarioDir(t, map[string]string{arPackage + ".json": body}); err != nil {
		t.Fatal(err)
	}
}

func capabilityDoc(doc map[string]any, id string) map[string]any {
	for _, c := range doc["capabilities"].([]any) {
		if m := c.(map[string]any); m["id"] == id {
			return m
		}
	}
	panic("no capability " + id)
}

// selectionService is a service over the loaded packages with a fixture
// orchestrator that compiles a team into steps; compiles counts its calls.
func selectionService(t *testing.T) (*Service, *atomic.Int64) {
	t.Helper()
	compose.RegisterGraph()
	compiles := &atomic.Int64{}
	orch := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		compiles.Add(1)
		var in struct {
			Agents []struct {
				ID        string   `json:"id"`
				DependsOn []string `json:"dependsOn"`
				Condition *struct {
					Type string `json:"type"`
					Node string `json:"node"`
					Text string `json:"text"`
				} `json:"condition"`
			} `json:"agents"`
		}
		if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
			w.WriteHeader(400)
			return
		}
		out := compileResult{Blueprint: "fixture", SpecDigest: "sha256:fixture", Type: "graph"}
		for _, a := range in.Agents {
			st := Step{ID: a.ID, Member: a.ID, DependsOn: a.DependsOn}
			if c := a.Condition; c != nil {
				st.When = &RunIf{Step: c.Node, Contains: c.Text, Negate: c.Type == "output_not_contains"}
			}
			out.Steps = append(out.Steps, st)
		}
		_ = json.NewEncoder(w).Encode(out)
	}))
	t.Cleanup(orch.Close)
	dir := t.TempDir()
	cat, err := LoadCatalog(dir)
	if err != nil {
		t.Fatal(err)
	}
	return &Service{cfg: Config{DataDir: dir, OrchestratorURL: orch.URL, LeaseSeconds: 30}, cat: cat,
		machines: map[string]*adopt.Machine{}, logs: map[string]*evidlog.Log{},
		st: &State{Plans: map[string][]*Plan{}, Runs: map[string]*Run{}}, writeState: writeJSON}, compiles
}

func arQuery(focus ...string) SelectionQuery {
	return SelectionQuery{Package: arPackage, Entity: "招商银行股份有限公司", Basis: "本集团", Period: "2025", ComparePeriod: "2024",
		Focus: focus, Deliverable: "demo/annual-report/report.schema.json", Materials: []string{"cmb-2025-annual-report-a"}}
}

func mustSelect(t *testing.T, s *Service, q SelectionQuery) *SelectionRecord {
	t.Helper()
	rec, err := s.Select(q)
	if err != nil {
		t.Fatal(err)
	}
	return rec
}

func mapping(rec *SelectionRecord) map[string]string {
	out := map[string]string{}
	for _, g := range rec.Graph {
		out[g.Node] = g.Capability + " <- " + strings.Join(g.DependsOn, ",")
	}
	return out
}

func planCount(s *Service) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, vs := range s.st.Plans {
		n += len(vs)
	}
	return n + len(s.st.Runs)
}

// Two task focuses over the same approved package give different, explained
// selections and different graphs.
func TestSelectionDiffersByTaskFocus(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)

	both := mustSelect(t, s, arQuery("经营", "资产质量"))
	if both.Outcome != "graph" {
		t.Fatalf("no graph: %s", both.Reason)
	}
	if got := strings.Join(both.Selection, ","); got != "ar.material-extractor,ar.operating-analyst,ar.asset-quality-analyst,ar.report-writer" {
		t.Fatalf("selection for 经营+资产质量: %s", got)
	}
	want := map[string]string{
		"extract":       "ar.material-extractor@1.0.0 <- ",
		"operating":     "ar.operating-analyst@1.0.0 <- extract",
		"asset_quality": "ar.asset-quality-analyst@1.0.0 <- extract",
		"write":         "ar.report-writer@1.0.0 <- operating,asset_quality",
	}
	if got := mapping(both); len(got) != len(want) {
		t.Fatalf("graph mapping: %v", got)
	} else {
		for k, v := range want {
			if got[k] != v {
				t.Fatalf("node %s: %q, want %q", k, got[k], v)
			}
		}
	}

	aq := mustSelect(t, s, arQuery("资产质量"))
	if got := strings.Join(aq.Selection, ","); got != "ar.material-extractor,ar.asset-quality-analyst,ar.report-writer" {
		t.Fatalf("selection for 资产质量: %s", got)
	}
	if why := aq.whyNot("ar.operating-analyst"); !strings.Contains(why, "低于选择阈值") || !strings.Contains(why, "不覆盖本任务重点") {
		t.Fatalf("operating analyst left out without its basis: %q", why)
	}
	if got := mapping(aq)["write"]; got != "ar.report-writer@1.0.0 <- asset_quality" {
		t.Fatalf("writer wiring for 资产质量: %q", got)
	}
	if both.ID == aq.ID || both.Digest == aq.Digest {
		t.Fatal("two tasks share one selection record")
	}
	for _, rec := range []*SelectionRecord{both, aq} {
		if rec.Digest != rec.digest() || rec.Package.Digest != scenarios.get(arPackage).Digest() || rec.ScoreSource == "" {
			t.Fatalf("record %s is not pinned to its package and contents", rec.ID)
		}
	}
}

// A candidate ranked at the top but outside the task's basis is excluded: it
// never enters the graph, the planner cannot put it back, and the run never
// dispatches it.
func TestHighScoreInapplicableCandidateExcludedAndNeverDispatched(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	rec := mustSelect(t, s, arQuery("经营", "资产质量"))

	parent := rec.eligibility("ar.parent-only-analyst")
	if parent.Result != "excluded" || !strings.Contains(strings.Join(parent.Reasons, ""), "口径") {
		t.Fatalf("parent-only candidate: %+v", parent)
	}
	for _, r := range rec.Ranking {
		if r.Score > parent.Score {
			t.Fatalf("fixture: %s outranks the parent-only candidate (%v > %v)", r.Capability, r.Score, parent.Score)
		}
	}
	if rec.selected("ar.parent-only-analyst") || !contains(rec.Recall.Candidates, "ar.parent-only-analyst") {
		t.Fatal("parent-only candidate must be recalled and not selected")
	}

	// the planner tries to put it back in place of the operating analyst
	before := planCount(s)
	forced := append([]PlanAgent(nil), rec.Team...)
	for i := range forced {
		if forced[i].ID == "operating" {
			forced[i].Expert = "ar.parent-only-analyst"
		}
	}
	if _, err := s.Propose(ProposeInput{SelectionID: rec.ID, Agents: forced}); err == nil || !strings.Contains(err.Error(), "was not selected") || !strings.Contains(err.Error(), "口径") {
		t.Fatalf("excluded candidate accepted into a plan: %v", err)
	}
	if planCount(s) != before {
		t.Fatal("a rejected proposal left a plan or run behind")
	}

	// the selected team is adopted, approved and run end to end
	p, err := s.Propose(ProposeInput{SelectionID: rec.ID})
	if err != nil {
		t.Fatal(err)
	}
	if !p.Precheck.OK || p.Scenario != arPackage || p.Selection == nil || p.Selection.Digest != rec.Digest {
		t.Fatalf("plan from selection: precheck=%+v scenario=%s selection=%+v", p.Precheck, p.Scenario, p.Selection)
	}
	g, err := s.graph(p, false)
	if err != nil {
		t.Fatal(err)
	}
	if g.Selection == nil || g.Selection.ID != rec.ID {
		t.Fatal("the graph the user approves does not show the selection")
	}
	shown := map[string]string{}
	for _, c := range g.Selection.Candidates {
		shown[c.Capability] = c.Result + "：" + c.Reason
	}
	if !strings.HasPrefix(shown["ar.parent-only-analyst"], "excluded：") || !strings.HasPrefix(shown["ar.generic-finance-qa"], "unknown：") {
		t.Fatalf("user projection hides exclusions: %v", shown)
	}
	for _, n := range g.Nodes {
		if n.Chosen == "" {
			t.Fatalf("node %s shows no selection basis", n.ID)
		}
	}

	r, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil)
	if err != nil {
		t.Fatal(err)
	}
	job, err := s.Lease("fixture-executor")
	if err != nil || job == nil || job.RunID != r.ID {
		t.Fatalf("lease: %+v %v", job, err)
	}
	if !strings.Contains(job.UserInput, "本集团") || !strings.Contains(job.UserInput, "cmb-2025-annual-report-a") {
		t.Fatalf("executor input lacks the task scope: %s", job.UserInput)
	}
	dispatched := map[string]int{}
	for _, id := range r.Order {
		reused, _, err := s.StepGet(r.ID, job.Token, id)
		if err != nil || reused {
			t.Fatalf("step %s: reused=%v %v", id, reused, err)
		}
		dispatched[r.Steps[id].Expert]++
		if err := s.StepPut(r.ID, job.Token, id, "[step:"+id+"] fixture output"); err != nil {
			t.Fatal(err)
		}
	}
	if err := s.Done(r.ID, job.Token, nil, "fixture report", ""); err != nil {
		t.Fatal(err)
	}
	if r.Status != "done" || len(dispatched) != 4 {
		t.Fatalf("run %s: %v", r.Status, dispatched)
	}
	for _, id := range []string{"ar.parent-only-analyst", "ar.generic-finance-qa"} {
		if dispatched[id] != 0 {
			t.Fatalf("%s dispatched %d times", id, dispatched[id])
		}
		for _, m := range job.Members {
			if m["expert"] == id {
				t.Fatalf("%s handed to the executor", id)
			}
		}
	}
}

// When every candidate is unknown nothing passes: no graph is generated and
// nothing can be proposed from the record.
func TestAllUnknownCandidatesYieldNoGraph(t *testing.T) {
	withAnnualReport(t, func(doc map[string]any) {
		for _, c := range doc["capabilities"].([]any) {
			m := c.(map[string]any)
			delete(m, "applicability")
			delete(m, "input")
			delete(m, "output")
		}
	})
	s, compiles := selectionService(t)
	rec := mustSelect(t, s, arQuery("经营", "资产质量"))
	for _, el := range rec.Eligibility {
		if el.Result != "unknown" {
			t.Fatalf("%s: %s, want unknown", el.Capability, el.Result)
		}
	}
	if rec.Outcome != "no_graph" || len(rec.Graph) != 0 || len(rec.Team) != 0 || !strings.Contains(rec.Reason, "未知不算通过") {
		t.Fatalf("all-unknown selection: outcome=%s graph=%v reason=%s", rec.Outcome, rec.Graph, rec.Reason)
	}
	if _, err := s.Propose(ProposeInput{SelectionID: rec.ID}); err == nil || !strings.Contains(err.Error(), "no graph") {
		t.Fatalf("proposed from an all-unknown selection: %v", err)
	}
	if planCount(s) != 0 || compiles.Load() != 0 {
		t.Fatalf("plans/runs %d, compiles %d", planCount(s), compiles.Load())
	}
}

// Data that does not bind is refused before adoption: the selection yields
// no graph, nothing is compiled, proposed or dispatched; a planner who drops
// a binding from a good selection is refused the same way.
func TestIncompatibleCompositionRejectedBeforeAdoption(t *testing.T) {
	withAnnualReport(t, func(doc map[string]any) {
		out := capabilityDoc(doc, "ar.material-extractor")["output"].(map[string]any)
		var kept []any
		for _, p := range out["provides"].([]any) {
			if p != "cet1_advanced" {
				kept = append(kept, p)
			}
		}
		out["provides"] = kept
	})
	s, compiles := selectionService(t)
	rec := mustSelect(t, s, arQuery("资产质量"))
	if rec.eligibility("ar.asset-quality-analyst").Result != "pass" {
		t.Fatal("fixture: asset quality analyst should be eligible; the break is in the data binding")
	}
	if rec.Outcome != "no_graph" || rec.Composition.Result != "incompatible" || !strings.Contains(rec.Reason, "cet1_advanced") {
		t.Fatalf("incompatible composition: outcome=%s composition=%+v", rec.Outcome, rec.Composition)
	}
	if _, err := s.Propose(ProposeInput{SelectionID: rec.ID}); err == nil {
		t.Fatal("proposed from an incompatible selection")
	}

	good := mustSelect(t, s, arQuery("经营"))
	if good.Outcome != "graph" {
		t.Fatalf("经营 alone should still compose: %s", good.Reason)
	}
	cut := append([]PlanAgent(nil), good.Team...)
	for i := range cut {
		if cut[i].ID == "operating" {
			cut[i].DependsOn = nil // reads indicators from nobody
		}
	}
	if _, err := s.Propose(ProposeInput{SelectionID: good.ID, Agents: cut}); err == nil || !strings.Contains(err.Error(), "does not compose") {
		t.Fatalf("team without its data binding accepted: %v", err)
	}
	if planCount(s) != 0 || compiles.Load() != 0 {
		t.Fatalf("plans/runs %d, compiles %d: rejection must come before adoption", planCount(s), compiles.Load())
	}
}

// A newly approved capability is recalled and selected by changing only the
// scenario package: no control-state or executor change.
func TestNewApprovedCapabilityOnlyChangesPackage(t *testing.T) {
	q := arQuery("流动性")
	t.Run("before", func(t *testing.T) {
		withAnnualReport(t, nil)
		s, _ := selectionService(t)
		if rec := mustSelect(t, s, q); rec.Outcome != "no_graph" {
			t.Fatalf("no capability covers 流动性 yet, got %v", rec.Selection)
		}
	})
	t.Run("after", func(t *testing.T) {
		withAnnualReport(t, func(doc map[string]any) {
			ext := capabilityDoc(doc, "ar.material-extractor")["output"].(map[string]any)
			ext["provides"] = append(ext["provides"].([]any), "lcr", "nsfr")
			w := capabilityDoc(doc, "ar.report-writer")["input"].(map[string]any)
			w["requires_sections_any"] = append(w["requires_sections_any"].([]any), "liquidity")
			doc["version"] = "0.2.0"
			doc["capabilities"] = append(doc["capabilities"].([]any), map[string]any{
				"id": "ar.liquidity-analyst", "version": "1.0.0", "node": "liquidity", "name": "流动性分析", "kind": "investigate",
				"summary": "分析流动性覆盖率和净稳定资金比例", "instruction": "你是流动性分析员。只用上游抽取的指标和引用作分析。",
				"requires": []any{map[string]any{"kind": "data", "selector": "data:annual-report/evidence/*", "actions": []any{"read"}, "effect": "none"}},
				"effects":  map[string]any{"level": "none"}, "bounds": map[string]any{"cost_cents": 30, "calls": 0, "seconds": 180},
				"focus":         []any{"流动性"},
				"applicability": map[string]any{"basis": []any{"本集团"}, "periods": []any{"2025", "2024"}},
				"input":         map[string]any{"schema": "indicator-set", "requires": []any{"lcr", "nsfr"}},
				"output":        map[string]any{"schema": "findings", "section": "liquidity"},
			})
		})
		s, _ := selectionService(t)
		rec := mustSelect(t, s, q)
		if rec.Outcome != "graph" || !rec.selected("ar.liquidity-analyst") {
			t.Fatalf("new capability not selected: %s %v", rec.Reason, rec.Selection)
		}
		if got := mapping(rec)["write"]; got != "ar.report-writer@1.0.0 <- liquidity" {
			t.Fatalf("writer wiring: %q", got)
		}
		p, err := s.Propose(ProposeInput{SelectionID: rec.ID})
		if err != nil || !p.Precheck.OK {
			t.Fatalf("plan with the new capability: %v %+v", err, p)
		}
	})
}

// The approval binds the selection record and the package: changing either
// after the plan is shown blocks approval.
func TestSelectionAndPackageBoundIntoApproval(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	rec := mustSelect(t, s, arQuery("经营", "资产质量"))
	p, err := s.Propose(ProposeInput{SelectionID: rec.ID})
	if err != nil {
		t.Fatal(err)
	}

	other := *p
	other.Selection = &SelectionRef{ID: rec.ID, Digest: "sha256:other"}
	if planHash(&other, s.cat.Hashes) == p.Hash {
		t.Fatal("plan hash does not cover the selection digest")
	}

	s.mu.Lock()
	rec.NotSelected = append(rec.NotSelected, NotSelected{"ar.operating-analyst", "edited after the plan was shown"})
	s.mu.Unlock()
	if _, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil); err == nil || !strings.Contains(err.Error(), "selection") {
		t.Fatalf("approved over an edited selection record: %v", err)
	}
	s.mu.Lock()
	rec.NotSelected = rec.NotSelected[:len(rec.NotSelected)-1]
	s.mu.Unlock()

	raw := strings.Replace(annualReportPackage, `"min_score": 0.5`, `"min_score": 0.45`, 1)
	changed, err := parseScenario([]byte(raw))
	if err != nil {
		t.Fatal(err)
	}
	original := scenarios.packages[arPackage]
	scenarios.packages[arPackage] = changed
	_, err = s.Approve(p.ID, p.Version, p.Hash, "alice", nil)
	scenarios.packages[arPackage] = original
	if err == nil || !strings.Contains(err.Error(), "changed") {
		t.Fatalf("approved under a different package: %v", err)
	}
	if _, err := s.Approve(p.ID, p.Version, p.Hash, "alice", nil); err != nil {
		t.Fatalf("unchanged selection and package: %v", err)
	}
}

// Credit plans keep working by hand and cannot borrow another package's
// capabilities; packages that publish a selection policy require one.
func TestCreditPlansUnchangedAndPackagesDoNotMix(t *testing.T) {
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	if _, err := s.Select(SelectionQuery{Package: defaultScenario}); err == nil || !strings.Contains(err.Error(), "no selection policy") {
		t.Fatalf("credit-dd selection: %v", err)
	}
	var agents []PlanAgent
	for _, st := range goodSteps() {
		agents = append(agents, PlanAgent{ID: st.ID, Expert: st.Member, DependsOn: st.DependsOn, RunIf: st.When})
	}
	p, err := s.Propose(ProposeInput{CaseID: "C006", Goal: "授信尽调", Agents: agents})
	if err != nil || !p.Precheck.OK || p.Scenario != defaultScenario || p.Selection != nil {
		t.Fatalf("credit plan: %v %+v", err, p)
	}
	if _, err := s.Propose(ProposeInput{PlanID: p.ID, SelectionID: mustSelect(t, s, arQuery("经营")).ID}); err == nil || !strings.Contains(err.Error(), "belongs to scenario") {
		t.Fatalf("a credit plan was revised into an annual-report plan: %v", err)
	}
	agents[1].Expert = "ar.operating-analyst"
	if _, err := s.Propose(ProposeInput{CaseID: "C006", Goal: "授信尽调", Agents: agents}); err == nil || !strings.Contains(err.Error(), "belongs to scenario") {
		t.Fatalf("credit plan borrowed an annual-report capability: %v", err)
	}
	q := arQuery("经营")
	q.Basis = ""
	if _, err := s.Select(q); err == nil || !strings.Contains(err.Error(), "basis") {
		t.Fatalf("selection without the task's basis: %v", err)
	}
}

// TestALK05Samples writes the selection records and the plan projection for
// the two task focuses when ALK05_SAMPLES_DIR is set (evidence for #27).
func TestALK05Samples(t *testing.T) {
	dir := os.Getenv("ALK05_SAMPLES_DIR")
	if dir == "" {
		t.Skip("ALK05_SAMPLES_DIR not set")
	}
	withAnnualReport(t, nil)
	s, _ := selectionService(t)
	write := func(name string, v any) {
		raw, err := json.MarshalIndent(v, "", "  ")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, name), append(raw, '\n'), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	both := mustSelect(t, s, arQuery("经营", "资产质量"))
	aq := mustSelect(t, s, arQuery("资产质量"))
	write("selection-S1-operating-and-asset-quality.json", both)
	write("selection-S2-asset-quality.json", aq)
	p, err := s.Propose(ProposeInput{SelectionID: both.ID})
	if err != nil {
		t.Fatal(err)
	}
	g, err := s.graph(p, false)
	if err != nil {
		t.Fatal(err)
	}
	write("plan-graph-from-S1.json", g)
}
