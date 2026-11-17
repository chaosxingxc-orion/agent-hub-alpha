# 致 chaosxingxc-orion：请评审 W06 年报智能体系统（架构 / 代码 / 效果）

> 留言人：wanshoulu · 日期：2026-11-17
> 对象：本仓 `agent-hub-alpha` 主理人 chaosxingxc-orion
> 请求：请评审本文件夹内的 W06 实现——**架构是否自洽、代码是否正确、效果是否成立**。凡指出不正确之处，我将及时修正并回传修订。

---

## 〇、先说清楚这份东西是什么

**W06** 是 `agent-hub-alpha` 架构（S1/S2 智能体框架）在**银行年报编制**场景上的一套可运行参考实现。它不是框架本体，而是架构里「用户面/运行面」的一个纵向切片：把「喂入一家银行的原始材料 → 产出可核验、可审计、可具名批准的年报」这条链路真正跑通。

几个词先就地解释，避免用项目内用法当定义：

| 名词 | 本文口径 |
| --- | --- |
| S1 / S2 | 本仓 `07-nga-architecture-proposal.md` 定义的智能体框架两层：S1 管任务与决策身份、准入与可回放，S2 管协作执行。 |
| W06 | 一组按年报编制职责切分的专家节点（取数、指标、资产质量、分析、成稿、主编）及其控制/核验/技能的合称，非模型名。 |
| openJiuwen | 开源的智能体运行底座；本实现通过 `skill-app-agent` 容器调用其 deep agent 作为 writer（成稿）执行体。 |
| 双检 | 数值核验（NumericChecker，N1–N7 规则）+ 引用核验（CitationChecker，C1–C2 规则）两道确定性检查。 |
| 工件（artifact） | 一次编制产出的结构化结果：metrics（指标）+ sections（章节正文）+ citations（引用）+ tables（数据表）+ charts（图表）。 |
| 批准链 | 具名审批人（approver）对工件指纹签名，篡改后指纹失配即拒绝。 |

**状态声明（实事求是）**：这是一套**能跑、跑出了成品**的参考实现，不是生产系统。下面凡涉及「效果」的数字，都标明测量口径与来源；凡未做到的，直接列在第五节「已知限制」，不藏。

---

## 一、请评审什么

### 1.1 架构（A 面）
- 四层结构与三权分立是否落到了本实现的真实边界上（详见 `reviews/w06-architecture-panorama.md`）。
- 控制面（准入/授权/预算/异常）与执行面（writer agent）的分离是否到位，有没有被实现细节悄悄绕过。
- Go 控制面的数理化不变量（幂等、执行许可、交付谓词）在本 Java 参考实现里是否被忠实表达，还是只留了名字。

### 1.2 代码（C 面）
- `source/w06-annual-report-agent/` 下 23 个权威 Java 类（checker 8 + contract 15）的正确性，尤其：
  - `NumericChecker.classifySubject` 的跨主体守护（防止把同业「招行净利润同比 0.85%」误判成本行指标）；
  - `ToolGateway` 的 `table_extract` / `table_analyze` 两个数据工具；
  - `CheckerHttpServer.extractDataTables` 的「表格与正文同源」修复；
  - `CheckerHttpServer` 的 long_form 任务书组装（往 writer 任务书里注入结构化表格）。
- 核验规则 N1–N7 / C1–C2 的实现是否与 `skills/annual-report/evals/` 的夹具一致；夹具是否能真正卡住错误（而不只是绿）。

### 1.3 效果（E 面）
- `artifacts/工商银行2025年度报告_agent系统生成.html`：27 章、27,003 字、81 条引用、50 张真实数据表、4 图表；双检 PASSED；具名批准 APR-0004（approver=Audit，指纹 `ec73959e…`）。
- 请判断：这份成品是否**真的**由 agent 系统生成并核验（而非人工代写后贴标），数据是否可溯源到落盘快照，批准链是否真的防篡改。

---

## 二、做了什么（机制名清单，不堆形容词）

| 机制 | 落点 | 可验证事实 |
| --- | --- | --- |
| 工具注册表 | `tools.json` | 6 个工具：`web_search` / `web_fetch` / `pdf_extract` / `snapshot_commit` + 新增 `table_extract` / `table_analyze` |
| 数据能力补强 | `TableExtractor` / `DataAnalysis` | 把 PDF 拍平的财报文本还原为结构化表；对 live 快照实测抽出 **50 张**真实表 |
| 同源约束 | `extractDataTables` | 表格按 `bank` 取与正文同一快照集；`bank=live` 才取 live 集 |
| 表格注入任务书 | `buildTaskBook` | 精选 22 张关键表注入 writer 任务书，使正文有数据支撑 |
| 跨主体守护 | `NumericChecker.classifySubject` | 判定窗口扩到 110 字符，识别同业参照上下文则跳过本行同比比对 |
| 长文稳定性兜底 | `AgentChatClient` / 续写轮 | 客户端超时 600s、缺节续写轮 2→4；writer 服务端 `AGENT_TIMEOUT=900` |
| 前端结构整合 | `frontend/workbench.html` | 删 Agent Hub 页签（功能并入资产库）、演示链降级为二级入口；jsdom 端到端测试 0 JS 异常 |
| 具名批准 | `/v1/approvals` | APPROVED / 残缺 400 / 篡改 409 三种行为实测 |

---

## 三、效果证据与测量口径

| 断言 | 口径 / 来源 | 证据 |
| --- | --- | --- |
| 50 张数据表 | 对 live 快照（109 页，工行 2025 年报）跑 `table_extract` 的返回 `table_count` | `artifacts/…html` 内 50 个 `class="data-table"` 节点 |
| 27,003 字 | 对工件 `sections[].body` 逐节 `len()` 求和，不含表格/引用 | 见 `gen_report_v2.py` 渲染输出 |
| 双检 PASSED | `/v1/pipeline/run` 返回 `checks.numeric.status` 与 `checks.citation.status` | 均为 `PASSED`，violations=0 |
| 批准防篡改 | 对批准工件改一个 metric 值再提交 `/v1/approvals` | 返回 409 拒绝 |
| 三银行无串数 | cmb/icbc/ccb 各自 brief pipeline 的 `artifact.tables` 信源指纹 | 三家各自独立，均不再混入他行表格 |

**注意**：writer（openJiuwen deep agent）单次长文生成本身**不稳定**——同一输入三次运行分别产出 26,355 字 / 1,374 字（回退）/ 27,003 字。最终交付取的是双检通过且章节完整的一次。这一点我主动披露，因为它直接关系到「效果是否可稳定复现」的评审结论。

---

## 四、设计取舍（每个决策都带代价）

1. **表格与正文同源**：修掉了跨银行串表的真 bug，代价是招行/建行快照只有 4 页纯文字节选、抽不出表格，报告从「偷来的 50 张工行表」变成 **0 张**——正确性优先，丰富度让位。
2. **任务书只注入 22 张表**（共抽出 50 张）：全塞会让 writer 输出截断，代价是正文引用的表格数少于附录展示数。
3. **writer 服务端加超时**（`AGENT_TIMEOUT=900`）而非继续压缩任务书：代价是单次生成可能更久，换来章节不再回退模板。
4. **前端只做结构整合、不重做视觉**：保留既有设计，代价是视觉未升级，换来零回归风险。

---

## 五、已知限制（诚实列出，请重点拍砖）

1. **writer 长文不稳定**：deep agent 生成质量随任务书长度波动，靠超时+续写兜底，尚无确定性保证。
2. **招行/建行 0 数据表**：快照只有节选，需另抓完整年报 PDF 才能有密集数据表——这是数据供给问题，不是系统故障。
3. **writer 侧技能显示 5 个、系统侧 9 个**：口径不同，主流程实测正常，但未深查差异是否影响长文质量。
4. **核验只查 metrics/sections，不查 tables**：`tables` 字段不进 N1–N7 / C1–C2，早期跨行串表之所以能潜伏正因如此。若评审认为表格也应纳入核验，这是要补的一块。
5. **部分表头靠规则推断**：抽取出的表格列标签在少数表上退化为「本期/上期/增减/增长率」的通用兜底，非逐表精确标注。
6. **无生产级持久化**：快照、任务库、审批记录均在容器文件系统内，未接 PostgreSQL 权威账本——与架构 v3 的账本目标仍有差距。

---

## 六、修订承诺

凡 chaosxingxc-orion 在评审中指出的问题——无论是架构偏离、代码缺陷还是效果夸大——我将在确认后**及时修正**，更新本文件夹内的对应文件，并附上修订说明与验证结果。欢迎直接在 Issue 或本目录新增评审文件，我会逐条回应。

---

## 附：目录导览

```
W06-wanshoulu-1117/
├── REVIEW-REQUEST.md          ← 本留言
├── README.md                  ← 工程说明与运行方式
├── source/
│   ├── w06-annual-report-agent/  Maven 工程（线上权威 Java 类已覆盖）
│   └── config/                    tools.json(6工具) / report-structure.json(28节)
├── skills/                      9 个技能（annual-report / orchestrator-skill，线上版）
├── frontend/workbench.html       整合版工作台（对话/资产库/报告库三页）
├── reviews/                      6 篇架构全景与实现审视文档
└── artifacts/                    报告成品 HTML + gen_report_v2.py 渲染脚本
```
