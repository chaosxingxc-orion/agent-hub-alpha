package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/openjiuwen/agent-festival/go/internal/replay"
)

type definition struct {
	scenario scenario
	run      func(*scenario, *engine) error
}

func selectCases(selection string) ([]definition, error) {
	all := []definition{
		{scenario{ID: "happy", Title: "材料齐全，正常授信五万美元", Kind: "正常业务", Purpose: "资产三千五百万美元、近六个月无逾期；观察从模拟判断到真实控制组件再到模拟银行的完整链路。"}, happy},
		{scenario{ID: "retry", Title: "同一申请重发，返回原回执", Kind: "正常重试", Purpose: "模拟客户端重复提交同一个操作；验证原结果复用，模型和银行提交次数均不增加。结果读取由本次 Demo 宿主提供。"}, retryRequest},
		{scenario{ID: "conflict", Title: "沿用编号却改了客户，拒绝办理", Kind: "冲突输入", Purpose: "保留操作编号，却把客户换成另一人；验证项目按内容摘要拒绝冲突，原办理结果保持不变。"}, conflictingRequest},
		{scenario{ID: "revoked", Title: "判断后权限被撤销，提交被拦截", Kind: "正常权限变更", Purpose: "登记后、派发前撤销执行权限；当前权限由模拟授权提供方传入，实际拦截由项目状态机执行。"}, revoked},
		{scenario{ID: "lost-reply", Title: "银行已办理但响应丢失，查询后恢复", Kind: "明确的故障注入", Purpose: "模拟银行先持久保存业务效果，再返回响应丢失错误；宿主沿原业务编号查询对账，不重新提交。未模拟真实网络链路。"}, lostReply},
		{scenario{ID: "restart", Title: "办理后进程被终止，换进程恢复", Kind: "真实子进程终止", Purpose: "子进程在银行回执已保存、宿主完成记录未保存的确定位置被父进程终止；重新打开文件后查询银行，验证不产生第二笔效果。"}, restart},
		{scenario{ID: "evidence", Title: "缺资产材料，先补证再办理", Kind: "正常补证", Purpose: "先返回待补证；模拟 S2 获取固定资料，项目取证日志检查实际记录，再以新证据修订提交。补证路由是 Demo 宿主逻辑。"}, missingEvidence},
		{scenario{ID: "uncovered", Title: "近期短期逾期，规则未覆盖", Kind: "正常业务边界", Purpose: "高资产客户最近一个月内有一次持续一个月的逾期。模拟判断应报告规则未覆盖，由宿主暂停，不能补成批准或拒绝。"}, uncovered},
		{scenario{ID: "history-veto", Title: "历史长期逾期，停止授信办理", Kind: "正常业务拒绝", Purpose: "两年前存在持续四个月的逾期，触发历史优先否决。此业务判断由固定规则替身产生，不是框架自动理解了授信制度。"}, historyVeto},
		{scenario{ID: "revision-gap", Title: "新修订要求替代旧判断，现有组件仍放行旧操作", Kind: "已知系统集成缺口", Purpose: "场景文件明确要求 er-2 替代 er-1，但项目原型缺少接收这份有效修订政策的控制入口。直接观察旧操作仍获许可，不替框架补上该保护。", knownGap: true}, revisionGap},
		{scenario{ID: "replay", Title: "隔离重演控制过程，不新增银行效果", Kind: "历史控制回放", Purpose: "保留真实控制输入和事件，在项目隔离回放器上重新执行；分别检查事件一致与账本规则，模拟银行提交次数不能增加。"}, replayScenario},
	}
	all = append(all, definition{scenario{ID: "grant-regroup", Title: "权限未减少，仅重组角色条目却被误拒绝", Kind: "正常授权表达变化 · 新发现", Purpose: "同一工具的修改与执行权限，从一条授权拆为两条，资源、动作和效果上限均未减少。先验证流程准入接受，再比较实际派发；不撤销权限、不破坏数据。"}, grantRegroup})
	if selection == "all" {
		return all, nil
	}
	wanted := map[string]bool{}
	for _, id := range strings.Split(selection, ",") {
		id = strings.TrimSpace(id)
		if id == "" {
			return nil, fmt.Errorf("empty scenario ID")
		}
		wanted[id] = true
	}
	var chosen []definition
	for _, d := range all {
		if wanted[d.scenario.ID] {
			chosen = append(chosen, d)
			delete(wanted, d.scenario.ID)
		}
	}
	if len(wanted) > 0 {
		return nil, fmt.Errorf("unknown scenario IDs: %v", wanted)
	}
	return chosen, nil
}

func metrics(s *scenario, e *engine, expectedModel, expectedEffects, expectedSubmits int) error {
	c, err := e.counts()
	if err != nil {
		return err
	}
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("S1 替身调用次数", expectedModel, c.S1Calls)
	s.expect("模拟银行产生的授信笔数", expectedEffects, b.Effects)
	s.expect("模拟银行提交次数", expectedSubmits, b.SubmitCalls)
	return nil
}

func happy(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	r, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("办理结果", "completed", r.Status)
	s.expect("实际授信金额（USD）", 50_000, r.Receipt.Payload.CreditUSD)
	s.expect("项目操作状态", "completed", e.machine.State(a.OperationID))
	return metrics(s, e, 1, 1, 1)
}

func retryRequest(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	first, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	second, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("重复请求处理", "reused", second.Status)
	s.expect("仍为原回执", first.Receipt.Reference, second.Receipt.Reference)
	return metrics(s, e, 1, 1, 1)
}

func conflictingRequest(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	first, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	a.Customer = "different-customer"
	second, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("相同编号不同实质内容", "conflict", second.Status)
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("原银行回执保持不变", first.Receipt.Reference, b.Receipts[a.BusinessID].Reference)
	return metrics(s, e, 1, 1, 1)
}

func revoked(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	r, err := e.submit(a, true, false)
	if err != nil {
		return err
	}
	s.expect("执行权限撤销后", "blocked", r.Status)
	s.expect("项目状态未推进至派发", "adopted", e.machine.State(a.OperationID))
	events, err := e.machine.Events()
	if err != nil {
		return err
	}
	refusals := 0
	for _, event := range events {
		if event.EventType == "control" && event.Outcome.Status == "rejected" {
			refusals++
		}
	}
	s.expect("持久拒绝记录数", 1, refusals)
	return metrics(s, e, 1, 0, 0)
}

func lostReply(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	r, err := e.submit(a, false, true)
	if err != nil {
		return err
	}
	s.expect("响应丢失后的状态", "unresolved", r.Status)
	again, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("未确认时重试不重新办理", "in-progress", again.Status)
	if _, err := e.reconcile(a); err != nil {
		return err
	}
	s.expect("查询对账后的项目状态", "completed", e.machine.State(a.OperationID))
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("银行查询次数", 1, b.QueryCalls)
	return metrics(s, e, 1, 1, 1)
}

const crashReady = "CREDIT_DEMO_BANK_COMMITTED"

func crashWorker(dir, rule string) error {
	e, err := openEngine(dir, rule)
	if err != nil {
		return err
	}
	e.afterBank = func() error {
		// The marker is emitted only after the bank snapshot is closed and
		// the framework dispatch is recorded, but before host completion.
		fmt.Println(crashReady)
		time.Sleep(30 * time.Second)
		return fmt.Errorf("crash worker was not terminated at the barrier")
	}
	_, err = e.submit(baseApplication("restart"), false, false)
	return err
}

func restart(s *scenario, parent *engine) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, exe, "-worker-dir", parent.dir)
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	ready := make(chan error, 1)
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			if scanner.Text() == crashReady {
				ready <- nil
				return
			}
		}
		ready <- fmt.Errorf("worker did not reach bank-committed barrier: %v", scanner.Err())
	}()
	select {
	case err := <-ready:
		if err != nil {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
			return fmt.Errorf("%w; stderr=%s", err, stderr.String())
		}
	case <-ctx.Done():
		_ = cmd.Process.Kill()
		_ = cmd.Wait()
		return fmt.Errorf("worker barrier timeout: %s", stderr.String())
	}
	if err := cmd.Process.Kill(); err != nil {
		_ = cmd.Wait()
		return err
	}
	if err := cmd.Wait(); err == nil {
		return fmt.Errorf("worker unexpectedly exited successfully")
	}
	s.expect("子进程被实际终止", false, cmd.ProcessState.Success())
	// The parent never opened a machine for this case. Only after the child
	// dies does a new owner reconstruct the durable state.
	e, err := openEngine(parent.dir, parent.rule)
	if err != nil {
		return err
	}
	a := baseApplication("restart")
	s.expect("重新打开后保留派发状态", "dispatched", e.machine.State(a.OperationID))
	if err := e.log("demo-host", "实际终止子进程后重新打开状态", "银行回执文件已存在；宿主尚无完成结果"); err != nil {
		return err
	}
	o := e.machine.Dispatch(a.OperationID, grants())
	s.expect("恢复后重复派发被组件拒绝", "rejected", o.Status)
	if _, err := e.reconcile(a); err != nil {
		return err
	}
	s.expect("对账后完成", "completed", e.machine.State(a.OperationID))
	return metrics(s, e, 1, 1, 1)
}

func missingEvidence(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	a.AssetReport = false
	before, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("缺材料时", "needs-evidence", before.Status)
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("补证前没有银行效果", 0, b.Effects)
	a, coverage, err := e.supplement(a)
	if err != nil {
		return err
	}
	a.Revision = "er-2"
	s.expect("项目取证覆盖检查", true, coverage)
	after, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("补证后的办理结果", "completed", after.Status)
	c, err := e.counts()
	if err != nil {
		return err
	}
	s.expect("S2 资料提供方替身调用", 1, c.S2Calls)
	return metrics(s, e, 2, 1, 1)
}

func uncovered(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	a.Overdue = []overdue{{1, 1}}
	r, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("业务结果", "uncovered", r.Status)
	return metrics(s, e, 1, 0, 0)
}

func historyVeto(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	a.Overdue = []overdue{{4, 24}}
	r, err := e.submit(a, false, false)
	if err != nil {
		return err
	}
	s.expect("业务结果", "deny", r.Status)
	return metrics(s, e, 1, 0, 0)
}

func revisionGap(s *scenario, e *engine) error {
	old := baseApplication(s.ID)
	newer := old
	newer.OperationID, newer.Revision, newer.EvidenceID = "op-new-revision", "er-2", "evidence-corrected-v2"
	newer.Overdue = []overdue{{4, 24}}
	// This policy is a scenario requirement, not an authority consumed by
	// the reference machine. The absent integration is exactly the gap.
	if err := writeJSON(filepath.Join(e.dir, "correction-policy.json"), map[string]any{
		"business_request_id": old.BusinessID, "expected_effective_revision": newer.Revision,
		"supersedes": old.Revision, "requirement": "新拒绝修订正式替代旧批准，旧操作不得派发",
		"boundary": "场景要求；现有组件无有效修订提供方入口，未由 Demo 代补保护",
	}); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(e.dir, "new-decision.json"), judge(newer)); err != nil {
		return err
	}
	for _, a := range []application{old, newer} {
		req, err := e.request(a)
		if err != nil {
			return err
		}
		if o := e.machine.Adopt(req); o.Status != "adopted" {
			return fmt.Errorf("revision scenario setup: %s", o.Status)
		}
	}
	o := e.machine.DispatchIfCurrent(old.OperationID, grants(), -1)
	s.expect("旧修订派发应被拒绝", "rejected", o.Status)
	if err := e.log("project-core", "检查旧操作能否派发", o.Status+"；当前组件没有消费场景声明的修订替代政策"); err != nil {
		return err
	}
	if o.Status == "dispatched" {
		if _, err := e.bank.submit(bankPayload{old.BusinessID, old.Customer, 50_000}, false); err != nil {
			return err
		}
	}
	b, err := e.bank.state()
	if err != nil {
		return err
	}
	s.expect("旧批准不应产生银行效果", 0, b.Effects)
	return nil
}

func replayScenario(s *scenario, e *engine) error {
	a := baseApplication(s.ID)
	if _, err := e.submit(a, false, false); err != nil {
		return err
	}
	req, err := e.request(a)
	if err != nil {
		return err
	}
	events, err := e.machine.Events()
	if err != nil {
		return err
	}
	rawReq, err := json.Marshal(req)
	if err != nil {
		return err
	}
	caseValue := map[string]any{
		"case_version": "0.1.0", "case_id": "local-credit-control-replay",
		"steps":           []replay.Step{{Op: "adopt", Request: rawReq}, {Op: "dispatch", OpID: a.OperationID, Grants: grants()}, {Op: "mark_completed", OpID: a.OperationID}},
		"recorded_events": events, "expected": map[string]any{"fidelity": "match", "conformance": "pass"},
	}
	if err := writeJSON(filepath.Join(e.dir, "replay-case.json"), caseValue); err != nil {
		return err
	}
	encoded, err := json.Marshal(caseValue)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.UseNumber()
	var doc map[string]any
	if err := decoder.Decode(&doc); err != nil {
		return err
	}
	c, err := replay.LoadCase(doc)
	if err != nil {
		return err
	}
	r, err := replay.Replay(c)
	if err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(e.dir, "replay-report.json"), r); err != nil {
		return err
	}
	s.expect("事件重演一致", "match", r.Fidelity)
	s.expect("原记录符合账本规则", "pass", r.Conformance)
	if err := e.log("project-core", "在临时隔离状态机上回放", fmt.Sprintf("fidelity=%s conformance=%s；银行调用不在回放路径中", r.Fidelity, r.Conformance)); err != nil {
		return err
	}
	return metrics(s, e, 1, 1, 1)
}
