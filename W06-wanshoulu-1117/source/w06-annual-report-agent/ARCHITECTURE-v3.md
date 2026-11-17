# W06 年报报告 Agent 与核验 · 整体架构与技术方案

**版本**：v3.2（可读性修订版）· 2026-11-17 09:00——v3.1 基础上补充读者背景与名词速查（面向未参与前期讨论的外部读者：评审方/联调方）
**前版**：v2（compose 设计期）——本版基于架构全景文档 v1.1 全面重写，对齐 #16 最新态势（W03/W04 已由 yougq 交付、W05 已 D1、W01 已签署、仅剩 W06/W07）
**文档关系**：本方案是唯一顶层文档（图文合一）；`w06-architecture-panorama.md` 为理论附档；架构图存于 `docs/diagrams/`（fig1 总架构 + fig2-4 分层细节图）
**读者**：W06 实现/验收方 · yougq（Contract-C 对齐）· xierongqiag2015（RDC/ALK-05 对齐）· WangHaowen99（W01 材料消费）

---

## 读者背景（第一次接触本项目的读者请先读这段）

**这个项目是什么**：GitHub agent-hub-alpha 仓库 Issue #16 的多团队协作 demo——用七个工作包（W01-W07）构建一个"银行年报多智能体编制系统"：用户用自然语言下达编制任务 → 系统自动规划一张执行图 → 多个 AI 专业角色按图协作 → 确定性核验 → 受控交付。全程可审计、可故障恢复、可核验。

**为什么架构长这样**：多智能体系统的天然风险是"模型不可控"——重复执行、篡改后放行、崩溃后重算、绕过权限。本项目用五轮对抗审查（#16 讨论区）收敛出一套控制架构：**执行流程先被规划成一张有向无环图（谁先谁后、谁依赖谁），图的推进由确定性代码（调度器）而非大模型驱动；系统唯一的"正式状态账本"由 Go+PostgreSQL 保管；大模型的智能只出现在两个受控位置——规划期生成候选图、节点内部干活**。这解释了下文的"三权分立"与各层职责。

**关键名词速查**：
- **候选图/采纳图**：大模型按任务生成的执行计划（候选图）经机器校验+用户确认后冻结为采纳图，此后不可改（防止运行中被模型偷改计划）。
- **A2A**：Agent-to-Agent 通信协议，角色之间不直连，一律经网关。
- **RDC/Gateway**：W05 提供的"角色注册中心/通信网关"——角色先注册（登记能力卡片），调用时经网关验权放行。
- **MCP**：Model Context Protocol，一种让 AI 调用外部工具的标准协议（JSON-RPC）。W06 的核验能力同时以 MCP 工具形式暴露。
- **Skill**：给 AI 角色配的"技能包"文件夹（说明书+脚本+参考资料），遵循 Anthropic Agent Skills 开放标准，按需渐进加载以节省上下文。
- **effect token**：每次调用前向 Go 申请的一次性授权凭证（绑定调用内容），防止绕过权限的"旁路调用"。
- **S03-S09/E01-E12/A01-A15**：验收用例编号（故障恢复类/对抗攻击类/功能验收类），定义在 #19/#20 两个 Issue。
- **D1/D2**：集成里程碑——D1（11-19）契约冻结、D2（11-24）完整链路联调。

---

## 0. 定位重述（一句话）

**W06 = 图上的专家节点 + 核验器**：为主 Agent 生成的年报任务图提供六个专业执行节点（资料/经营/资产质量/资本/写作/核验批准），节点内以 Agent Loop + Skill + MCP 干活，产出经确定性核验、绑定工件指纹后回流 Go 控制面参与交付谓词判定。

**W06 不是**：自研编排器（推进权归 W03 BlueprintScheduler）、自建账本（提交权归 W04 Go/PG）、自建前端主壳（归 W02 WorkSwarm）。

---

## 1. 目标架构（W06 在全景中的位置）

### 1.1 系统四层与 W06 的接缝

**Figure 1 · 完整架构总图**（`docs/diagrams/fig1-complete-architecture-v3.svg`）：单图呈现四层全组件——主 Agent 知识包与四步纪律、Go 四服务与 PG 账本、调度循环与故障恢复、W06 六节点流水线与三接口锚点、五个编号数据流、支撑组件三件。下文 ASCII 为同构速览：

```
┌─ 用户协作层（W02）─────────────────────────────────────────────┐
│  WorkSwarm · 主Agent（图唯一生成者：四类知识→四步纪律→候选图）    │
└──────────────①命令（幂等）──────────②只读投影────────────────┘
┌─ 控制与事实层（W04，yougq 已交付）──────────────────────────────┐
│  Go：命令幂等·图采纳·执行许可（lease/epoch/effect token）        │
│      结果/审计（o=(T,g,path,iter) 唯一）· 交付谓词              │
│  PG：graph_revisions · node_attempts · committed_results        │
│      outbox · checks · approvals · audit_log                   │
└──────────────③采纳图+dispatch─────────────────────────────────┘
┌─ 运行推进层（W03，yougq 已交付）────────────────────────────────┐
│  Java BlueprintScheduler：ready(n)⟺∀p∈pred(n):Committed(p)     │
│  并行/join/有限修订 · kill 续跑（S03：模型调用增量=0）           │
└──────────────④effect token 经 Gateway─────────────────────────┘
┌─ 专业执行层（W05 接入 + W06 本体）───────────⑤observation 回流─┐
│  【W06 职责区】                                                 │
│  material → metrics → finance∥risk∥capital → writer            │
│       → numeric-checker + citation-checker → approver           │
│  节点内：Agent Loop + Skill 五件套 + Rail + MCP 工具             │
│  零提交权：一切产出回流 Go 判定                                  │
└────────────────────────────────────────────────────────────────┘
支撑件（W05，xierongqiag2015 已 D1）：Gateway（验 token）· RDC（AgentCard）· Model Proxy（外呼唯一出口）
```

### 1.2 W06 节点图（年报任务的标准拓扑）

```
material-agent ──→ metrics-agent ──┬─→ finance-analyst ──┐
（快照装载，      （八指标抽取，    ├─→ risk-analyst ─────┤
 唯一入口）        三元组）         └─→ capital-analyst ───┤
                                                        ↓ join
                                                    writer-agent
                                                  （AnnualReportArtifacts）
                                                        ↓
                                    ┌ numeric-checker（N1-N7）
                                    └ citation-checker（支持度三级）
                                                        ↓ 双 PASSED
                                                  approver（具名批准，
                                                   绑 digest，独立于主编）
```

- 并行组 `analysis`：三分析师输入互斥隔离（经营四指标 ∥ 资产质量三指标 ∥ 资本一指标），越界引用被 citation 核验拦截
- join：writer 等全部前驱草稿
- 修订环：REVISION_DIRECTIVE → writer 重跑 ×1（与 GatekeepRail maxRounds=1 一致，预声明）

### 1.3 与三方的接口契约

| 接口方 | 契约 | W06 义务 |
|---|---|---|
| **yougq（W03/W04）** | Contract-C：`o=(T,g,node_path,iteration_path)`；同 o 同输入同内容→复用；改内容→冲突；无结果≠重发 | 节点执行携带身份；核验结论进 checks 表须带 `artifact_sha256`；对账语义对齐 |
| **xierongqiag2015（W05）** | RDC 真实 AgentCard 注册；Gateway effect token 验证；c↔request/task/context 映射 | 六节点以独立 Runtime 进程注册 RDC；调用验证 token；能力卡挂 inputSchema/outputSchema@版本 |
| **WangHaowen99（W01）** | 冻结材料（快照 SHA 52c49dcd…+8 指标参考答案+10 反例）；PR #21 `ae17ee3` | 消费真实材料核验；10 反例全部拦截（W01 交付卡明示的 W06 验收题） |

---

## 2. 技术方案分项

### 2.1 执行形态：从"DeepAgent root 自编排"到"专家节点"

**Figure 3/4 · 分层细节图**（`docs/diagrams/fig2/fig3`，会话内已渲染）：上半部（用户+Go 层字段级）与下半部（Runtime 调度循环+S03-S09+专家层内部）。

**决策**：v2 的 OrchestratorAgentFactory（root DeepAgent 指挥五角色）**退役为节点工厂**——每个角色独立进程、独立 A2A 暴露、被 BlueprintScheduler 调度。

**为什么改**：v2 形态里"下一步调谁"由大模型（root 编排）临场决定，这带来三个问题——同输入可能走出不同路径（不可审计）、中断后无法精确恢复（模型"看着办"）、可能无限循环（靠上限参数兜底）。v3 把推进权交给确定性调度器后，这三个问题结构性消失；而角色内部的智能干活不受影响。理由：#16 三权分立（推进权不归 W06）；yougq 已交付调度器；W06 的不可替代价值在节点内部。

**保留复用**（全部资产不废弃）：

| v2 资产 | v3 去向 |
|---|---|
| 5×AgentFactory + Properties | 独立节点进程入口（compose 已有雏形） |
| GatekeepRail | 挂 writer→checker 链尾（节点内门禁，与 Go 交付谓词双保险） |
| NumericChecker/CitationChecker | 独立 checker 进程（已是）+ MCP 工具面 + skill 脚本（三形态同源） |
| skills/ 五件套 | 节点知识层（bind 即装配） |
| 线上工作台 | 注册/装配/MCP/运行轨迹的演示与治理面 |

### 2.2 节点内三层结构（每个专家的自治单元）

```
┌─ Agent Loop（LLM，智能在节点内）────────────────┐
│  openJiuwen ReActAgent/DeepAgent 工厂           │
├─ 知识层：Skill 五件套（渐进披露）────────────────┤
│  SKILL.md 方法论 · scripts/ 确定性脚本            │
│  references/×6 · assets/ 机器可读 · evals/        │
├─ 能力层：Rail + MCP ────────────────────────────┤
│  GatekeepRail（fail-closed 状态机）              │
│  MCP 5 server 12 工具（JSON-RPC 2.0 全限定名）    │
└─────────────────────────────────────────────────┘
纪律：值照抄不心算 · 每论断一条 [C-n] · 派生值带算式
      · 百分点/百分比区分（N7）· 零提交权
```

### 2.3 核验器（W06 核心，三形态同源）

**为什么核验器是 W06 的核心**：整个系统的可信度取决于"报告里的数字和引用是真的"这一条。核验器不用大模型，用确定性代码复算每个数字（是否真在年报原文页出现、口径换算是否正确、引用摘录是否逐字存在），是"检查失败阻止验收"原则的执行者。

**为什么有三种形态**：同一套规则要在三种场景生效——独立服务进程（供 Go 控制面直接调）、Python 脚本（供角色节点在技能包内自主执行）、MCP 工具（供跨节点协议调用）。三份拷贝必须严格同步，否则出现"服务版拦了、脚本版放过"的漏洞。

| 形态 | 用途 | 消费方 |
|---|---|---|
| Java checker 进程（18083） | 服务版门禁（现有） | Go checks 表 / pipeline |
| skill scripts（verify_metrics.py/check_citation.py） | agent 节点内自主执行 | 各专家节点 |
| MCP w06-checker/w06-compute | 跨节点协议调用 | writer 自检 / 修订前检查 |

**规则集**：N1 完整性 · N2 单位白名单 · N3 页内字面溯源 · N4 三口径复算 · N5 同比复算 · N6 同比方向 · **N7 百分点口径（新增，W01 反例）**；L1 引用三元组 + 信源注册表 + 支持度三级（SUPPORTED/PARTIAL→人工/UNSUPPORTED）。

**v3 新增（N3 缺口修复）**：
```
CheckResult += artifact_sha256     # 核验结论绑定工件指纹
交付侧语义：sha256(verdict.artifact) ≠ sha256(当前工件) → verdict 无效
E03 反例测试：改正文 → 旧 PASSED 失效 → 旧批准连锁失效
```

### 2.4 MCP 工具面（已上线，v3 补版本字段）

5 server（snapshot/checker/compute/artifacts/approval）12 工具，JSON-RPC 2.0 完整协议（initialize→tools/list→tools/call），bind 即装配全限定名工具。v3 补：`version` 字段入能力卡（AR-04 版本绑定）；w06-charts（图表）待 chart-agent 有实现基础后注册。

### 2.5 主 Agent 知识包（orchestrator-skill，新建）

**Figure 2 · 主 Agent 知识与候选图生成**（`docs/diagrams/fig4-orchestrator-knowledge.svg`，会话内已渲染）：目录索引四元组→四类知识→四步纪律→双重门禁→冻结 T/g。

配给主 Agent（图生成者）的四类知识，实体化为技能包：

```
orchestrator-skill/
├── SKILL.md                  # 流程方法：四步纪律（召回→适用性→排序→选择，
│                             #   每步留结构化记录）+ 多轮调整 + 失败三态处理
├── references/
│   ├── planning-method.md    # "模板可作候选起点，不得冒充自由编排"（11-13）
│   ├── constraints.md        # Go 会拒什么（未知能力/无界循环/缺字段/超预算）
│   │                         #   + 口径适用性 + 生成前自检清单
│   └── domain-brief.md      # 年报业务要点（澄清用）
├── knowledge/
│   └── alk05-annual-report.json   # 能力卡四元组×（6 角色+5 MCP）：
│       # ①id+描述 ②applies_to ③constraints（如"仅集团口径，母公司任务排除"）
│       # ④版本@v —— 兼容 xierongqiag2015 样例格式
└── evals/
    ├── case-full-report.json      # 完整年报 → 完整图 + 选择记录
    ├── case-quality-focus.json    # 资产质量重点 → 图缺经营节点（有依据差异）
    ├── case-rejected.json        # 母公司口径 → 高分候选被排除
    └── evals.json                # 官方格式（prompt+expected_output）
```

ALK-05 验收三条全对应：选择差异可解释 / 高分不适用被排除 / 全 unknown 显式拒绝。

### 2.6 注册与装配治理（线上工作台，已上线部分+整改）

**已上线**：三注册表（agent 13/skill 7/MCP 5）· 装配引擎（bind skill=指令注入+工具挂载；注册主编=A2A 通道；bind MCP=全限定名）· 装配状态查询 · 运行轨迹面板（9 节点/thinking/工具调用/token/成本）。

**v3 整改**（对照审查报告 R1-R7，须在接 RDC 前完成）：

| # | 整改 | 说明 |
|---|---|---|
| R1 | 注册表从骨架派生 | 消灭分派双真相源（team[]/骨架/六工位三处矛盾） |
| R2 | 补 writer-agent | Java WriterAgent 上注册表；disclosure-officer 回归纯合规 |
| R3 | skill 绑定改消费关系驱动 | 修四处错绑（骨架/分工→主编+writer；citation-style→三分析师；metric-dictionary→risk；source-library→citation-checker） |
| R4 | kind 补全 | TOOL（material）/WRITER（writer）/RENDERER（chart） |
| R5 | approver 移出主编 team | 批准链独立（E02 精神） |
| R6 | 骨架加 metric_keys | 每节声明消费指标（输入注入配置依据+一致性校验） |
| R7 | 注册表分两级 | 主技能 vs reference，绑定语义不混淆 |

### 2.7 部署与运行

```
deploy/compose.yaml（v2 已有，v3 扩充）
├── 六节点独立进程（w06-net 内部网络，互不直通，只经 Gateway）
│     密钥经 secrets 挂载不进镜像/日志；内存限额 512M/节点
├── checker 进程（确定性核验，无 LLM 依赖）
├── MCP 面（随 checker 暴露 /mcp/*）
├── skills 卷（五件套只读挂载全部节点）
└── external: control-net（接 W04 Go / W03 Runtime / W05 Gateway+RDC）
      W01 快照卷 annual-report-snapshot 只读挂载 material-agent
```

LLM 接入：Model Proxy 为唯一外呼出口（每次调用=子 effect，前置授权+按 effect 缓存）；W06 节点不自持模型凭据。

---

## 3. 数据契约（W06 对外三接口的精确字段）

### 3.1 给 Go 的 checks 记录（N3 修复后）

```json
{
  "checker_id": "numeric-checker", "checker_version": "0.3.0",
  "task_id": "TASK-2025-AR-001",
  "artifact_sha256": "sha256:…",        // ← 核验结论绑定工件指纹
  "status": "PASSED|FAILED|WAITING_HUMAN",
  "violations": [{"rule":"N7","target":"NPL_RATIO","detail":"…","severity":"BLOCKING"}],
  "checked_at": "…", "execution_o": "T/g/path/iter"   // ← 执行身份（N5）
}
```

### 3.2 给 W03 的节点执行声明（消费 Contract-C）

```
o = (T, g, node_path, iteration_path)   // yougq 共享类型
attempt 与执行所有权代次分离
同 o 重交：同输入同内容→复用；改内容→冲突（不覆盖）
UNKNOWN：无结果≠允许重发 → 显式等待人工
```

### 3.3 给 W02/Builder 的能力卡（ALK 四元组）

```json
{
  "id": "risk-analyst", "version": "1.0.0",
  "description": "资产质量分析：不良/拨备口径，产出带引用草稿",
  "applies_to": {"scenario": "年报", "sections": ["贷款质量"], "scope": "bank-analysis"},
  "constraints": {"scope_exclusive": "资产质量三指标", "caliber": "集团合并口径"},
  "inputs": {"metrics": ["NPL_BALANCE","NPL_RATIO","PROVISION_COVERAGE"]},
  "outputs": {"draft": "sections[S-2] with [C-n] citations"}
}
```

---

## 4. 验证方案

### 4.1 反例回归（消费 W01 材料，替换自造数据）

| 层 | 用例 | 期望 |
|---|---|---|
| 单元 | 自有 16 反例 + W01 10 反例（PR #21） | 全部 BLOCKING 拦截 |
| E03 | 检查后改正文 | 旧 verdict+批准失效 |
| E09 | 发送失败 | 不伪装成功，状态传播 |
| N7 | 不良率降 0.94 个百分点写成"0.94%" | BLOCKING |
| S03 | kill 节点进程后恢复 | 已提交结果复用，模型调用增量=0 |
| ALK-05 | 两任务重点差异图 / 高分不适用排除 / 全 unknown 拒绝 | 三例通过 |

### 4.2 声明层级（诚实口径）

当前全部为**作者测试**层；D2 联调后按 #16 四层分级如实升级（组件验证→集成验证→系统特性），不越级声明。

---

## 5. 里程碑

| 阶段 | 内容 | 期限 |
|---|---|---|
| **M1（D1 前）** | N3 digest 绑定 + N4 百分点 + N2 换 W01 材料+10 反例拦截记录 + R1-R3 注册表整改 | 11-19 20:00 |
| **M2（D1）** | 交付卡回填（图/身份/事件契约样例）+ orchestrator-skill 包 + 架构三图 | 11-19 |
| **M3（D2 前）** | N5 执行身份（消费 Contract-C）+ N6 六节点挂 RDC/Gateway（联调 xierongqiag2015）+ N7 section 类型 | 11-24 20:00 |
| **M4（D2）** | 完整纵向链路：WorkSwarm→Go→Runtime→六节点→核验→批准→交付 | 11-24 |
| 阻塞上报 | N8 LLM key（W07/owner 提供） | M3 前需要 |

---

## 6. 风险与对策

| 风险 | 对策 |
|---|---|
| R1-R3 未做就接 RDC → 错绑固化进能力卡 | 严格时序：M1 完成整改再联调 |
| W05 环境交接延迟（rt77789→xierongqiag2015，11-18 截止） | M1 期间用 compose 替身注册中心并行开发（yougq 的 Contract-C 测试宿主可先行） |
| 真实 368 页 PDF 解析鲁棒性 | M1 的 N2 先跑通再评估 findNumberAfter 加固需求 |
| 三方契约漂移 | 每接口以对方已交付物为冻结基准（Contract-C 08e2e40 / ALK-05 820e1d4 / W01 ae17ee3），变更走 #16 评论留痕 |

---

## 附 A：文档体系

| 文档 | 角色 |
|---|---|
| **本文（v3.0 技术方案）** | 顶层：定位/架构/方案/契约/验证/里程碑 |
| w06-architecture-panorama.md v1.1 | 理论附档：数理化推导/图执行机制/主 Agent 知识体系/验收映射 |
| 架构图（docs/diagrams/） | fig1 完整总架构（v3.0）· fig2 上半部（用户+Go 层字段级）· fig3 下半部（Runtime+专家层+故障对照）· fig4 主 Agent 知识与候选图生成 |
| w06-agent-skill-rationality-analysis-1117.md | R1-R7 整改依据（逐项审查） |
| w06-gap-vs-github-1117.md | N1-N8 缺口依据（#16 对照） |
| README.md（工程内） | 构建/运行（待按 v3 更新形态说明） |

## 附 B：版本演进记录

- **v1（11-13 前）**：单 checker 服务演示期
- **v2（11-13/14）**：五角色+GatekeepRail+compose 设计（DeepAgent root 自编排形态）
- **v2.5（11-16 夜）**：skill 五件套/MCP 闭环/装配引擎/运行面板上线（工作台治理层）
- **v3.0（本版，11-17）**：**形态重构定稿**——专家节点+核验器定位，三接口契约冻结，M1-M4 里程碑与 #16 最后两个工作包的收官路径
