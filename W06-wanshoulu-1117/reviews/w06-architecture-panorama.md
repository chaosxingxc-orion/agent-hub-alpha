# Agent Hub Alpha · W06 视角架构全景文档

**版本**：v1.2 · 2026-11-17 09:00（v1.2 补读者背景）

> **一句话背景**：这是多团队协作项目 Issue #16 中 W06 工作包（年报报告 Agent 与核验）的理论支撑文档。项目整体在做"银行年报多智能体编制系统"：任务→生成执行图→AI 角色按图协作→确定性核验→受控交付；核心理念是"智能在规划期和节点内，确定性在调度与提交"。W06 负责其中的专业角色节点与核验器。不熟悉术语请先读 `ARCHITECTURE-v3.md` 开头的"读者背景"节。

**版本历史**：v1.1 · 2026-11-17 07:50（主 Agent 显式化）（v1.1：主 Agent 显式化——图的唯一生成者）
**范围**：Go 控制面与 Agent 系统的逻辑关系（数理化）· 图执行机制 · 主 Agent 知识体系与候选图生成 · 现状差距与整改
**用途**：D1 交付样例附图配套说明 · W06 集成对齐文档（对接 yougq 的 Contract-C / xierongqiag2015 的 ALK-05 / W01 材料规格）
**基线**：GitHub #16 全部 26 条评论（截至 2026-11-16 13:20 UTC）+ #19 S01-S12 + #20 E01-E12 + #27 AR-01—08 + 线上系统实测数据

---

## 一、系统总架构与三权分立

### 1.1 四层结构

```
┌─ 用户协作层（W02 + 主 Agent）───────────────────────────────┐
│  WorkSwarm 对话/编排台（人机界面）                          │
│  【主 Agent / 编排智能体】←— 图的唯一生成者                  │
│    · 澄清意图（对话）                                       │
│    · 消费能力知识（四类：能力卡/专业/流程/约束/案例）          │
│    · 生成候选图 + 选择记录 ←— 智能发生在这里                 │
│    · 主编团队/装配（bind agent/skill/MCP 给图节点）          │
└──────────────┬───────────────────────────────────────────┘
               │ ①命令（commandId 幂等）  ②只读投影（SSE）
┌─ 控制与事实层（W04 · Go + PostgreSQL）────────────────┐
│  命令服务 · 图采纳 · 执行许可 · 结果/审计              │
│  PostgreSQL 权威账本 · 交付谓词判定                     │
└──────────────┬───────────────────────────────────────┘
               │ ③采纳图下发 + dispatch（执行许可）
┌─ 运行推进层（W03 · Java · openJiuwen）────────────────┐
│  BlueprintScheduler：拓扑推进/并行/join/有限修订        │
│  故障恢复 · 对账/未知处理                               │
└──────────────┬───────────────────────────────────────┘
               │ ④携 effect token 经 Gateway 调专家
┌─ 专业执行层（W05/W06 · 专家 Agent）───────────────────┐
│  资料分析专家×N · 核验器 · 批准官 · MCP 工具面          │
│  （节点内：Agent Loop + Skill 五件套 + Rail + MCP）     │
└──────────────┴──────── ⑤observation 回流 Go ──────────┘

支撑组件（横跨）：Agent Bus Gateway（验 token/epoch/payload）
                 · RDC（AgentCard 发现，目录≠授权）
                 · Model Proxy（外呼唯一出口，模型调用=子 effect）
```

### 1.2 三权分立（11-14 定论）

| 权力 | 归属 | 语义 |
|---|---|---|
| **推进权** | Java BlueprintScheduler（W03） | 算"哪个节点该跑了"：拓扑序/并行/join/路径推进。Go 不做图遍历、不建 ready queue、不做 join——"Go 原子提交准入/dispatch 可以成立，并不自动等于 Go 调度" |
| **提交权** | Go/PostgreSQL（W04） | 算"什么算数"：正式结果唯一提交、命令幂等、图冻结、准入授权、交付谓词。**全系统唯一有权写正式状态的层** |
| **干活权** | 专家 Agent（W06） | 生产内容：节点内可用 LLM（agent loop），但零提交权——一切产出回流 Go 判定 |

### 1.3 层间数据流（编号对照上图）

0. **候选图生成**（主 Agent → 用户）：主 Agent 消费四类知识生成候选图+选择记录 → WorkSwarm 图卡展示 → 用户核对后发出 ADOPT_GRAPH 命令
1. **命令**：`POST /v1/commands {commandId, actor, type, payload}`，type ∈ {CREATE_TASK, ADOPT_GRAPH, CONFIRM, PAUSE, RESUME, REVISION}
2. **投影**：`GET /v1/projection?taskId&afterSeq`（SSE，重连补齐）——界面只读库投影，禁止从聊天记录/内存推导状态
3. **下发**：采纳图 T/g + dispatch{attemptId, lease, epoch}——Runtime 提 ready-proposal，Go 事务校验（依赖已提交∧未暂停∧版本匹配∧epoch）后才发 dispatch
4. **调用**：RemoteAgentCaller 经 Gateway，携 effect token（短时效+payload digest 绑定）+ 回调坐标（taskPushNotificationConfig）
5. **回流**：Agent 产出=候选 observation（非正式结果），POST /v1/runtime/observations → Go 校验落库才算记录

---

## 二、Go 控制面的数理化

### 2.1 核心不变量

```
(1) 提交权唯一性：∀ 状态转移 s: Commit(s) ⟹ 提交者 ∈ {Go}

(2) 提议与许可分离：Ready(n) 由 Java 计算（提议权）
                    Dispatch(n) 由 Go 签发（执行许可）
                    Ready(n) ↛ Execute(n)；需 Go 事务校验通过

(3) 落库即真相：∀ 事件 e ∈ {产出, 检查, 批准}:
    对用户可见(e) ⟹ e 已落 Go 账本
    （未落库即不存在；投影只读）

(4) checkpoint 边界：Runtime checkpoint 不用于宣布业务成功
```

### 2.2 幂等律（命令与结果两套）

```
命令层（用户面→控制面）：
  same(commandId, content)   → 返回原回执，副作用计数不增   [A03]
  same(commandId, ¬content)  → 409 Conflict，旧记录不变     [A04]

结果层（节点执行）：
  身份 o = (T, g, node_path, iteration_path)，attempt 与执行所有权代次分离
  same(o, input, content)    → 复用原正式结果（回执号 ≠ 输入摘要）
  same(o, input, ¬content)   → 冲突拒绝，不覆盖
  ¬∃result(o)               → ≠允许重发（UNKNOWN 显式化 → 人工入口）
```

### 2.3 执行许可与防旁路

```
effect token = 签名(短时效, payload_digest, epoch)
Gateway 放行条件：token 有效 ∧ epoch 匹配 ∧ payload digest 一致
∀ 调用 c: ¬valid_token(c) → Gateway 拒绝          [E06/E07]
每次 LLM 外呼 = 子 effect_id（Model Proxy 前置授权+按 effect 缓存响应）
旧 epoch worker 的提交 → 代次拒绝（先判重复再判代次——S09 实验教训）
```

### 2.4 交付谓词（W06 与 Go 的接口点）

```
Deliverable(T) ⟺ Numeric(T) = PASSED
             ∧ Citation(T) = PASSED
             ∧ UserConfirm(T) = CONFIRMED
             ∧ sha256(三方记录绑定的工件) 相同

语义：
- 谓词里没有"核验通过"这个裸事实，只有"绑定同一工件指纹的通过"
- 改工件内容 → 旧检查/批准全部失效 → 必须重查重批 [E03]
- W06 的 CheckResult 若不带 artifact_sha256（缺口 N3）→ 谓词数学上拼不上
```

### 2.5 PostgreSQL 权威账本

| 表 | 内容 |
|---|---|
| graph_revisions | 冻结的 T/g + 绑定规约版本/材料摘要/能力卡版本（AR-04） |
| node_attempts | attemptId / lease / epoch / dispatch |
| committed_results | o 唯一正式结果 · 内容寻址 |
| outbox | 事件发布（投影/通知源） |
| checks | W06 核验结论 + artifact_sha256 |
| approvals | 具名批准 + 同 digest 绑定 |
| audit_log | 审计（必要审计失败→传播，AR-03） |

---

## 三、图执行机制（Graph Execution）

### 3.1 图的生命周期

```
用户任务（自然语言，经 WorkSwarm 到达主 Agent）
  ↓ ①【主 Agent 动态生成图】LLM 规划——智能只发生在这一步
  │   主 Agent 执行四步纪律（召回→适用性→排序→选择，§4.3）
  │   消费：能力卡目录（agent/skill/MCP 注册表）+ 能力知识（ALK）
  │        + 流程方法（自己的 SKILL.md）+ 规范约束 + 案例经验
  │   产出：候选图{digest} + 完整选择记录
  ↓ ②静态校验（Go，确定性——主 Agent 的产出在此被机器审判）
  │   schema → 能力存在+版本 → 依赖有效 → 循环有界 → 预算
  │   不合法 → 拒绝+原因 → 返回主 Agent 修改（可多轮）
  ↓ ③用户确认（WorkSwarm 图卡展示候选图+选择记录，受信边界）
  ↓ ④冻结（Go 落 graph_revisions，此后只读——改即版本漂移即拒）
  │   【主 Agent 的职责到此为止——它生成图，但不采纳图、不执行图】
采纳图（adopted graph, T/g）——执行期唯一事实源
  ↓ ⑤机械执行（BlueprintScheduler 按拓扑序推进，不再有任何"再生成"）
```

**主 Agent 的职责边界**（#16 明确）：
- **做**：澄清意图 · 消费能力知识 · 生成/修改候选图及说明依据 · 多轮调整（确认前）
- **不做**：采纳图（用户确认+Go 冻结）· 推进图（BlueprintScheduler）· 提交正式结果（Go）· 批准（独立批准官）
- 11-13 原话："编排智能体负责澄清意图、消费能力知识、生成/修改候选及说明依据。检查通过及所需确认完成后，控制面采纳 T/g，运行面执行该版本。"
- 未配置规约/未覆盖/技术错误时，主 Agent 不得自行放行（§4.4）

**关键分界**：生成是智能的（一次性，主 Agent），执行是机械的（确定性，调度器）。#16 原话："知识消费必须影响图，不能先写死图再让模型补解释"——①必须真做（按能力知识规划），④之后图就死了，谁也不能改（包括模型）。同任务确认前可多轮调整；改已采纳拓扑 → 停在明确边界，以关联新任务+新采纳版本承接，不做隐式热改图（11-13 定论）。

**演示防伪指标**（#16 明确要求）：两个不同分析重点的任务，生成的图必须有可解释的差异——预录图伪装规划当场穿帮。

### 3.2 图推进 vs Agent 自由循环

| 维度 | DeepAgent 自由循环 | 图执行（#16 要求形态） |
|---|---|---|
| 推进者 | LLM 每轮决定调谁 | 调度器按拓扑序机械推进 |
| 下一步确定性 | 同输入可能走不同路径 | 图定了路径就定了 |
| 可恢复性 | 靠模型"看着办" | 重启后从已提交节点续跑，模型调用增量=0（S03） |
| 可审计 | 无法回答"为什么走这条分支" | 每节点派发/提交有记录，路径是声明出来的 |
| 并行 | 模型想起来才并行 | 声明并行分支，调度器保证同时派发+join 汇合 |
| 防失控 | maxIterations 兜底 | DAG 无环 → 结构性不可能死循环 |

### 3.3 调度循环（BlueprintScheduler，yougq 已交付 W03）

```
loop:
  读 Go snapshot + dispatch(afterSeq)
  ready(n) ⟺ ∀p ∈ pred(n): Committed(p)
  并行组 analysis：finance ∥ risk 同时派发（输入互斥隔离）
  join 节点 writer：全部前驱提交才放行
  修订环：REVISION_DIRECTIVE → writer 重跑 ×1（上限预声明）
  提交观察 POST /v1/runtime/observations{eventId, o, ...}
  Go ack 后才算记录（checkpoint ≠ 业务成功）
```

### 3.4 故障恢复对照表（#19 S03-S09）

| 用例 | 场景 | 系统行为 |
|---|---|---|
| S03 | 完成节点后 kill -9 Java | 新 worker 读账本续跑；已完成节点结果原样复用；模型调用增量=0 |
| S04 | 纯计算后、提交前崩溃 | 租约过期被接管重算 1 次；正式结果至多 1 条 |
| S05 | 双 worker 并发 | 只有一方能提交（epoch 栅栏）；旧 worker 提交被代次拒绝；旧 owner 已提交时重复提交被识别为复用 |
| S06 | 发送意图已提交、请求未发出 | 接管后按 b 查询，查无此单→同 b 新 attempt，不换身份 |
| S07 | 远端已生效、回执丢失 | 按 b 查到已受理→补记回执，不新建申请 |
| S08 | 映射未落地 | 无法证明未发生→UNKNOWN+明确人工入口，不静默新建 |
| S09 | 本地提交成功、回复丢失 | 重试取原 R(o)；**先判重复再判代次**（实验教训：顺序错会误判重算） |

---

## 四、主 Agent 知识体系（能力配置全景）

**总纲**：主 Agent（编排智能体）是图的唯一生成者。它生成高质量图的能力上限 = 配给它的知识质量。本节定义这套知识体系的完整配置——四类知识（能力/专业/流程/案例）以 skill 与结构化卡片形态挂载给主 Agent，经渐进披露消费：目录常驻、正文触发加载、脚本执行时才调。

### 4.1 知识分类（四分法 × ALK 机器可读形式）

| 知识类型 | 形态 | 回答什么 | 现状 |
|---|---|---|---|
| **能力知识**（注册的 agent/skill/MCP 本身） | 目录索引+能力卡 | "有什么能力？" | ✅ 三注册表齐（7 skill/13 agent/5 MCP server） |
| **专业知识** | domain-brief.md | "业务要点是什么？（澄清用）" | ⚠️ 有 banking-domain 但未绑主编 |
| **流程方法** | 主编 SKILL.md | "怎么规划：四步纪律" | ❌ 完全缺位 |
| **规范约束** | constraints.md + W01 规约 | "什么会被拒？为什么？" | ⚠️ 散落，无主编视角汇总 |
| **案例经验** | evals（好图/坏图/差异图） | "参照长什么样？" | ❌ 完全空白 |

**关键区分**：注册表回答"存在什么"，不回答"何时用哪个"。适用性判断是规则不是描述——"finance-analyst 描述=经营分析"是注册信息；"它只支持集团口径、母公司口径任务必须排除"是选择规则，须进能力卡结构化字段（机器可校验），且每次选择留痕（可审计）。这正是 ALK（能力知识）的定义：**注册表之外的结构化适用性数据**，与注册表解耦——Bus/RDC 管"能不能调到"，能力卡管"该不该用"（11-13 定论："不要求先建设通用知识图谱或向量平台"）。

### 4.2 能力卡四元组（目录索引的深加工）

```
① id + 描述        有什么（具体动词+触发语境，pushy 防漏触发）
② applies_to       何时该用（适用场景，结构化）
③ constraints      何时不该用（口径限制/前置条件；如"仅集团合并口径"）
④ 版本 @v          消除版本漂移（AR-04：采纳图绑定能力版本）

缺 ②③ → 主 agent 只能文本匹配描述 → "高分但不适用"误选无法结构性排除
卡片质量 = 选图质量的上限
```

### 4.3 候选图生成四步纪律（ALK-05 验收标准）

```
①召回     按任务匹配目录（description + applies_to）
          全 unknown → 显式拒绝，不得瞎选          [验收：全 unknown 不算通过]
②适用性   读 constraints；口径不符 → 排除
          高分但不适用 → 仍被排除，留排除理由      [验收：不适用候选被排除]
③排序     按适用度+成本+版本稳定性
          不保证全空间最优，保证可解释选用
④选择     节点=能力卡@版本，边=数据依赖，预算+有限循环声明
          输出候选图{digest} + 完整选择记录
          两种任务重点 → 有依据的图差异           [验收：选择/组合差异可解释]

每步留结构化记录（召回/排除/排序/选择）——选择记录进审计
```

### 4.4 规划期失败处理（11-13 原则）

```
未配置规约      → 按管理面配置进入澄清
规约未覆盖     → 显式等待/降级（查 W01 降级表）
技术校验错误   → 拒绝 + 原因返回
—— 三种情况分别显式处理，模型不得自行放行 ——
模板可作候选起点，但不能把固定模板选择冒充无限自由编排
```

---

## 五、W06 专家执行层规格

### 5.1 专家节点全景

| 节点 | 职责 | 输入 | 输出 |
|---|---|---|---|
| material-agent | 快照装载（唯一入口） | read_page{bank,page} | digest/page/excerpt 三元组；缺料→MATERIAL_MISSING |
| finance-analyst | 经营分析（营收/净利/息差/成本） | 指标子集（隔离注入） | 带 [C-n] 引用草稿；派生值调 MCP 不心算 |
| risk-analyst | 资产质量（不良/拨备） | 指标子集（与经营互斥） | 同上 |
| capital-analyst | 资本充足 | CET1 指标 | 同上 |
| writer | 结构化成稿 | 双分析草稿+report-template | AnnualReportArtifacts（metrics 原样转抄） |
| numeric-checker | 数值核验 N1-N7 | 工件+快照 | CheckResult（须带 artifact_sha256） |
| citation-checker | 引用核验+支持度三级 | 工件+信源注册表 | CheckResult（同上） |
| approver | 具名批准（独立于主编） | 核验通过+工件 digest | 批准记录（绑定 digest） |

### 5.2 节点内装配（bind 即真实装配）

```
Agent Loop（LLM 智能在节点内）
+ Skill 五件套：SKILL.md（pushy description）· scripts/（确定性核验脚本）
              · references/×6 · assets/（模板/注册表/schema）· evals/
+ Rail：GatekeepRail（fail-closed：PASSED→deliverable；FAILED→修订×1；
        再失败→waiting_human）
+ MCP 工具：JSON-RPC 2.0（initialize→tools/list→tools/call）
        bind MCP → 全限定名工具（Server:tool）
渐进披露：元数据（常驻）→ 正文（触发时）→ 链接文件（按需）
```

### 5.3 执行层铁律

- 值照抄原文（含千分位），不心算——派生值一律 `基于 [C-k] 计算：<算式>`
- 每论断恰好一条 [C-n] 引用；引用三元组必须页内字面命中
- 百分点 vs 百分比严格区分（N7）
- fail-closed：检查失败阻止验收，语义存疑转人工
- **零提交权**：一切产出回流 Go 判定

---

## 六、现状差距与整改清单

### 6.1 W06 缺口（对照 #16 验收）

| # | 缺口 | 性质 | 工作量 | 期限 |
|---|---|---|---|---|
| N3 | CheckResult 无 artifact_sha256 | E03 直接命中（谓词拼不上） | 0.3 天 | D1 |
| N4 | 无百分点 vs 百分比检测（N7） | W01 反例集明列项 | 0.2 天 | D1 |
| N2 | 核验器未消费 W01 真实材料 | 自造数据反例仍在 | 0.5 天 | D1 |
| N5 | 无执行身份 o/a/b | yougq 契约已就位，直接消费 | 0.5 天 | D1-D2 |
| N6 | 专家未挂 W05 RDC/Gateway | 集成 | 1-2 天联调 | D2 |
| N7 | ReportSection 缺类型字段 | W01 模板要求 | 0.3 天 | D2 |
| N8 | LLM key 未接入 | 外部依赖（W07/owner） | 上报 | D2 |

### 6.2 Agent/Skill 配置矛盾（R 整改，1.8 人日）

| # | 问题 | 整改 |
|---|---|---|
| R1 | 分派双真相源（team[] vs 骨架 vs 六工位互相矛盾） | 注册表启动时从骨架派生 |
| R2 | writer 缺位（disclosure 三合一瓶颈） | WriterAgent 上注册表 |
| R3 | skill 四处错绑漏绑 | 绑定改消费关系驱动 |
| R4 | kind 失真 | 补 TOOL/WRITER/RENDERER |
| R5 | approver 在主编 team（批准链不独立，违 E02 精神） | 移出 team |
| R6 | 骨架无 metric_keys（输入注入无配置依据） | 每节声明消费指标 |
| R7 | 主技能/reference 概念混淆 | 注册表分两级 |

**时序约束**：R1-R3 必须在角色接 W05 RDC 之前完成——否则错绑固化进 A2A 能力卡。

### 6.3 主 Agent 知识包（orchestrator-skill，待建）

```
orchestrator-skill/
├── SKILL.md                  流程方法：四步纪律+多轮调整规则+失败处理
├── references/
│   ├── planning-method.md    召回→适用性→排序→选择，每步留记录
│   ├── constraints.md        Go 会拒什么+口径适用性+生成前自检清单
│   └── domain-brief.md      年报业务要点（主编澄清用）
├── knowledge/
│   └── alk05-annual-report.json   能力卡四元组×（5 角色+5 MCP）
│       （兼容 xierongqiag2015 已交付样例格式）
└── evals/
    ├── case-full-report.json      完整年报→完整图+选择记录
    ├── case-quality-focus.json    资产质量重点→图缺经营节点（有依据差异）
    ├── case-rejected.json        母公司口径→高分候选被排除记录
    └── evals.json                官方格式
```

对应 AR-06/ALK-05 的 W06 侧交付物。

---

## 七、验收映射速查

| 验收项 | 本文对应 |
|---|---|
| A03/A04 命令幂等 | §2.2 命令层 |
| A05 恢复复用 | §3.4 S03-S05 |
| A07-A09 报告质量 | §5.2/5.3 + N3/N4 |
| AR-02/03 提交条件/失败传播 | §2.2 结果层 + §2.5 audit_log |
| AR-04 版本绑定 | §2.5 graph_revisions + §4.2 ④ |
| E02 伪造批准 | §2.1 命令服务（服务端验 actor）+ R5 |
| E03 改正文旧检查放行 | §2.4 交付谓词 + N3 |
| E06/E07 旁路 | §2.3 effect token |
| E08 四态降级 | §4.4 规划期失败处理 |
| E09 发送失败不伪装成功 | §2.2 AR-03 |
| S03-S09 故障用例 | §3.4 对照表 |
| ALK-05 选择差异 | §4.3 + §6.3 evals 三例 |
| ALK-06 能力知识进规划 | §4.2 四元组 |

---

## 附：三张架构图索引

1. **上半部**：用户协作层 → Go 控制/事实层（命令协议字段 · 四服务规则链 · PG 六表 · 交付谓词）
2. **下半部**：Java Runtime → 专家执行层（调度循环就绪谓词 · S03-S09 故障对照 · 专家节点内部 · Gateway/RDC/ModelProxy）
3. **主 Agent 知识与候选图生成**（目录索引四元组 · 四类知识 · 四步纪律 · 双重门禁→冻结）

三图连续接力，可直接作 D1 交付样例架构附图。
