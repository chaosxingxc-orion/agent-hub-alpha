package main

// This file is DEMO HOST code. It wires the existing reference components;
// its workflow, result storage and simulated providers are not SDK features.
import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/openjiuwen/agent-festival/go/internal/adopt"
	"github.com/openjiuwen/agent-festival/go/internal/evidlog"
	"github.com/openjiuwen/agent-festival/go/internal/ledger"
)

type overdue struct {
	DurationMonths int `json:"duration_months"`
	MonthsAgo      int `json:"months_ago"`
}

type application struct {
	BusinessID  string    `json:"business_request_id"`
	OperationID string    `json:"operation_id"`
	Revision    string    `json:"revision"`
	EvidenceID  string    `json:"evidence_id"`
	Customer    string    `json:"customer"`
	AssetsUSD   int64     `json:"assets_usd"`
	AssetReport bool      `json:"asset_report_present"`
	ReferenceAt string    `json:"reference_time"`
	Overdue     []overdue `json:"overdue"`
}

type verdict struct {
	Kind      string `json:"kind"`
	CreditUSD int64  `json:"credit_usd"`
	Reason    string `json:"reason"`
}

// Deterministic S1 DOUBLE for the repository's fixed credit example. This
// is ordinary Go code, not a Jev model or a general semantic rule engine.
func judge(a application) verdict {
	for _, o := range a.Overdue {
		if o.DurationMonths > 3 {
			return verdict{"deny", 0, "历史存在持续超过三个月的逾期，触发优先否决"}
		}
	}
	if !a.AssetReport {
		return verdict{"needs-evidence", 0, "资产材料缺失，补证后再判断"}
	}
	if a.AssetsUSD > 30_000_000 {
		for _, o := range a.Overdue {
			if o.MonthsAgo <= 6 {
				return verdict{"uncovered", 0, "近期存在短期逾期：高档批准条件不成立，也未触发长期否决"}
			}
		}
		return verdict{"grant", 50_000, "资产超过三千万美元，近六个月无逾期"}
	}
	if a.AssetsUSD < 1_000_000 {
		return verdict{"grant", 10_000, "资产低于一百万美元且无长期逾期否决"}
	}
	return verdict{"uncovered", 0, "资产处于示例规则未覆盖的范围"}
}

type bankPayload struct {
	BusinessID string `json:"business_request_id"`
	Customer   string `json:"customer"`
	CreditUSD  int64  `json:"credit_usd"`
}

type receipt struct {
	Reference string      `json:"reference"`
	Payload   bankPayload `json:"payload"`
	Digest    string      `json:"payload_digest"`
}

type bankState struct {
	SubmitCalls int                `json:"submit_calls"`
	QueryCalls  int                `json:"query_calls"`
	Effects     int                `json:"created_credit_lines"`
	Receipts    map[string]receipt `json:"receipts"`
}

var errLostReply = errors.New("模拟银行已受理，但响应丢失")

// File-backed BANK DOUBLE: durable receipts, provider-side deduplication and
// query-by-business-ID are explicitly supplied assumptions, not framework guarantees.
type bank struct{ path string }

func (b bank) state() (bankState, error) {
	s := bankState{Receipts: map[string]receipt{}}
	err := readJSON(b.path, &s)
	if errors.Is(err, os.ErrNotExist) {
		return s, nil
	}
	return s, err
}

func (b bank) submit(p bankPayload, loseReply bool) (receipt, error) {
	if p.BusinessID == "" || p.Customer == "" || p.CreditUSD <= 0 {
		return receipt{}, errors.New("bank: business identity and positive credit amount required")
	}
	s, err := b.state()
	if err != nil {
		return receipt{}, err
	}
	digest, err := payloadDigest(p)
	if err != nil {
		return receipt{}, err
	}
	s.SubmitCalls++
	r, exists := s.Receipts[p.BusinessID]
	if exists && r.Digest != digest {
		if err := writeJSON(b.path, s); err != nil {
			return receipt{}, err
		}
		return receipt{}, fmt.Errorf("bank: same business request has different payload")
	}
	if !exists {
		r = receipt{"bank-demo-" + p.BusinessID, p, digest}
		s.Receipts[p.BusinessID] = r
		s.Effects++
	}
	if err := writeJSON(b.path, s); err != nil {
		return receipt{}, err
	}
	if loseReply {
		return receipt{}, errLostReply
	}
	return r, nil
}

func (b bank) query(id string) (receipt, bool, error) {
	s, err := b.state()
	if err != nil {
		return receipt{}, false, err
	}
	s.QueryCalls++
	if err := writeJSON(b.path, s); err != nil {
		return receipt{}, false, err
	}
	r, ok := s.Receipts[id]
	return r, ok, nil
}

type counters struct {
	S1Calls int `json:"s1_double_calls"`
	S2Calls int `json:"s2_double_calls"`
}

type stored struct {
	CommandDigest string  `json:"command_digest"`
	Receipt       receipt `json:"receipt"`
	Decision      verdict `json:"decision"`
}

type trace struct {
	Layer  string `json:"layer"`
	Action string `json:"action"`
	Detail string `json:"detail"`
}

type engine struct {
	dir       string
	rule      string
	machine   *adopt.Machine
	bank      bank
	afterBank func() error
}

func openEngine(dir, rule string) (*engine, error) {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return nil, err
	}
	m, err := adopt.Open(filepath.Join(dir, "adopt.jsonl"))
	if err != nil {
		return nil, err
	}
	return &engine{dir: dir, rule: rule, machine: m, bank: bank{filepath.Join(dir, "bank.json")}}, nil
}

func (e *engine) log(layer, action, detail string) error {
	b, err := json.Marshal(trace{layer, action, detail})
	if err != nil {
		return err
	}
	f, err := os.OpenFile(filepath.Join(e.dir, "trace.jsonl"), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	_, err = f.Write(append(b, '\n'))
	closeErr := f.Close()
	if err != nil {
		return err
	}
	return closeErr
}

func (e *engine) count(s2 bool) error {
	c, err := e.counts()
	if err != nil {
		return err
	}
	if s2 {
		c.S2Calls++
	} else {
		c.S1Calls++
	}
	return writeJSON(filepath.Join(e.dir, "counters.json"), c)
}

func (e *engine) counts() (counters, error) {
	var c counters
	err := readJSON(filepath.Join(e.dir, "counters.json"), &c)
	if errors.Is(err, os.ErrNotExist) {
		err = nil
	}
	return c, err
}

func grants() []ledger.Grant {
	return []ledger.Grant{
		{Kind: "data", Selector: "ledger://credit/demo", Actions: []string{"read"}, Effect: "none"},
		{Kind: "tool", Selector: "tool://credit-demo/submit", Actions: []string{"execute"}, Effect: "irreversible"},
	}
}

func (e *engine) request(a application) (*adopt.Request, error) {
	digest, err := payloadDigest(a)
	if err != nil {
		return nil, err
	}
	return &adopt.Request{
		AdoptionVersion: "0.1.0",
		Identity:        ledger.Identity{BusinessRequestID: a.BusinessID, EvaluationRevision: a.Revision, OperationID: a.OperationID},
		Versions:        ledger.Versions{Evidence: a.EvidenceID, Rule: e.rule, Model: "demo-s1-double-v1", Readout: "demo-typed-v1", Policy: "demo-fixed-policy-v1"},
		PayloadDigest:   digest, GrantsAtDecision: grants(),
		Accountability: &ledger.Accountability{Session: "local-credit-demo", Skill: "credit", Role: "demo-employee"},
	}, nil
}

type response struct {
	Status  string  `json:"status"`
	Receipt receipt `json:"receipt"`
	Verdict verdict `json:"verdict"`
}

func (e *engine) submit(a application, revoke, loseReply bool) (response, error) {
	if a.OperationID == "" || a.BusinessID == "" || a.Customer == "" {
		return response{}, errors.New("demo host: required business identity missing")
	}
	req, err := e.request(a)
	if err != nil {
		return response{}, err
	}
	state := e.machine.State(a.OperationID)
	if state == adopt.StateCompleted {
		var old stored
		if err := readJSON(filepath.Join(e.dir, "result-"+safeID(a.OperationID)+".json"), &old); err != nil {
			return response{}, err
		}
		if old.CommandDigest == req.PayloadDigest {
			if err := e.log("demo-host", "复用原结果", "同一操作与相同结构化输入；模型和银行提交均不调用"); err != nil {
				return response{}, err
			}
			return response{"reused", old.Receipt, old.Decision}, nil
		}
	}
	if state != "" {
		o := e.machine.Adopt(req)
		if err := e.log("project-core", "重复登记检查", o.Status+": "+o.Detail); err != nil {
			return response{}, err
		}
		if o.Status == "rejected" {
			return response{Status: "conflict"}, nil
		}
		return response{Status: "in-progress"}, nil
	}
	if err := writeJSON(filepath.Join(e.dir, "application.json"), a); err != nil {
		return response{}, err
	}
	if err := e.count(false); err != nil {
		return response{}, err
	}
	v := judge(a)
	if err := e.log("simulated-model", "S1 判断", v.Kind+": "+v.Reason); err != nil {
		return response{}, err
	}
	// This routing is DEMO HOST policy, not a built-in business correctness gate.
	if v.Kind != "grant" {
		if err := writeJSON(filepath.Join(e.dir, "last-decision.json"), v); err != nil {
			return response{}, err
		}
		return response{Status: v.Kind, Verdict: v}, nil
	}
	o := e.machine.Adopt(req)
	if err := e.log("project-core", "登记操作并绑定版本", o.Status); err != nil {
		return response{}, err
	}
	if o.Status != "adopted" {
		return response{}, fmt.Errorf("adopt: %s: %s", o.Status, o.Detail)
	}
	live := grants()
	if revoke {
		live = live[:1] // Simulated authority provider revoked execution permission.
	}
	o = e.machine.Dispatch(a.OperationID, live)
	if err := e.log("project-core", "派发前检查当前权限", o.Status+": "+o.Detail); err != nil {
		return response{}, err
	}
	if o.Status == "rejected" {
		return response{Status: "blocked", Verdict: v}, nil
	}
	if o.Status != "dispatched" {
		return response{}, fmt.Errorf("dispatch: %s", o.Status)
	}
	r, err := e.bank.submit(bankPayload{a.BusinessID, a.Customer, v.CreditUSD}, loseReply)
	if errors.Is(err, errLostReply) {
		o = e.machine.MarkUnresolved(a.OperationID)
		if o.Status != "unresolved" {
			return response{}, fmt.Errorf("mark unresolved: %s", o.Status)
		}
		if err := e.log("simulated-bank", "银行已办理但模拟响应丢失", "业务效果已保存；框架状态为结果未知"); err != nil {
			return response{}, err
		}
		return response{Status: "unresolved", Verdict: v}, nil
	}
	if err != nil {
		return response{}, err
	}
	if err := e.log("simulated-bank", "创建授信并保存回执", r.Reference); err != nil {
		return response{}, err
	}
	if e.afterBank != nil {
		if err := e.afterBank(); err != nil {
			return response{}, err
		}
	}
	if err := e.finish(a, req.PayloadDigest, v, r); err != nil {
		return response{}, err
	}
	return response{"completed", r, v}, nil
}

func (e *engine) finish(a application, digest string, v verdict, r receipt) error {
	if err := writeJSON(filepath.Join(e.dir, "result-"+safeID(a.OperationID)+".json"), stored{digest, r, v}); err != nil {
		return err
	}
	o := e.machine.MarkCompleted(a.OperationID)
	if o.Status != "completed" {
		return fmt.Errorf("complete: %s: %s", o.Status, o.Detail)
	}
	return e.log("project-core", "记录操作完成", o.Status)
}

func (e *engine) reconcile(a application) (receipt, error) {
	// Host explicitly queries the provider; the reference machine does not do this.
	r, ok, err := e.bank.query(a.BusinessID)
	if err != nil {
		return receipt{}, err
	}
	if !ok {
		return receipt{}, errors.New("bank result not found; retain unresolved state, do not blindly resend")
	}
	if err := e.log("demo-host", "沿原业务编号查询银行", r.Reference); err != nil {
		return receipt{}, err
	}
	if e.machine.State(a.OperationID) == adopt.StateDispatched {
		if o := e.machine.MarkUnresolved(a.OperationID); o.Status != "unresolved" {
			return receipt{}, fmt.Errorf("reconcile mark unresolved: %s", o.Status)
		}
	}
	req, err := e.request(a)
	if err != nil {
		return receipt{}, err
	}
	if err := e.finish(a, req.PayloadDigest, judge(a), r); err != nil {
		return receipt{}, err
	}
	return r, nil
}

func (e *engine) supplement(a application) (application, bool, error) {
	if err := e.count(true); err != nil {
		return a, false, err
	}
	// S2 DOUBLE retrieves a fixed fixture; it is not making a model/tool call.
	if err := writeJSON(filepath.Join(e.dir, "asset-source.json"), map[string]any{"customer": a.Customer, "assets_usd": a.AssetsUSD, "source": "simulated-asset-provider"}); err != nil {
		return a, false, err
	}
	l, err := evidlog.Open(filepath.Join(e.dir, "evidence.jsonl"))
	if err != nil {
		return a, false, err
	}
	selector := "customer=" + a.Customer
	if err := l.Append(evidlog.Entry{Tool: "demo.assets", Selector: selector, Caller: "s2-double", At: time.Now().UTC().Format(time.RFC3339), ResultRef: "asset-source.json"}); err != nil {
		return a, false, err
	}
	coverage := evidlog.CheckCoverage(l, evidlog.Declaration{QueriedTool: "demo.assets", Selector: selector})
	if err := writeJSON(filepath.Join(e.dir, "coverage.json"), coverage); err != nil {
		return a, false, err
	}
	a.AssetReport = true
	a.EvidenceID = "evidence-supplemented-v2"
	return a, coverage.OK, e.log("project-core", "核对补证调用记录", fmt.Sprintf("coverage=%v", coverage.OK))
}

func readJSON(path string, value any) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, value)
}

// The existing JCS entry accepts parsed JSON, not arbitrary Go structs.
func payloadDigest(value any) (string, error) {
	b, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	dec.UseNumber()
	var doc map[string]any
	if err := dec.Decode(&doc); err != nil {
		return "", err
	}
	return adopt.Digest(doc)
}

// New demo-only snapshot store. This does not replace the project's resultStore.
// It assumes a single writer and makes no power-loss or multi-instance claim.
func writeJSON(path string, value any) error {
	b, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".credit-demo-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err := f.Write(append(b, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func baseApplication(id string) application {
	return application{"br-" + id, "op-" + id, "er-1", "evidence-v1", "customer-demo", 35_000_000, true, "2026-11-19T00:00:00Z", nil}
}

func safeID(id string) string {
	// Demo-generated identifiers only; never turn arbitrary input into file paths.
	var out []byte
	for _, c := range []byte(id) {
		if c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '-' {
			out = append(out, c)
		}
	}
	return string(out)
}
