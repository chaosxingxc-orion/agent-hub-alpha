# Local credit scenario demonstrator

本地授信 Demo：用真实 Go 控制组件串起一笔模拟授信，观察重试、权限变化、
补证和中断恢复，并生成带逐项断言和原始证据链接的 HTML 报告。

从仓库根目录运行 `bash demo/credit-scenarios/run.sh`；需要 Go 1.24+ 和 Bash。
当前默认运行 12 个场景，预期为 `PASS=10 GAP=1 FAIL=1 ERROR=0`，退出码为 1，
因为有效修订保护缺口和权限拆分误拒绝仍未修复。

用户指定的早期 11 场景结果保存在 [sample-report](sample-report/README.md)，
可下载仓库后用浏览器打开其中的 [report.html](sample-report/report.html)。
该历史报告未包含后加的 `grant-regroup`，不能代表当前 12 场景已全部通过。

This executable demonstration reuses Agent Hub Alpha's **actual Go reference
components**. It runs a small credit workflow, saves its evidence, and produces
a readable HTML report plus machine-readable JSON.

It is a local scenario experiment, not a released SDK, openJiuwen adapter,
real-model evaluation, or production banking integration. No model key or
business credentials are needed. Existing core code is unchanged.

## Run

Requirements: Go 1.24 or newer and Bash. Initial dependency resolution needs
access to the Go modules declared in `agent-festival/go/go.mod`; there are no
additional module dependencies. Set `GO_BIN` if Go is not on `PATH`.

From the repository root:

```bash
bash demo/credit-scenarios/run.sh
```

The command creates a fresh directory under the system temporary directory and
prints the path to `report.html`. Open that file in a browser. It contains
per-scenario checks, an execution trace, and links to the actual state files.
The launcher builds a temporary executable and removes only that executable
after the run; it preserves the report and all scenario evidence.

Run selected scenarios or choose a new output directory:

```bash
bash demo/credit-scenarios/run.sh -scenario happy,retry,revoked
bash demo/credit-scenarios/run.sh -output /tmp/my-fresh-credit-report
```

The explicit output directory must not already exist; the command never clears
or overwrites an existing run. Its parent directory must exist.

The underlying command can also be built normally from the Go module:

```bash
cd agent-festival/go
go build -buildvcs=false -o /tmp/credit-demo ./cmd/credit-demo
/tmp/credit-demo -scenario happy,retry,replay
```

Run the binary with `agent-festival/go` as its working directory, because the
existing schema discovery uses relative paths. The launcher handles this.
The temporary binary omits Go's automatic VCS stamp for worktree compatibility;
the report still queries and records Git HEAD explicitly.

## Read the result correctly

The full run includes one **unclosed integration requirement** and one
newly reproduced **permission representation defect**.
At the current reference baseline it reports:

```text
PASS=10 GAP=1 FAIL=1 ERROR=0
```

This is not an all-green acceptance result. `GAP` means assertions were not
satisfied for an existing integration gap; `FAIL` denotes other unmet
assertions. Neither is counted as a pass. Exit codes are:

| Code | Meaning |
|---|---|
| 0 | Every selected scenario assertion passed |
| 1 | At least one assertion failed, including a known `GAP` |
| 2 | Setup, execution, or evidence-collection error |

Unexpected execution errors are always `ERROR`, even in a known-gap scenario.
If the revision integration is later implemented, the same requirement should
pass rather than being kept red by changing the assertion.

## Business fixture

The example follows the repository's fixed credit specification:

- Assets above USD 30 million and no overdue in the preceding six months:
  eligible for USD 50,000.
- Assets below USD 1 million: eligible for USD 10,000, unless the historical
  veto applies.
- An overdue lasting more than three months anywhere in history overrides
  the asset rules and denies credit.
- Missing evidence and uncovered rules are separate outcomes, not automatic
  refusals or approvals.

The demonstrator uses whole-dollar assets, integer month durations/recencies,
and a fixed application reference time. It does not validate real calendar
arithmetic. These are fictional scenario inputs, not financial advice or a
bank's real underwriting policy.

The original fixture is validated by `judgment.CheckSpec`, copied into the
report, and fingerprinted. The deterministic S1 double implements this one
reviewed snapshot in ordinary Go code. It does not execute arbitrary natural
language rules. The command rejects changed fixture bytes until the double
and its qualification are reviewed; changing a version label alone does not
update the executable policy.

## Scenarios

| ID | Input or event | Required observation | Main owner |
|---|---|---|---|
| `happy` | Complete evidence, USD 35m assets, clean history | Completed; USD 50k; one bank effect | Project control + demo host + doubles |
| `retry` | Repeat the same operation and structured input | Original receipt; one S1 call and one bank submit overall | Project identity + demo result store |
| `conflict` | Same operation ID, different customer | Conflict; original receipt preserved; no second submit | Project adoption machine |
| `revoked` | Revoke execution after registration | Dispatch refused; durable control refusal; no bank effect | Project dispatch gate + simulated authority |
| `lost-reply` | Provider saves its effect but returns a lost-reply error | Unresolved; repeated request does not resend; query completes it | Project state machine + host reconciliation |
| `restart` | Terminate a child after bank persistence, before host completion | Dispatch state survives; no new permit; original-ID query completes | Real subprocess termination + project loader + host |
| `evidence` | Missing asset report, then fixed evidence provider supplies it | Separate initial outcome; recorded query; er-2 evidence revision; one bank effect | Host routing + project evidence log |
| `uncovered` | USD 35m assets; one-month overdue one month ago | Uncovered; no bank submit | S1 double + host policy |
| `history-veto` | Four-month overdue 24 months ago | Deny; no bank submit | S1 double + host policy |
| `revision-gap` | Explicit scenario policy says er-2 supersedes er-1 | Old operation must not dispatch; current prototype still permits it | Known unimplemented system integration |
| `replay` | Recorded successful controller inputs and events | Fidelity match; ledger conformance pass; no new bank submit | Project replay component |
| `grant-regroup` | The same tool's write/execute authority is split into two entries, without loss of any permission | Graph admission and dispatch should agree; current dispatch wrongly rejects the split form | New normal-lifecycle defect in project authority coverage |

`revision-gap` writes `correction-policy.json`, explicitly describing the
intended replacement. The reference machine currently has no interface that
consumes that authority. The demo observes the missing integration; it does
not pretend that two independent registrations automatically revoke each
other, or add its own replacement guard and attribute that success to the
project. The only banking effect here is in the local simulated provider.

`grant-regroup` uses valid inputs and an ordinary IAM role regrouping. It
checks the unchanged actions, resource and effect envelope through graph
admission, uses grouped authority as a positive dispatch control, and then
exercises split authority on the actual operation. The control only tests
permission; it never submits to the bank. The split operation is currently
blocked, so the legitimate application produces zero bank effects rather
than the required one. No corrupt data, malformed model output or permission
revocation is injected in this case. The proposed correction must retain
rejection of genuinely missing authority.

## What is reused, new, and simulated?

| Part | Provenance and responsibility |
|---|---|
| `internal/adopt` | Existing operation registration, digest conflict detection, pinned version/authority records, dispatch permission checks, durable state transitions |
| `internal/compose` | Existing flow admission and atomic permission aggregation, used to compare equivalent authority representations |
| `internal/ledger` | Existing identity, version and event representations, plus load-time event checking |
| `internal/evidlog` | Existing query records and declaration/record coverage checks |
| `internal/replay` | Existing scratch-isolated control replay and separate fidelity/conformance verdicts |
| `internal/judgment` | Existing specification structure validation; not business correctness evaluation |
| Workflow and result snapshots | **New demo host code** in `cmd/credit-demo`; completed results are read from disk by the demo, not by a newly delivered SDK |
| Recovery and reporting | **New demo host code**: explicitly query the provider, save its returned result, mark completion, and generate traces/reports |
| S1 and S2 | **Deterministic doubles**, not model services; S2 supplies fixed local evidence |
| Bank | **File-backed double** with native deduplication and query-by-business-ID; no HTTP service or actual banking API |
| Authority | **Simulated grant provider**, not a connection to enterprise IAM |

The native bank deduplication/query contract is an explicit supplied
assumption, tested in `runtime_test.go`. The framework cannot manufacture
these capabilities for an arbitrary external provider.

The model call counter sits on the exercised S1-double path; the bank counters
are read from the persisted provider state. The normal execution is a positive
control, so a retry showing zero additional calls does not rely on a counter
that was never connected.

## Failure and restart experiment

`lost-reply` is an explicit provider-error injection: the bank snapshot is
saved before the double returns an error. It is not a test of actual network
timeouts.

`restart` starts a separate executable. The worker emits a readiness marker
only after the project has recorded dispatch and the simulated bank has saved
its receipt. The parent then kills that worker. After termination, a newly
opened controller reconstructs the state, refuses duplicate dispatch, and
the host queries the original bank request. The parent does not keep another
active machine owner during the worker's lifetime. This observes process
interruption, not power loss, distributed failover, or a production scheduler.

## Evidence files

Each scenario gets an independent directory. Depending on the path taken:

- `application.json`: actual structured input sent through the demo.
- `adopt.jsonl`: events written by the existing project controller.
- `bank.json`: provider-side calls, created effects, and original receipts.
- `result-*.json`: demo-host result snapshots.
- `counters.json`: exercised S1/S2-double call counts.
- `trace.jsonl`: demo execution explanation, labelled by component origin.
- `evidence.jsonl`, `coverage.json`, `asset-source.json`: supplementation facts.
- `correction-policy.json`, `new-decision.json`: explicit revision requirement.
- `replay-case.json`, `replay-report.json`: preserved controller inputs,
  recorded events, and the actual isolated replay verdict.

At the report root, `report.json` carries every expected/actual assertion,
`report.html` renders them, and `rule-source.json` preserves the source fixture.

## Limits and existing audit issues

This demo assumes sequential use with one active state-machine owner per file.
It supplies a small new JSON result store rather than promoting the test-only
T2 host into production. Therefore it does not fix, close, or retest away the
previously reported T2 result-aliasing and write-failure issues.

The host itself chooses whether a deterministic business verdict permits a
submit. This is not evidence that the reference framework now has a universal
business judgment checker. The demo's refusal snapshots and explanation trace
are not a complete production audit subsystem. Full rule publication,
continuous real authority, durable scheduling, real models and openJiuwen
integration remain separate work.

## Verification

本次提交前的检查记录见 [verification.md](verification.md)。

From `agent-festival/go`:

```bash
go test ./cmd/credit-demo -count=1 -race
go test ./... -count=1 -race
```

Then run the executable scenarios. Keep the nonzero full-run exit code while
either recorded failure remains open; do not convert it into a passing
assertion merely to obtain a green demonstration.
