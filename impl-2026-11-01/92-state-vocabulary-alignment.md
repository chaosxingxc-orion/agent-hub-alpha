# 92 · 五状态 ↔ 六区分 ↔ 我方判定词（三源映射矩阵 v1）

> 76 号任务 #9 交付件＝71 号 A6/C8 开件。体裁沿用 67 号附 3/附 4：**逐字锚先行、每格给映射
> 依据（引文逐字+坐标）、无定义处不硬编、表尾只列裁项不给结论**。持有侧：本件是我方侧翼
> 对齐件（同 67 附 12 对附 3 归类表的处置——不随摘录件送达、不进预注册计分面），可供 74 号
> 回复池取用。凡本文粘命令处，输出均为本机实跑回填；执行目录除注明外＝
> `/Users/yaojun/sandbox-agent-b`。

## 〇、取法与快照声明（两邢源为 GH 线上件，本表按快照坐标引用）

**gh5 附件（源 A）**：issue #5 正文仅含该附件链接，真源是附件
`NGA架构方案评审与核心科学问题分析报告.md`。取法＝先 `gh api .../issues/5 --jq .body` 拿
URL，再 `curl` 取件（github.com 面须走本机代理 127.0.0.1:7897；api.github.com 面免代理）。

```
$ gh api repos/chaosxingxc-orion/agent-hub-alpha/issues/5 --jq .updated_at
2026-11-10T07:34:01Z
$ curl -sS -L -x http://127.0.0.1:7897 https://github.com/user-attachments/files/34616826/NGA.md -o /tmp/agentb/live9/NGA-live.md
$ diff /tmp/agentb/live9/NGA-live.md /tmp/hub-issues/gh5-attachment-NGA.md; echo rc=$?
rc=0
$ shasum -a 256 /tmp/agentb/live9/NGA-live.md /tmp/hub-issues/gh5-attachment-NGA.md /tmp/fs72/gh5-attachment-NGA.md
07832c5413d7c0a8eb97b17d4d702ba97a94cfeb25dcfb35c7bd66c267ce316b  /tmp/agentb/live9/NGA-live.md
07832c5413d7c0a8eb97b17d4d702ba97a94cfeb25dcfb35c7bd66c267ce316b  /tmp/hub-issues/gh5-attachment-NGA.md
07832c5413d7c0a8eb97b17d4d702ba97a94cfeb25dcfb35c7bd66c267ce316b  /tmp/fs72/gh5-attachment-NGA.md
```

**快照性声明（源 A）**：本表引用的是**快照件** `/tmp/hub-issues/gh5-attachment-NGA.md`
（#8 取法所留）；上列命令把线上件重新取回并比对——sha256 三面全等即「快照＝现行态」。
issue #5 自 `2026-11-10T07:34:01Z` 未动，附件 URL 指向固定文件号 34616826，故该快照的失效
方式是「作者换附件重传」而非「行内编辑」；若日后再引 gh5，必须重跑上列命令组，不得沿用行号。

**gh4 正文（源 B）**：取法同 76 号 #8（`gh api .../issues/4 --jq .body`）；快照件
`/tmp/hub-issues/gh4-full.md` 首行为采集头 `===== GH #4 全文 =====`，正文自第 2 行起。

```
$ gh api repos/chaosxingxc-orion/agent-hub-alpha/issues/4 --jq .updated_at
2026-11-10T12:48:47Z
$ gh api repos/chaosxingxc-orion/agent-hub-alpha/issues/4 --jq .body > /tmp/agentb/live9/gh4-live.md
$ diff /tmp/agentb/live9/gh4-live.md <(tail -n +2 /tmp/hub-issues/gh4-full.md); echo rc=$?   # 进程替换需 bash/zsh
rc=0
$ grep -n 必须区别 /tmp/agentb/live9/gh4-live.md /tmp/hub-issues/gh4-full.md /tmp/fs72/gh-issue-4.md
/tmp/agentb/live9/gh4-live.md:306:必须区别：未配置、明确配置为空、规则未覆盖、证据不足/冲突、能力执行失败、结果未知。发布前拒绝缺失必需规约；运行时遇到未覆盖场景，按已发布的补证、等待、复核或明确未决路线处理。没有获准降级规则时不能由模型补成许可，也不应默认解释为业务拒绝。
/tmp/hub-issues/gh4-full.md:307:必须区别：未配置、明确配置为空、规则未覆盖、证据不足/冲突、能力执行失败、结果未知。发布前拒绝缺失必需规约；运行时遇到未覆盖场景，按已发布的补证、等待、复核或明确未决路线处理。没有获准降级规则时不能由模型补成许可，也不应默认解释为业务拒绝。
/tmp/fs72/gh-issue-4.md:308:必须区别：未配置、明确配置为空、规则未覆盖、证据不足/冲突、能力执行失败、结果未知。发布前拒绝缺失必需规约；运行时遇到未覆盖场景，按已发布的补证、等待、复核或明确未决路线处理。没有获准降级规则时不能由模型补成许可，也不应默认解释为业务拒绝。
```

**三套行号坐标系（引用必带件名，否则差一格）**：§7 句在 `gh4-full.md:307`（快照头行占第 1
行）、在 `/tmp/fs72/gh-issue-4.md:308`（该件头部后多一空行）、在实时正文 `:306`（无头行）。
本表正文一律用 `gh4-full.md` 坐标。**快照性声明（源 B）**：diff 为空且 `updated_at` 未动，
故本表 §7/§8 引用与线上逐字等；gh4 是可编辑正文（与源 A 的附件形态不同），失效方式为行内
编辑，随取随比。

**我方件（源 C）**：只读沙箱 `agent-festival/`——本件**只新建这一个 md**，不改动任何代码或
契约文件（自证＝§六粘的本件提交文件面
`git show --name-only --format='' HEAD`，其输出只有本 md 一行）。工作树与并行的 #2/#4/#6/#8
建造位共用，故本件**不粘全局 `git status`/`git diff` 当「仓库干净」的证据**（那是别人的在途
改动，粘了就会随对方提交而失真）；`77/78/80` 号是兄弟建造位产物，本件不依赖也不改它们。

## 一、逐字锚（三源全部坐标；A=gh5，B=gh4，C=我方）

**A1** `gh5-attachment-NGA.md:197`（全文唯一命中）——
> 研究系统如何识别证据不足、规则冲突、输入异常和分布外样本，并区分不确定、拒绝和无法判断。建议至少区分 Determined、Uncertain、Unresolved、Conflict 和 Refused 五种状态。

**A2** 五态名在附件中只命中 1 行（`grep -c "Determined\|Uncertain\|Unresolved\|Conflict\|Refused" /tmp/hub-issues/gh5-attachment-NGA.md`＝1，同命令亦粘在 §六）——**邢方对五态无逐条定义**。本表所有
gh5 侧映射都是「依名＋A1 同句触发词/gloss（证据不足→Uncertain、规则冲突→Conflict、
输入异常与分布外样本→未指派、拒绝→Refused、无法判断→Unresolved）」的**推定映射**，不是
引用定义；推定映射的强度上限即 D4。

**B1** `gh4-full.md:307`——
> 必须区别：未配置、明确配置为空、规则未覆盖、证据不足/冲突、能力执行失败、结果未知。发布前拒绝缺失必需规约；运行时遇到未覆盖场景，按已发布的补证、等待、复核或明确未决路线处理。没有获准降级规则时不能由模型补成许可，也不应默认解释为业务拒绝。

**B2** `gh4-full.md:313`（六格之外的判定侧出口，节引）——
> 可用性验收除了拒绝不安全动作，还要证明有效流程有用：正常授信/拒绝、补证后继续、故障后恢复与获准知识快路径均可完成。总是拒绝不是系统目标。

**C1** `agent-festival/go/internal/e2e/e2e_test.go:38-39`（词表自述，五词）——
```
$ grep -n "vocabulary of judgment-spec-credit" -A1 agent-festival/go/internal/e2e/e2e_test.go
38:// vocabulary of judgment-spec-credit (deny / grant-50000 / grant-10000 /
39-// on_missing_evidence / on_undefined_rule), not the ledger vocabulary.
```

**C2** `agent-festival/go/internal/e2e/t1_llm_test.go:96`（真模型臂输出模板 enum，五词封闭）——
```
$ grep -n '{"verdict"' agent-festival/go/internal/e2e/t1_llm_test.go
96:{"verdict": "<grant-50000|grant-10000|deny|on_missing_evidence|on_undefined_rule>", "reason": "<one sentence>"}`
```

**C3** `e2e_test.go` s1Judge 发射点逐字（命令与输出原样粘）——
```
$ grep -n 's1Verdict{kind:' agent-festival/go/internal/e2e/e2e_test.go
66:			return s1Verdict{kind: "deny",
71:		return s1Verdict{kind: "on_missing_evidence", details: "asset valuation not retrievable — request evidence, record the issue, no automatic denial"}
84:				return s1Verdict{kind: "on_undefined_rule", details: fmt.Sprintf("upper-band premise unsatisfied: an overdue event (duration %d months, %d months ago) is inside the six-month window at any duration — no rule covers this case, no automatic deny (xingchao 198789782 §2)", o.durationMonths, o.monthsAgo)}
87:		return s1Verdict{kind: "grant-50000", details: "assets above the upper cutoff with a clean six-month window"}
90:		return s1Verdict{kind: "grant-10000", details: "assets below the lower cutoff"}
92:	return s1Verdict{kind: "on_undefined_rule", details: "assets between the bands — an open specification question for the owner, no default grant"}
```

**C4** `schema/examples/judgment/valid/judgment-spec-credit.json` 三 level 定义逐字——
`:31` `"eligible for the 50,000 USD line"`；`:36` `"eligible for the 10,000 USD line"`；
`:41` `"credit denied under the override rule"`。

**C5** 同件各 rule 的 precedence：`:47` `"precedence": 0`、`:52` `"precedence": 1`、
`:59` `"precedence": 2`（**互异，无同序位**）；`:55` exceptions 逐字——
> "assets exactly at the 30 cutoff are an open boundary — the outcome is on_undefined_rule until the owner settles the cutoff convention"

**C6** 同件 `:65/66` resolution 两键动作语义逐字——
> `"on_missing_evidence": "request admissible asset valuation; record a semantic issue for the owner — not an automatic denial"`
> `"on_undefined_rule": "apply no rule; record a semantic issue for the authorized business owner"`

**C7** `schema/judgment.schema.json:173-190`（resolution **闭集**）：`:176`
`"additionalProperties": false`；`:177-180` 的 `required` 数组每值单独成行，二值依次为
`"on_missing_evidence"`（`:178`）与 `"on_undefined_rule"`（`:179`）。

**C8** 同件 `:156`（Conflict 在我方契约里的唯一指名，节引）——
> equal precedence is a conflict the resolution section must name.

**C9** 同件 `:5`（我方自己的分层原理句，节引）——
> unknown/unresolved is distinct from refusal unless an authorized specification defines otherwise — overall-design §2

**C10** 同件 `:105-106`＋`:110-114`：levels 为 `"type": "array"` 且 `"minItems": 1`，每元素
`:110-114` 的 `required` 数组逐行列出 `"name"`／`"definition"`／`"type"`——「缺失必需规约
发布前拒绝」在我方是机械执法，
而「明确配置为空」在 levels 上**不可表达**（空数组被 minItems 挡）。

**C11** `go/internal/ledger/ledger.go:33/34`（C1 末句所指的 ledger vocabulary，另一层词表）——
```
$ grep -n "AdoptStatuses = \|EffectStatuses = " agent-festival/go/internal/ledger/ledger.go
33:var AdoptStatuses = map[string]bool{"adopted": true, "reused": true, "conflict": true}
34:var EffectStatuses = map[string]bool{"dispatched": true, "unresolved": true, "completed": true, "failed": true}
```

**C12** `t1_llm_test.go:90-93`（「不得默认」二句硬约束的真模型臂表述，逐字）——
```
$ grep -n "Open specification questions" -A3 agent-festival/go/internal/e2e/t1_llm_test.go
90:- Open specification questions: assets exactly at a cutoff, or strictly between 1 and 30, are
91-  UNSETTLED — output "on_undefined_rule" (record the issue; do NOT default to any grant).
92-- Missing evidence: if the asset valuation report cannot be retrieved, output
93-  "on_missing_evidence" (request the evidence; an automatic denial is forbidden).
```

**C13** 拒绝（refusal）在我方的落点在**控制层不是判定词层**：`e2e_test.go:203`
`// controlEvents returns the indices of control (refusal-record) events in`；
`correctness_probe_test.go:140` `// grant; the seat must refuse BEFORE adoption.`；且
`t1_llm_test.go` 内 `refus` 零命中（`grep -in refus … ; rc=1`，见 §六）——判定词 enum 里
没有任何 refusal 值。

**C14** `e2e_test.go:487` 逐字
`if o := h.machine.MarkUnresolved("op-out-1"); o.Status != "unresolved" {`——我方
`unresolved` 一词现属**效果对账**语义（已派发、结果未到），非判定语义。

**C15** `impl-2026-11-01/67-xs3-preregister.md` 附 3 归类表（本件体裁来源）——
```
$ grep -n "语义归类表预注册如下" -A2 impl-2026-11-01/67-xs3-preregister.md
91:原文；②双方对分类仍有分歧的格，**保守归「歧义」**（呈属主）；③语义归类表预注册如下
92-（同义词归此类后比较）：grant-50000≈「上带授信/五万/50000」、deny≈「拒绝授信」、
93-on_undefined_rule≈「未覆盖/无规则覆盖」、on_missing_evidence≈「缺证/证据不足」。
```

**C16** 同件附 4（五词口径改真，起始 `:95`）与附 12（更正，起始 `:138`，下粘其更正句
`:140`）＋夹具 `_expect` 节引——
```
$ grep -n "词汇表口径改真" -A1 impl-2026-11-01/67-xs3-preregister.md
95:**附 4（词汇表口径改真）**：§四模板例举四词与 §二.3「我方五词表本身是隔离物」自相矛盾
96-——实况：五词为公共线程已立词汇（56 号对邢已含其中三词），对邢方臂非隔离物；对弱形态
$ grep -n "按五词口径实为" impl-2026-11-01/67-xs3-preregister.md
140:「56 号对邢已含其中三词」按五词口径实为**四**词。附 1 的发现时序全链 git 可验
$ grep -n "demonstration assumptions pending the business owner" agent-festival/schema/examples/judgment/valid/judgment-spec-credit.json
123:  "_expect": "ACCEPT: the P0-A ontology split (v1.1.0) — the term 'overdue' carries no duration threshold (P3's definitional absorption reverted); the veto duration lives in rule 0's coverage alone, the window scope in the time_window convention alone. Units and thresholds remain demonstration assumptions pending the business owner; the unbounded veto is the jointly confirmed semantic (xingchao counterexample accepted, E3 validated); the any-duration window premise is xingchao's second counterexample accepted (198789782 §2, P0-A cell 1)",
```

## 二、映射矩阵

档位记号：**主**＝该格为唯一/首选归属；**次**＝条件性共名（条件在 §三钉死）；**—**＝不映；
**✗**＝假朋友（字面像而语义反，见 §四）。

### 表 2-1 我方词 → gh5 五状态（依 A1 的推定映射）

| 我方词＼gh5 态 | Determined | Uncertain | Unresolved | Conflict | Refused |
|---|---|---|---|---|---|
| grant-50000 | 主 | — | — | — | — |
| grant-10000 | 主 | — | — | — | — |
| deny | 主 | — | — | — | ✗ |
| on_missing_evidence | — | 主 | — | —（见 g7） | — |
| on_undefined_rule | — | — | 主 | 次（闭包代言，现不可达） | — |

**格依据**：

- **g1 grant-50000→Determined 主**：A1 里 Determined 是诸失效触发词之外唯一的「已出判断」
  名；C3 `:87` details 逐字 `assets above the upper cutoff with a clean six-month window`
  （前提齐备）；C4 `:31`。g2（grant-10000）同构：C3 `:90` `assets below the lower cutoff`＋
  C4 `:36`。
- **g3 deny→Determined 主、非 Refused**：C4 `:41` `credit denied under the override rule`——
  deny 由覆盖它的规则（precedence 0，C5 `:47`）产出的**业务答案**；B1 句末「也不应默认解释为
  业务拒绝」反证「业务拒绝」属已判定侧；C15 `:92` 早把 deny≈「拒绝授信」（不是裸「拒绝」）。
- **g4 Refused 无我方判定词**：C2 enum 封闭五值无 refusal；C13 证明我方拒绝在控制层。故该列
  对全部五词为 —／✗，唯一「✗」在 deny 行。
- **g5 on_missing_evidence→Uncertain 主**：A1 触发词「证据不足」对 gloss「不确定」；C3 `:71`
  `asset valuation not retrievable — request evidence, record the issue, no automatic denial`；
  C6 `:65`；C12 `:93` `an automatic denial is forbidden`。不映 Unresolved：该案未触及规则侧
  前提，词的定义只谈证据。
- **g6 on_undefined_rule→Unresolved 主／Conflict 次**：见 §三专项。
- **g7 承载面缺口（不是我方词的缺口）**：gh4 把「证据不足/冲突」合为一格（B1）、gh5 单列
  Conflict（A1），而我方 enum（C2）与闭集（C7）里既无「证据相互冲突」值也无「规则冲突」值
  ——**证据相互冲突目前在我方词表无处可说**。入 D2/D5。

### 表 2-2 我方词 → gh4 §7 六区分（B1）＋六格外出口（B2）

| 我方词＼gh4 区分 | 未配置 | 明确配置为空 | 规则未覆盖 | 证据不足/冲突 | 能力执行失败 | 结果未知 | 六格外：正常授信/拒绝 |
|---|---|---|---|---|---|---|---|
| grant-50000 | — | — | — | — | — | — | 主 |
| grant-10000 | — | — | — | — | — | — | 主 |
| deny | — | — | — | — | — | — | 主 |
| on_missing_evidence | — | — | — | 主（不足半格） | — | — | — |
| on_undefined_rule | — | — | 主 | 次（冲突半格） | — | — | — |
| （判定词无承载） | 无词 | 无词 | — | — | 无词：C11 `failed` 在机器层 | 无词：C11 `unresolved` 在机器层（C14） | — |

**格依据**：三 level 行主映 B2「正常授信/拒绝」——六区分是「未判定原因」清单，不含已判定
出口，故已判定值只能落六格外；执法证据 `verifier_test.go:409` 逐字
`the verdict is a deny but the raw record carries no overdue beyond the anchored threshold — rule 0 is the only deny path in this spec, so this deny is fabricated`
（触发条件在 `:408`）——我方把 deny 当**有规则的判定**执法，而非六格之一。
on_undefined_rule→「规则未覆盖」：C3 `:84` `no rule covers this case`、C6 `:66` `apply no rule`、
C12 `:91` `UNSETTLED`，词面与 B1「规则未覆盖」同型；B1 处置路线「明确未决路线」对 C6 `:66`
`record a semantic issue for the authorized business owner`，「补证」对 C6 `:65`
`request admissible asset valuation`——二路线逐别有对应物。**趋同成立于这二句硬约束
（不得默认成许可／不得默认成业务拒绝），不成立于状态个数。**

### 表 2-3 gh5 五状态 ↔ gh4 六区分（邢方两源自洽性核对）

| gh5 态 | gh4 §7 落点 | 判定 |
|---|---|---|
| Determined | 无格（在 B2「正常授信/拒绝」侧） | 分层一致，无分歧 |
| Uncertain | 「证据不足/冲突」的**不足半格** | 与 Conflict 被合写——分叉入 D5 |
| Unresolved | 「规则未覆盖」＋处置句「明确未决路线」 | 强对应 |
| Conflict | **gh4 无独立格**：六格中「规则未覆盖」是空集不是相撞，「证据不足/冲突」写的是**证据**冲突 | 入 D6 |
| Refused | **gh4 无格**：§7 的「拒绝」全是「发布前拒绝缺失必需规约／业务拒绝」义 | 入 D7 |

## 三、on_undefined_rule 非单射专项（逐案给档，不留中间地带）

**案一（C3 `:92` 的带间案，含 C5 `:55` 的恰在 30 界上案）**：档位＝**Unresolved，单射，不带
Conflict 分量**。
逐字依据：该发射点自述 `assets between the bands — an open specification question for the owner,
no default grant`——二带规则（`>30`、`<1`）对该输入**都不适用**，这是空集；C8 把冲突定义为
"equal precedence"，而 C5 诸 rule 的 precedence 是 0/1/2 **互异**，相撞条件在文本上不成立。
gh5 侧对 A1 gloss「无法判断」，gh4 侧对 B1「规则未覆盖」。**不给 Conflict 分量的理由**：若把
「两规则都不覆盖」读成冲突，则 Conflict 与 Unresolved 在 gh5 里失去可分性，违反 A1「建议至少
区分」本身。

**案二（C3 `:84`，six-month 窗内任意时长逾期案）**：档位＝**Unresolved 主（现行冻结文本下），
Conflict 仅在属主改判时接管**。逐字依据：该点自述（节引，`…` 处为 fmt 占位串，完整逐字见 C3
`:84`）`upper-band premise unsatisfied…no rule covers this case, no automatic deny (xingchao
198789782 §2)`，即 rule 1 前提失败、rule 0 不触发
（时长未过三月）→ 空集，与案一同型。**但**它派生的读法分歧有文本记录。C3 上方注释原文按 Go
注释换行（斜杠处＝换行，行号随附）：`e2e_test.go:49-51`「The judgment-spec-credit fixture
AGREES since the P3 correction landed / (unbounded veto "at any time", asserted by /
TestE3SpecSelfContradiction_WindowScope)」；`:60-63`「The old T0 stand-in checked
assetReportOK first, returning / on_missing_evidence when both a veto and a gap were
present — a / priority-order discrepancy between T0 and T1 (minor-ledger L23, / validated
by TestE3SpecSelfContradiction_C8Precedence).」——我方自己的测试名把这段历史称为**规格自相矛盾**。
若邢方/属主裁定「规则文本对 window scope 的两种读法」即 A1 的「规则冲突」，本案必须改映
Conflict（届时 `:92` 案仍是 Unresolved，非单射当场坐实）。这一档**不由我方定**：判据文本出自
邢方裁语 198789782 §2 与 08:181-183／10-walkthrough:145，入 D3。

**结构层结论（「非单射」的硬形式）**：非单射**成立于契约闭包**，**暂不成立于现夹具可达路径**。
成立面：C8 要求同序位冲突「必须由 resolution 段命名」，而 C7 把该段封闭为两键
（`"additionalProperties": false` ＋ 上述 `required` 两值）——一旦某规格真出现同序位规则，Conflict
**只能借 on_undefined_rule 说话**，该词于是必然同时代 Unresolved 与 Conflict 两态表言。
不成立面：现 credit 夹具 precedence 互异（C5），借用路径当前不可达，故**今日二发射点实测都是
空集型＝Unresolved**。推论：把 on_undefined_rule 直接对译「Conflict」或宣称「一对一」都不成立；
对 gh5 的正确表述是「一词承载多态，承载面由闭集决定，非由语义决定」。生产码侧另有事实：
词表命中面全在 e2e 测试层与夹具（`grep -rl grant-50000 agent-festival/go agent-festival/schema`
计得 8，同命令亦粘在 §六），`internal/judgment/check.go` 不含该
enum——**词表尚未落进生产判定码**，故改键成本只在 schema＋夹具（D2 的成本注）。

## 四、跨层同名与假朋友（映射表必须显式排除的读法）

| 对 | 两侧坐标 | 为什么是陷阱 |
|---|---|---|
| deny ↔ Refused（拒绝） | C3 `:66`／C4 `:41` ↔ A1 gloss「拒绝」 | deny 是有规则的**业务否定**，Refused 是**系统不答**；B1 句末「不应默认解释为业务拒绝」正是分层原文，C15 `:92` 钉的译名是「拒绝授信」 |
| ledger `unresolved` ↔ gh5 Unresolved ↔ gh4 结果未知 | C11 `:34`／C14 ↔ A1 ↔ B1 | 三向同名异层：C14 的 unresolved 是「已派发、结果未到」；gh5 Unresolved 是「无规则可判」＝B1 第三格。混读会把效果对账问题报成语义缺口 |
| ledger `conflict` ↔ gh5 Conflict | C11 `:33` ↔ A1 | 我方 conflict 是采纳机幂等冲突（同键改载荷），不是规则冲突 |
| 「补证／证据不足」 ↔ gh4 合写格 | C6 `:65` ↔ B1 | gh4 并「证据不足/冲突」为一格，我方两键分工只覆盖「不足」半格——按 gh4 口径核对会误判我方「缺一值」 |
| 附 3 归类表 ↔ 本表 | C15 ↔ 全表 | 附 3 是**同义词归类**（判分歧用），不是状态定义；本表是跨词表映射，二者不可互为证据 |

## 五、表尾分歧格清单（九格全部为裁项，我方不自裁）

- **D1 词数口径（四词／五词）**：裁对外呈现我方判定词表按几词报——76 号 #9 与 71 号 A6 写
  「四词」，代码实况五词（C1/C2/C15 附 4）。不能自裁：五词已由 67 附 4 定为「公共线程已立
  词汇」，而附 3 归类表只列四例（未含 grant-10000）且系**锁定后**追加、67 附 10 已把此类时序
  问题「裁决归夏」；改口径＝动预注册面＋外发随动件，两者都越出我方单方权限。
- **D2 resolution 是否开第三键（如 `on_rule_conflict`）**：裁 gh5 Conflict 要不要在我方契约有
  独立承载。不能自裁：C7 闭集＋C16「pending the business owner」把键集语义置于属主未决面，
  而 gh5 只点名不定义（A2），增键＝替邢方划定五态边界；成本注见 §三末段。
- **D3 案二（`:84`）的档位归属**：裁 window-scope 读法史是否构成 A1 的「规则冲突」。不能自裁：
  判据文本是邢方裁语 198789782 §2 与 08:181-183／10-walkthrough:145，定性权属原文作者；我方
  只能给现行冻结文本下的主档（§三案二）。
- **D4 gh5 五态逐态定义**：裁 Uncertain 与 Unresolved 的分界，以及「输入异常／分布外样本」落
  哪一态（A2：原文未定义、未指派）。不能自裁：定义权在 gh5 作者；附 3 体裁只是归类不是定义，
  缺定义时本表全部 gh5 侧格只能标「推定」。
- **D5 gh4「证据不足/冲突」合写 vs gh5 Uncertain/Conflict 分立**：裁两源取一还是并存。
  不能自裁：这是邢方两份文档之间的内部不一致，我方代选即替对方定验收口径。
- **D6 gh4 §7 无「规则冲突」独立格**：裁规则冲突并入「规则未覆盖」、并入「证据不足/冲突」，
  还是 §7 漏项（表 2-3 第四行）。不能自裁：三种读法都改 §7「必须区别」的强度，属原文作者。
- **D7 Refused 的表达层**：裁系统拒绝须在判定词层有值，还是我方现态（控制层 refusal-record，
  C13）即满足 §7／A1。不能自裁：若判「必须在判定词层」，则同时动 C2 的 T1 冻结模板（该文本
  正落在邢方裁语引用面）与 C7 闭集，两者都是共同财产。
- **D8 gh4 四格（未配置／明确配置为空／能力执行失败／结果未知）是否要求同层表达**：裁 B1 的
  「必须区别」是否等于「必须在同一词表」——我方现把四者分置发布层（C10 机械执法）与机器层
  （C11/C14）。不能自裁：判据是邢方 §7 的验收表述，我方无权解释对方的「必须」；若判「必须
  同词表」，表 2-2 末行四格全部转为我方欠项。
- **D9 ledger `unresolved`/`conflict` 的对外译名分工**：裁是否改译（如「效果待定」）以免与
  gh5 Unresolved/Conflict 混（见 §四 ledger 二例）。不能自裁：ledger 状态名已有外部消费者在场
  （76 号 #6：WangHaowen99 拟复用 internal/adopt；MR !1 评审面在途），改名影响对方代码。

## 六、计数实测与行数（先数后写：本文每处计数都是下列命令的输出）

```
$ git show --name-only --format='' HEAD
impl-2026-11-01/92-state-vocabulary-alignment.md
$ git diff --stat HEAD -- impl-2026-11-01/92-state-vocabulary-alignment.md; echo rc=$?
rc=0
$ wc -l /tmp/hub-issues/gh5-attachment-NGA.md /tmp/hub-issues/gh4-full.md /tmp/fs72/gh-issue-4.md
     352 /tmp/hub-issues/gh5-attachment-NGA.md
     424 /tmp/hub-issues/gh4-full.md
     425 /tmp/fs72/gh-issue-4.md
    1201 total
$ grep -c "Determined\|Uncertain\|Unresolved\|Conflict\|Refused" /tmp/hub-issues/gh5-attachment-NGA.md
1
$ grep -c 's1Verdict{kind:' agent-festival/go/internal/e2e/e2e_test.go
6
$ grep -o 'grant-50000|grant-10000|deny|on_missing_evidence|on_undefined_rule' agent-festival/go/internal/e2e/t1_llm_test.go | tr '|' '\n' | wc -l
       5
$ grep -in refus agent-festival/go/internal/e2e/t1_llm_test.go; echo rc=$?
rc=1
$ grep -rl grant-50000 agent-festival/go agent-festival/schema | wc -l
       8
$ wc -l impl-2026-11-01/92-state-vocabulary-alignment.md
    355 impl-2026-11-01/92-state-vocabulary-alignment.md
```

**计数口径逐条**：附件 **352 行**；gh4 快照件 **424 行**、fs72 副本 **425 行**（同一正文，
后者头部后多一空行——故 §7 句行号差一格，见 §〇）；五态名命中 **1 行**（A2 据此标「推定」）；
s1Judge 发射点 **6 行**（`on_undefined_rule` 占两案）；T1 enum 串按 `|` 切分得 **5 词**；
`refus` 在 T1 模板 **0 命中**（rc=1 是「无匹配」的正常返回，非报错）；词表命中文件 **8 个**
（全在 e2e 与夹具面，`internal/judgment/check.go` 不在内——§三成本注据此）。
**本文件行数见上列 `wc -l` 末行输出**（先跑命令、后写本节文字，回填不改行数）。

## 七、一句总判

三源在**二句硬约束**上逐字趋同（不得由模型补成许可 ↔ C12 `:91` `do NOT default to any grant`；
不得默认解释为业务拒绝 ↔ C12 `:93` `an automatic denial is forbidden` ＋ C6 `:65`
`not an automatic denial`），在**状态个数与分层轴**上不趋同：gh5 五态（无定义）／gh4 六格＋
六格外出口／我方五词（三值已判定、两值未判定路由、其中一词带闭包代言）。三套轴不同向，
映射是**多对多且跨层**；on_undefined_rule 的非单射由 C7 闭集构造决定，而非语义选择。
对外可呈报形态＝本表＋D1–D9 九格裁项，其余不裁。
