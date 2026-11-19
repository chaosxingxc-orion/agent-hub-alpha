// credit-demo is a local scenario demonstrator, not an enterprise SDK or
// openJiuwen integration. See demo/credit-scenarios/README.md for ownership.
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/openjiuwen/agent-festival/go/internal/judgment"
)

type check struct {
	Name     string `json:"name"`
	Expected string `json:"expected"`
	Actual   string `json:"actual"`
	Passed   bool   `json:"passed"`
}

type artifact struct {
	Name string `json:"name"`
	Href string `json:"href"`
}

type scenario struct {
	ID        string     `json:"id"`
	Title     string     `json:"title"`
	Kind      string     `json:"kind"`
	Purpose   string     `json:"purpose"`
	Status    string     `json:"status"`
	Checks    []check    `json:"checks"`
	Trace     []trace    `json:"trace"`
	Artifacts []artifact `json:"artifacts"`
	Error     string     `json:"error,omitempty"`
	knownGap  bool
}

func (s *scenario) expect(name string, expected, actual any) {
	want, got := fmt.Sprint(expected), fmt.Sprint(actual)
	s.Checks = append(s.Checks, check{name, want, got, want == got})
}

type report struct {
	GeneratedAt string     `json:"generated_at"`
	Source      string     `json:"reference_commit"`
	RuleVersion string     `json:"rule_version"`
	RuleSHA256  string     `json:"rule_sha256"`
	Passed      int        `json:"passed"`
	Gaps        int        `json:"gaps"`
	Failed      int        `json:"failed"`
	Errors      int        `json:"errors"`
	Cases       []scenario `json:"scenarios"`
}

func main() {
	output := flag.String("output", "", "fresh report directory (default: a new temporary directory)")
	selected := flag.String("scenario", "all", "all, or comma-separated scenario IDs")
	worker := flag.String("worker-dir", "", "internal crash worker; do not invoke manually")
	flag.Parse()
	if flag.NArg() != 0 {
		fmt.Fprintln(os.Stderr, "unexpected positional arguments")
		os.Exit(2)
	}
	spec, version, err := loadSpec()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if *worker != "" {
		if err := crashWorker(*worker, version); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		return
	}
	cases, err := selectCases(*selected)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	dir, err := freshDirectory(*output)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := os.WriteFile(filepath.Join(dir, "rule-source.json"), spec, 0o600); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	sum := sha256.Sum256(spec)
	r := report{GeneratedAt: time.Now().UTC().Format(time.RFC3339), Source: "unknown", RuleVersion: version, RuleSHA256: fmt.Sprintf("%x", sum)}
	if b, err := exec.Command("git", "rev-parse", "HEAD").Output(); err == nil {
		r.Source = strings.TrimSpace(string(b))
	}
	for _, definition := range cases {
		s := definition.scenario
		caseDir := filepath.Join(dir, s.ID)
		var e *engine
		var runErr error
		if s.ID == "restart" {
			// The child must be the only machine owner until it is killed.
			runErr = os.MkdirAll(caseDir, 0o700)
			e = &engine{dir: caseDir, rule: version}
		} else {
			e, runErr = openEngine(caseDir, version)
		}
		if runErr == nil {
			runErr = definition.run(&s, e)
		}
		s.Status = "PASS"
		for _, c := range s.Checks {
			if !c.Passed {
				s.Status = "FAIL"
				if s.knownGap {
					s.Status = "GAP"
				}
			}
		}
		if len(s.Checks) == 0 && runErr == nil {
			runErr = errors.New("scenario produced no assertions")
		}
		if runErr != nil {
			s.Status, s.Error = "ERROR", runErr.Error()
		}
		if err := collectArtifacts(&s, caseDir); err != nil {
			s.Status, s.Error = "ERROR", err.Error()
		}
		switch s.Status {
		case "PASS":
			r.Passed++
		case "GAP":
			r.Gaps++
		case "FAIL":
			r.Failed++
		case "ERROR":
			r.Errors++
		}
		r.Cases = append(r.Cases, s)
		fmt.Printf("%-5s %-15s %s\n", s.Status, s.ID, s.Title)
		if s.Error != "" {
			fmt.Printf("      %s\n", s.Error)
		}
		for _, c := range s.Checks {
			if !c.Passed {
				fmt.Printf("      %s: expected=%s actual=%s\n", c.Name, c.Expected, c.Actual)
			}
		}
	}
	if err := writeJSON(filepath.Join(dir, "report.json"), r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	if err := renderReport(filepath.Join(dir, "report.html"), r); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	fmt.Printf("\nPASS=%d GAP=%d FAIL=%d ERROR=%d\nReport: %s\n", r.Passed, r.Gaps, r.Failed, r.Errors, filepath.Join(dir, "report.html"))
	if r.Errors > 0 {
		os.Exit(2)
	}
	if r.Gaps+r.Failed > 0 {
		os.Exit(1)
	}
}

func loadSpec() ([]byte, string, error) {
	if judgment.SchemaPath == "" {
		return nil, "", errors.New("judgment schema unavailable; run from agent-festival/go using the supplied launcher")
	}
	path := filepath.Join(filepath.Dir(judgment.SchemaPath), "examples", "judgment", "valid", "judgment-spec-credit.json")
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, "", err
	}
	// The stand-in implements one reviewed fixture, not arbitrary rule text.
	const qualifiedSHA = "92cbb5f50961c702ba5d98cff36d68b147c4d00fefa87cddfa16ce2955ed943f"
	if digest := fmt.Sprintf("%x", sha256.Sum256(b)); digest != qualifiedSHA {
		return nil, "", errors.New("credit fixture changed: review the deterministic stand-in and qualify a new fingerprint before running")
	}
	var doc map[string]any
	if err := json.Unmarshal(b, &doc); err != nil {
		return nil, "", err
	}
	delete(doc, "_expect")
	delete(doc, "_expect_match")
	if es := judgment.CheckSpec(doc); len(es) > 0 {
		return nil, "", fmt.Errorf("repository judgment specification rejected: %v", es)
	}
	question, ok := doc["question"].(map[string]any)
	if !ok {
		return nil, "", errors.New("question missing from credit specification")
	}
	version, ok := question["spec_version"].(string)
	if !ok || version == "" {
		return nil, "", errors.New("credit specification has no version")
	}
	return b, version, nil
}

func freshDirectory(path string) (string, error) {
	if path == "" {
		return os.MkdirTemp("", "agent-hub-credit-demo-")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	// An exclusive directory creation prevents accidental overwrite/reuse.
	if err := os.Mkdir(abs, 0o700); err != nil {
		return "", fmt.Errorf("output must be a new directory: %w", err)
	}
	return abs, nil
}

func collectArtifacts(s *scenario, dir string) error {
	files, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, f := range files {
		if !f.IsDir() && !strings.HasPrefix(f.Name(), ".") {
			s.Artifacts = append(s.Artifacts, artifact{f.Name(), s.ID + "/" + f.Name()})
		}
	}
	b, err := os.ReadFile(filepath.Join(dir, "trace.jsonl"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	for {
		var t trace
		if err := dec.Decode(&t); err == io.EOF {
			break
		} else if err != nil {
			return err
		}
		s.Trace = append(s.Trace, t)
	}
	return nil
}

func renderReport(path string, r report) error {
	t, err := template.New("report").Parse(reportHTML)
	if err != nil {
		return err
	}
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	if err := t.Execute(f, r); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

const reportHTML = `<!doctype html><html lang="zh-CN"><meta charset="utf-8"><meta name="viewport" content="width=device-width,initial-scale=1"><title>Agent Hub Alpha · 本地授信场景验证</title>
<style>body{margin:0;background:#f4f6fa;color:#182331;font:16px/1.65 system-ui,sans-serif}main{max-width:1060px;margin:auto;padding:36px 22px}h1{font-size:30px;margin:0 0 8px}h2{font-size:20px}p{margin:10px 0}.muted{color:#536279}.banner,article,.scope{background:white;border:1px solid #d9e1eb;border-radius:12px;padding:22px;margin:18px 0}.stats{display:flex;gap:12px;flex-wrap:wrap}.stats span{padding:10px 18px;border-radius:8px;background:#e8edf5}.badge{font-size:13px;font-weight:700;border-radius:5px;padding:4px 9px;margin-right:10px}.PASS{background:#d6f2df;color:#165c32}.GAP,.FAIL{background:#fde4cd;color:#8a3b0b}.ERROR{background:#fbd8d8;color:#8c2121}table{border-collapse:collapse;width:100%;font-size:14px;margin:12px 0}th,td{text-align:left;border-bottom:1px solid #e1e6ee;padding:9px;vertical-align:top;word-break:break-word}code{font-size:13px;overflow-wrap:anywhere}.failed{color:#a32d17;font-weight:600}.trace{border-left:3px solid #ced9e8;padding-left:14px;margin:12px 0}.trace small{color:#607087}summary{cursor:pointer;font-weight:600}a{color:#175b9b}.files{display:flex;gap:12px;flex-wrap:wrap}.flow{display:flex;flex-wrap:wrap;gap:8px;align-items:center}.flow b{padding:8px 12px;background:#edf2f8;border-radius:6px;font-size:14px}@media(max-width:650px){main{padding:22px 12px}article{padding:14px}h1{font-size:24px}th,td{padding:6px}}</style>
<main><p class="muted">Agent Hub Alpha / 本地场景实验</p><h1>一笔授信，从申请到可追溯结果</h1><p>通过真实 Go 参考组件，观察请求重试、权限变化和中断后的实际行为。</p>
<div class="banner"><strong>运行边界</strong><p>S1、S2 和银行均为明确的模拟实现；未接入真实模型、银行或 openJiuwen。绿色表示本场景断言通过，GAP 表示已观察到的系统缺口，不能计为通过。</p><div class="stats"><span>通过 {{.Passed}}</span><span>缺口 {{.Gaps}}</span><span>失败 {{.Failed}}</span><span>执行错误 {{.Errors}}</span></div></div>
<div class="scope"><strong>这次分别运行了什么</strong><table><tr><th>来源</th><th>职责</th></tr><tr><td>项目原有 Go 组件</td><td>采纳/去重、派发权限检查、状态保存、取证覆盖检查、隔离控制回放</td></tr><tr><td>本次新增 Demo 宿主</td><td>业务流程、结果快照、银行查询与对账、进程重启编排、报告</td></tr><tr><td>模拟提供方</td><td>确定性的 S1 判断、S2 资料补充，以及支持去重与查询的本地银行</td></tr></table><div class="flow"><b>模拟材料与判断</b><span>→</span><b>项目登记和派发检查</b><span>→</span><b>模拟银行回执</b><span>→</span><b>宿主保存/恢复</b></div><p class="muted">示例采用整美元资产值、整数逾期月数及固定参考时间，不验证真实日历语义。规则文件经过项目格式检查；S1 仍是针对固定样例编写的普通代码。</p></div>
{{range .Cases}}<article><h2><span class="badge {{.Status}}">{{.Status}}</span>{{.Title}}</h2><p class="muted">{{.ID}} · {{.Kind}}</p><p>{{.Purpose}}</p>{{if .Error}}<p class="failed">执行错误：{{.Error}}</p>{{end}}<table><tr><th>检查项</th><th>期望</th><th>实际</th><th>结果</th></tr>{{range .Checks}}<tr><td>{{.Name}}</td><td>{{.Expected}}</td><td>{{.Actual}}</td><td>{{if .Passed}}通过{{else}}<span class="failed">未通过</span>{{end}}</td></tr>{{end}}</table><details><summary>查看执行过程与证据</summary>{{range .Trace}}<div class="trace"><small>{{.Layer}}</small><br><strong>{{.Action}}</strong><br>{{.Detail}}</div>{{end}}<div class="files">{{range .Artifacts}}<a href="{{.Href}}">{{.Name}}</a>{{end}}</div></details></article>{{end}}
<p class="muted">生成时间 {{.GeneratedAt}}<br>参考代码提交 <code>{{.Source}}</code><br>规则版本 {{.RuleVersion}} · SHA-256 <code>{{.RuleSHA256}}</code></p><p><a href="report.json">结构化报告</a> · <a href="rule-source.json">本次规则原件</a></p></main></html>`
