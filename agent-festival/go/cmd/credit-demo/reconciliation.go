package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
	"github.com/openjiuwen/agent-festival/go/internal/replay"
)

// reconcileAdoptions runs before serving requests. It uses the existing replay
// engine on scratch storage; never edits ledger evidence or guesses a lost
// model result. A mismatch withdraws this run's lease, not other runs' leases.
//
// A run that can still execute (queued, running, waiting) is blocked. A run
// with nothing left to execute (done, failed, or already blocked for another
// reason) keeps its status, final output and original block reason; the
// failed check is recorded beside it, so the finished result stays visible.
func (s *Service) reconcileAdoptions() error {
	changed := false
	for _, r := range s.st.Runs {
		issue := s.checkAdoptionSnapshot(r)
		stopped := r.Status == "done" || r.Status == "failed" ||
			(r.Status == "blocked" && (r.Blocked == nil || r.Blocked.Code != "RECONCILIATION_REQUIRED"))
		switch {
		case issue == nil:
			if stopped && r.Reconciliation != nil {
				r.Reconciliation = nil // the records pass now; a blocked run is never unblocked here
				changed = true
			}
		case stopped:
			if r.Reconciliation == nil || r.Reconciliation.Cause != issue.cause || r.Reconciliation.Note != issue.note {
				r.Reconciliation = &BlockReason{Code: "RECONCILIATION_REQUIRED", Cause: issue.cause, Note: issue.note, At: now()}
				changed = true
			}
		default:
			reason := "RECONCILIATION_REQUIRED: " + issue.note + "；保留原账本与快照，需人工核对，不能重新生成决策。"
			if r.Status != "blocked" || r.Final != reason || r.Token != "" {
				s.blockRun(r, BlockReason{Code: "RECONCILIATION_REQUIRED", Cause: issue.cause, Note: issue.note}, reason)
				changed = true
			}
		}
	}
	if changed {
		// Keep the pre-reconciliation snapshot for operator comparison. The
		// content-derived name makes repeated restarts retain the same evidence.
		backup := filepath.Join(s.cfg.DataDir, "reconciliation-"+strings.TrimPrefix(sha(string(s.lastSaved)), "sha256:")+".json")
		if err := s.writeState(backup, json.RawMessage(s.lastSaved)); err != nil {
			return fmt.Errorf("save reconciliation evidence: %w", err)
		}
		return s.save()
	}
	return nil
}

// reconcileIssue is a failed startup check: a stable cause code and the
// operator note (fixed text and identifiers only).
type reconcileIssue struct{ cause, note string }

func issue(cause, note string) *reconcileIssue { return &reconcileIssue{cause, note} }

func (s *Service) checkAdoptionSnapshot(r *Run) *reconcileIssue {
	m, err := adopt.Open(filepath.Join(s.cfg.DataDir, "runs", r.ID, "adopt.jsonl"))
	if err != nil {
		return issue("LEDGER_UNREADABLE", "采纳账本不可读取或版本/格式不兼容")
	}
	if m.TornTail() {
		return issue("LEDGER_TORN_TAIL", "采纳账本存在不完整尾记录")
	}
	events, err := m.Events()
	if err != nil {
		return issue("LEDGER_READ_FAILED", "采纳账本读取失败")
	}
	if len(events) == 0 && len(r.ReplaySteps) == 0 && r.Decision == nil && r.DecisionOp == "" && r.BusinessReq == "" {
		return nil
	}
	// Records of another adoption protocol are not judged by the current
	// replay: the current machine refuses them, which would read as a
	// sequence mismatch. Say what it is instead; nothing is migrated here.
	for _, step := range r.ReplaySteps {
		if req, ok := step["request"].(map[string]any); ok {
			if v, _ := req["adoption_version"].(string); v != adoptionVersion {
				return issue("LEGACY_ADOPTION_PROTOCOL", fmt.Sprintf("采纳记录使用旧协议 %s（当前 %s），当前回放不接受，未作一致性判定；需迁移或按旧协议人工核对", clip(v, 16), adoptionVersion))
			}
		}
	}

	// Adopt.Open validates raw documents before typed projection. Replay then
	// checks snapshot commands against the whole validated event sequence.
	raw, err := json.Marshal(map[string]any{"case_version": "0.2.0", "case_id": strings.ToLower(r.ID),
		"steps": r.ReplaySteps, "recorded_events": events,
		"expected": map[string]any{"fidelity": "match", "conformance": "pass"}})
	if err != nil {
		return issue("SNAPSHOT_UNENCODABLE", "采纳快照无法编码")
	}
	var doc map[string]any
	if err := numberDecode(raw, &doc); err != nil {
		return issue("SNAPSHOT_UNENCODABLE", "采纳快照无法编码")
	}
	c, err := replay.LoadCase(doc)
	if err != nil {
		return issue("SNAPSHOT_PROTOCOL", "采纳快照不满足当前回放协议")
	}
	result, err := replay.Replay(c)
	if err != nil || result.Fidelity != replay.Match || result.Conformance != replay.Pass {
		return issue("LEDGER_SNAPSHOT_MISMATCH", "采纳账本与快照操作序列不一致")
	}
	for _, event := range events {
		if event.Outcome.Status != "adopted" {
			continue
		}
		digest, err := adopt.Digest(r.Decision)
		if err != nil || r.Decision == nil || event.Idempotency == nil || digest != event.Idempotency.PayloadDigest ||
			r.DecisionOp != event.Identity.OperationID || r.BusinessReq != event.Identity.BusinessRequestID ||
			event.Versions.Policy != r.PlanHash || event.Versions.Model != s.runModel(r) {
			return issue("DECISION_MISMATCH", "决策内容、身份或绑定与采纳记录不一致")
		}
		// A pre-fix snapshot can contain a decision without its node result.
		// Do not ask a model to reproduce it or silently manufacture a result.
		for id, st := range r.Steps {
			if e := expertByID(st.Expert); e != nil && e.Kind == "judge" && st.Status == "done" &&
				st.Output == "[step:"+id+"]\n"+judgmentOutput(r.Decision) {
				return nil
			}
		}
		return issue("RESULT_MISSING", "已采纳决策缺少匹配的持久节点结果")
	}
	return issue("ADOPTION_MISSING", "快照缺少可核对的原始采纳事件")
}

func judgmentOutput(decision map[string]any) string {
	return fmt.Sprintf("DECISION=%s\n依据：%v\n证据步骤：%v", strings.ToUpper(fmt.Sprint(decision["level"])), decision["rationale"], decision["evidence_steps"])
}
