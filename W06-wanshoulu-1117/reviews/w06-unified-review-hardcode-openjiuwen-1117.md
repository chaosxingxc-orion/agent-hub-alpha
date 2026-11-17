# 统一完整审视报告（按用户两条硬性标准）

**标准**：①不允许硬编码写死，必须多任务泛化 ②Agent/skill/MCP 符合 openjiuwen 与业界实践（记忆、上下文引擎、agent loop、rail 机制都要用）
**方法**：全代码扫描（本地工程+线上 server+skill 脚本）+ 与 openjiuwen 上游 demo（multi-deep-research-demo）逐项对照
**结论先行**：泛化性 **B+**（发现 4 处真硬编码，其中 2 处必须改）；openjiuwen 规范符合性 **C+**（框架用了但只用了 40%——记忆/上下文引擎/多数 rail 未接入，四个分析角色是"裸 ReActAgent"）。

---

## 一、硬编码问题（标准①）

### H1（必改）· CitationChecker 实体表写死
```java
String[] entities = {"招商银行", "本集团", "营业收入", ...}  // 摘录与论断的实体重叠判定
```
**问题**：机构名（招商银行）+指标词全部写死。换银行=换指标集时此函数语义漂移——工商银行任务里"招商银行"实体永不命中。
**修法**：实体表从 MetricKey 枚举的 displayName 动态生成 + 机构名从工件 company 字段取。**0.2 天**。

### H2（必改）· N7 指标名词表写死
```java
"(不良贷款率|不良率|拨备覆盖率|成本收入比|核心一级资本充足率|资本充足率|净息差)"
```
**问题**：比率类指标清单硬编码在正则里。骨架换成"零售贷款占比"类指标时 N7 失明。
**修法**：从 MetricKey 枚举中按"unit=%"筛选动态拼正则（枚举已含 unit 信息或补字段）。**0.2 天**。

### H3（可接受但需声明）· MetricKey 八指标枚举
**判定**：**不算违规硬编码**——它是"口径字典"（W01 冻结的验收标准本身就是八指标）。但泛化边界必须声明：换指标集=换枚举+重跑 evals，这是**数据配置变更**而非代码缺陷。与 report-structure.json/banks 配置同层。建议在 MetricKey 注释里写明"本枚举是 W01 验收口径的代码映射，指标集变更走配置流程"。

### H4（可接受）· bankOf 三银行映射
**判定**：降级可接受——快照目录结构（cmb/icbc/ccb）本身是 W01 材料的冻结事实，映射表与其对齐。**但**新增银行时此函数静默返回 null（默认行兜底），应在返回 null 时打 WARN 日志而非静默。

### H5（已达标项）
- ✅ 骨架驱动：节序/标题/分派/metric_keys 全在 report-structure.json（P6/P7 后配置即泛化）
- ✅ 注册表骨架派生（P7）：团队/能力从配置算，不再硬编码
- ✅ pipeline 服务端"零机构特判"（注释自证+代码验证）
- ✅ bankDisplayName switch——演示层文案，展示用可接受（但 run/start 的"快照 v1.0"版本号写死，应从 manifest 取）

## 二、openjiuwen 规范符合性（标准②）

### 符合的部分（做对了）

| 项 | 现状 | 上游对照 |
|---|---|---|
| 角色工厂模式 | 5×ReActAgent 工厂，照 agent-search 模式 | ✓ 同构 |
| ReActAgent + AgentCard + configureModelClient | 标准 | ✓ |
| GatekeepRail 挂 root（extends MemoryRail，afterInvoke 生命周期） | fail-closed 语义正确 | ✓ 上游 AutoPersistMemoryRail 同位 |
| CriticAgent 的 reactrails 套件 | ReplanRail/CriteriaReplanBridgeRail/RootCauseRail/Observability 全套 | ✓ 与上游 verify-agent 同款 |
| DeepAgent root（orchestrator）enableTaskLoop+TaskPlanning | v3 降级为节点工厂后保留 | ✓ |
| A2A：RemoteAgentCaller + 回调坐标 | compose/工厂已备 | ✓ |

### 不符合的部分（差距清单，按严重度）

**G1（大）· 四个分析角色是"裸 ReActAgent"——零 rail 零工具**
实测：material 有 4 个工具（read_page 等），而 **metrics/operation/assetquality/writer 四工厂：tools=0, rails=0**——纯 system prompt 的 LLM。这意味着：
- **agent loop 没有纪律约束**：上游 search-agent 至少挂 AskUserRail（人工中断），我们的分析角色没有任何 rail——模型重试/歧义时没有结构化出路；
- **分析角色没挂工具**：上游模式是"查数据经 Go 按契约放行"（search-agent 的 web_search 工具），我们分析角色本应挂 w06-compute（不心算铁律的工具面）——**MCP 做好了但没接到角色上**；
- 记忆不落地：每次调用冷启动，跨任务不积累。

**G2（中）· openjiuwen 记忆引擎未接入**
GatekeepRail extends MemoryRail 但**只为拿生命周期钩子**——write_memory/read_memory 的实际记忆读写没用。上游 AutoPersistMemoryRail 的语义（"每个实质答案确定性落盘"）我们没继承。应做：核验结论/修订指令写入记忆，供下轮任务复用（对应 #16 的"知识复用：命中/换版本不命中"验收）。

**G3（中）· 上下文引擎（ModelContext）浅用**
GatekeepRail 拿 ModelContext 只取了 output 文本。openjiuwen 的 ModelContext 承载会话/工件/迭代状态——执行身份（P4 的 ExecutionO）现在走自建 ThreadLocal，**正确姿势是挂进 ModelContext 的会话语义**（与上游对齐，调度侧无需感知我们的私有 ThreadLocal）。

**G4（低）· AskUserRail 未挂任何角色**
上游所有对话位角色挂 AskUserRail（中断-恢复）。我们"澄清"职责写在主编 SKILL.md 里但没有机制支撑——语义灰区（PARTIAL→waiting_human）只有 GatekeepRail 一条路，节点内的澄清中断缺失。

**G5（低）· MCP 三形态与 openjiuwen 工具注册的关系未闭合**
MCP 工具的"bind 即装配"是自建装配引擎语义；openjiuwen 侧的工具注册走 resourceMgr.addTool + LocalFunction（material-agent 就是这么挂 read_page 的）。**两套工具面并存**：节点内用 LocalFunction，跨节点用 MCP——业界实践是 MCP 工具可经 adapter 挂成 LocalFunction，统一进 agent loop 的工具选择。

## 三、整改清单（并入 M2）

| # | 项 | 内容 | 工作量 |
|---|---|---|---|
| **G1-a** | 四角色挂纪律 rail | 各挂 ReplanRail（重试上限）+ AskUserRail（歧义澄清）；writer 额外挂修订指令接收语义 | 0.5 天 |
| **G1-b** | 分析角色挂计算工具 | metrics/三分析/ writer 挂 w06-compute 的 eval_expression/yoy_calc 为 LocalFunction（MCP→LocalFunction adapter，上游 ToolCard 模式）——"不心算"从纪律变成机制 | 0.5 天 |
| **G2** | 记忆接入 | GatekeepRail 真 extends MemoryRail 语义：verdict/修订指令 write_memory；任务开始 read_memory 查历史结论（知识复用验收） | 0.5 天 |
| **G3** | ExecutionO 进 ModelContext | ThreadLocal 改挂 ModelContext 属性（调度侧从 dispatch 上下文设置） | 0.3 天 |
| **G4** | 主编挂 AskUserRail | 澄清职责机制化 | 0.2 天 |
| **H1** | 实体表动态化 | CitationChecker 从 MetricKey+company 生成 | 0.2 天 |
| **H2** | N7 词表动态化 | 从 MetricKey 按 unit=% 筛选拼正则 | 0.2 天 |
| H5-补 | run/start 版本号从 manifest 取 | 一行改 | 0.05 天 |

合计 **2.45 人日**。全部完成后：泛化性 A（配置驱动+字典声明边界）、openjiuwen 符合性 B+（rail/记忆/上下文/工具全接入；A 级需接 W05 真实 RDC 联调后才能声明）。

## 四、诚实边界

- **G1 是结构性差距**（不是漏改）：四个角色当前是"prompt-only agent"，这是 v2 快速验证期的产物。修 G1 本质是把角色从"能演示"升级到"符合上游工程标准"——0.5+0.5 天可以完成 rail+工具接线，但**真正生效需要 LLM key**（工具选择与 rail 中断都发生在 loop 里）。
- 演示面（线上工作台）的规范符合性已达标：装配/MCP/运行轨迹都按 openjiuwen 语义建模；本节差距全部在**执行面**（角色进程内部）。
