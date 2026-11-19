package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
	"github.com/openjiuwen/agent-festival/go/internal/compose"
	"github.com/openjiuwen/agent-festival/go/internal/ledger"
)

// A normal enterprise IAM representation change, with identical authority.
// The grouped positive control grants permission only; it never calls the bank.
func grantRegroup(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	if err := e.count(false); err != nil {
		return err
	}
	v := judge(a)
	if err := e.log("simulated-model", "正常申请判断", v.Reason); err != nil {
		return err
	}
	req, err := e.request(a)
	if err != nil {
		return err
	}
	grouped := grants()
	grouped[1].Actions = []string{"write", "execute"}
	split := []ledger.Grant{
		grouped[0],
		{Kind: grouped[1].Kind, Selector: grouped[1].Selector, Actions: []string{"write"}, Effect: grouped[1].Effect},
		{Kind: grouped[1].Kind, Selector: grouped[1].Selector, Actions: []string{"execute"}, Effect: grouped[1].Effect},
	}
	if err := writeJSON(filepath.Join(e.dir, "grants-before.json"), grouped); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(e.dir, "grants-after.json"), split); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(e.dir, "application.json"), a); err != nil {
		return err
	}
	// Independently exercise the existing graph permission admission. This
	// establishes that split authority is supported by another project component.
	raw, err := os.ReadFile(filepath.Join(compose.ExamplesDir, "valid", "doc-extractor.json"))
	if err != nil {
		return err
	}
	var contract map[string]any
	if err := json.Unmarshal(raw, &contract); err != nil {
		return err
	}
	delete(contract, "_expect")
	delete(contract, "_expect_match")
	resource := map[string]any{"kind": grouped[1].Kind, "selector": grouped[1].Selector}
	contract["requires"] = []any{map[string]any{"resource": resource, "actions": []any{"write", "execute"}, "effect": "irreversible"}}
	contract["effects"] = map[string]any{"level": "irreversible", "compensation": "manual"}
	var held []any
	for _, g := range split {
		actions := make([]any, len(g.Actions))
		for i, action := range g.Actions {
			actions[i] = action
		}
		held = append(held, map[string]any{"resource": map[string]any{"kind": g.Kind, "selector": g.Selector}, "actions": actions, "effect": g.Effect})
	}
	graph := map[string]any{"nodes": map[string]any{"credit": contract}, "edges": []any{}, "plan": map[string]any{"seq": []any{"credit"}}, "caller_holds": held}
	compose.RegisterGraph()
	admission := compose.ComposeErrors(graph)
	if err := writeJSON(filepath.Join(e.dir, "graph-input.json"), graph); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(e.dir, "graph-admission.json"), admission); err != nil {
		return err
	}
	if len(admission.Errors) != 0 {
		return fmt.Errorf("scenario prerequisite failed: split authority rejected by graph: %v", admission.Errors)
	}
	s.expect("流程准入接受拆分权限", 0, len(admission.Errors))
	req.GrantsAtDecision = grouped
	positive, err := adopt.Open(filepath.Join(e.dir, "grouped-control.jsonl"))
	if err != nil {
		return err
	}
	controlReq := *req
	controlReq.Identity.OperationID += "-control"
	if o := positive.Adopt(&controlReq); o.Status != "adopted" {
		return fmt.Errorf("grouped control setup: %s", o.Status)
	}
	control := positive.Dispatch(controlReq.Identity.OperationID, grouped)
	s.expect("合并写法的派发正例", "dispatched", control.Status)
	if o := e.machine.Adopt(req); o.Status != "adopted" {
		return fmt.Errorf("split authority setup: %s", o.Status)
	}
	actual := e.machine.Dispatch(a.OperationID, split)
	s.expect("等价拆分写法也应允许派发", "dispatched", actual.Status)
	if err := e.log("project-core", "权限分组对照", "流程准入通过；合并写法="+control.Status+"；等价拆分写法="+actual.Status+"："+actual.Detail); err != nil {
		return err
	}
	if actual.Status == "dispatched" {
		r, err := e.bank.submit(bankPayload{a.BusinessID, a.Customer, v.CreditUSD}, false)
		if err != nil {
			return err
		}
		if err := e.finish(a, req.PayloadDigest, v, r); err != nil {
			return err
		}
	}
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("合法申请应产生一笔模拟银行效果", 1, b.Effects)
	return nil
}
