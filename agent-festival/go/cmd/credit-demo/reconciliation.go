package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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
//
// A run an earlier startup check blocked (an older build also blocked
// finished runs and replaced their final output) gets its result shown again
// only from original evidence: a pre-reconciliation snapshot in which it had
// finished, with nothing else about the run changed since. Without that
// evidence it stays blocked and is marked for review.
//
// Every disposition is also kept in disposition-<digest>.json beside the
// pre-reconciliation snapshot reconciliation-<digest>.json.
func (s *Service) reconcileAdoptions() error {
	var records []reconcileRecord
	backups := s.reconciliationBackups()
	for _, r := range s.st.Runs {
		issue := s.checkAdoptionSnapshot(r)
		from := r.Status
		record := func(b *BlockReason) {
			records = append(records, reconcileRecord{Run: r.ID, From: from, To: r.Status, Cause: b.Cause, Note: b.Note, Source: b.Source, Disposition: b.Disposition})
		}
		ledger := "runs/" + r.ID + "/adopt.jsonl"
		if r.Status == "blocked" && blockReasonOf(r).Code == "RECONCILIATION_REQUIRED" {
			orig, source, why := originalResult(r, backups)
			switch {
			case orig != nil:
				r.Status, r.Final, r.Blocked, r.Reconciliation = orig.Status, orig.Final, nil, nil
				b := &BlockReason{Code: "RECONCILIATION_REQUIRED", Note: "此前启动对账阻断了该任务；已按对账前快照恢复原终态与最终输出", At: now(), Source: source, Disposition: "RESTORED"}
				if issue != nil {
					b.Cause, b.Note = issue.cause, issue.note
					r.Reconciliation = b
				}
				record(b)
			case r.Blocked == nil || r.Blocked.Disposition == "":
				b := &BlockReason{Code: "RECONCILIATION_REQUIRED", Note: "此前启动对账阻断了该任务，" + why, At: now(), Source: source, Disposition: "PENDING_REVIEW"}
				if why == "" { // the snapshot shows it could still execute then
					b.Note, b.Disposition = "此前启动对账阻断了该任务；对账前它仍可执行，保持阻断", "BLOCKED"
				}
				if issue != nil {
					b.Cause = issue.cause
				}
				r.Blocked = b
				record(b)
			}
			continue
		}
		stopped := r.Status == "done" || r.Status == "failed" || r.Status == "blocked"
		switch {
		case issue == nil:
			if stopped && r.Reconciliation != nil {
				r.Reconciliation = nil // the records pass now; a blocked run is never unblocked here
				record(&BlockReason{Note: "启动对账通过，清除标记", Source: ledger, Disposition: "CLEARED"})
			}
		case stopped:
			if r.Reconciliation == nil || r.Reconciliation.Cause != issue.cause || r.Reconciliation.Note != issue.note {
				r.Reconciliation = &BlockReason{Code: "RECONCILIATION_REQUIRED", Cause: issue.cause, Note: issue.note, At: now(), Source: ledger, Disposition: "FLAGGED"}
				record(r.Reconciliation)
			}
		default:
			reason := "RECONCILIATION_REQUIRED: " + issue.note + "；保留原账本与快照，需人工核对，不能重新生成决策。"
			if r.Status != "blocked" || r.Final != reason || r.Token != "" {
				s.blockRun(r, BlockReason{Code: "RECONCILIATION_REQUIRED", Cause: issue.cause, Note: issue.note, Source: ledger, Disposition: "BLOCKED"}, reason)
				record(r.Blocked)
			}
		}
	}
	if len(records) > 0 {
		// Keep the pre-reconciliation snapshot for operator comparison. The
		// content-derived name makes repeated restarts retain the same evidence.
		digest := strings.TrimPrefix(sha(string(s.lastSaved)), "sha256:")
		if err := s.writeState(filepath.Join(s.cfg.DataDir, "reconciliation-"+digest+".json"), json.RawMessage(s.lastSaved)); err != nil {
			return fmt.Errorf("save reconciliation evidence: %w", err)
		}
		sort.Slice(records, func(i, j int) bool { return records[i].Run < records[j].Run })
		if err := s.writeState(filepath.Join(s.cfg.DataDir, "disposition-"+digest+".json"), map[string]any{
			"at": now(), "pre_state": "reconciliation-" + digest + ".json", "records": records}); err != nil {
			return fmt.Errorf("save reconciliation dispositions: %w", err)
		}
		return s.save()
	}
	return nil
}

// reconcileRecord is one startup disposition: status before and after, the
// failed check, the record it rests on, and what was done.
type reconcileRecord struct {
	Run         string `json:"run"`
	From        string `json:"from"`
	To          string `json:"to"`
	Cause       string `json:"cause,omitempty"`
	Note        string `json:"note"`
	Source      string `json:"source"`
	Disposition string `json:"disposition"`
}

type stateBackup struct {
	name string
	st   State
}

// reconciliationBackups reads the pre-reconciliation snapshots kept in the
// data directory. Unreadable ones are skipped: they cannot be evidence.
func (s *Service) reconciliationBackups() []stateBackup {
	paths, _ := filepath.Glob(filepath.Join(s.cfg.DataDir, "reconciliation-*.json"))
	sort.Strings(paths)
	var out []stateBackup
	for _, path := range paths {
		raw, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		var st State
		if json.Unmarshal(raw, &st) == nil {
			out = append(out, stateBackup{filepath.Base(path), st})
		}
	}
	return out
}

// originalResult finds the run as it was before a startup check blocked it.
// Only a finished original whose record matches the current one apart from
// what blocking changes counts; otherwise why says what is missing. A source
// without an original or a reason means the run could still execute then.
func originalResult(r *Run, backups []stateBackup) (orig *Run, source, why string) {
	if len(backups) == 0 {
		return nil, "", "未找到对账前快照，原结果待核对"
	}
	// The finished original is cited by its own snapshot, never by a snapshot
	// in which the run was still running; among several, the smallest name, so
	// the scan order does not change the answer.
	var found *Run
	var finishedSource, runningSource string
	for _, b := range backups {
		o := b.st.Runs[r.ID]
		if o == nil || o.Status == "blocked" {
			continue
		}
		if o.Status != "done" && o.Status != "failed" {
			if runningSource == "" || b.name < runningSource {
				runningSource = b.name
			}
			continue
		}
		if !sameApartFromBlock(r, o) {
			return nil, b.name, "对账前快照之后任务记录有其他变化，原结果待核对"
		}
		if found != nil && (found.Status != o.Status || found.Final != o.Final) {
			return nil, b.name, "多份对账前快照记录的原结果不一致，原结果待核对"
		}
		if found == nil || b.name < finishedSource {
			found, finishedSource = o, b.name
		}
	}
	if found != nil {
		return found, finishedSource, ""
	}
	if runningSource != "" {
		return nil, runningSource, ""
	}
	return nil, "", "对账前快照中没有该任务的原结果，原结果待核对"
}

// sameApartFromBlock compares two records of one run, ignoring the fields
// blocking a run changes.
func sameApartFromBlock(cur, orig *Run) bool {
	c, o := *cur, *orig
	c.Status, c.Final, c.Token, c.TokenSaved, c.Blocked, c.Reconciliation = "", "", "", "", nil, nil
	o.Status, o.Final, o.Token, o.TokenSaved, o.Blocked, o.Reconciliation = "", "", "", "", nil, nil
	a, err1 := json.Marshal(c)
	b, err2 := json.Marshal(o)
	return err1 == nil && err2 == nil && bytes.Equal(a, b)
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
