# W06 架构方案与功能对照 #16 最新要求（2026-11-17 00:17）

> 基准：#16 顶部交付表 + 26 条评论中的**技术要求**（yougq 的 W03/W04 契约、xierongqiag2015 的 W05 接口要求、W01 的材料与反例规格、11-13/14 的产品预期与 AR 清单）。流程性内容一概不谈。

## 一、目标架构 vs 我们的架构（组件级逐项对照）

#16 最新评论拼出的目标架构（yougq 已实现 W03/W04 侧）：

```
WorkSwarm 对话 → 消费获准能力目录 → 生成候选图
→ Go 校验、Java 编译、用户确认 → 采纳图 T/g
→ Java BlueprintScheduler 推进图（W03）
→ Gateway/RDC 经真实 A2A 调四个专家（W05→W06 的角色）
→ PG 正式提交结果与审计（W04：o=(T,g,node_path,iteration_path) 唯一正式结果）
→ 报告检查（W06 核验器）→ 用户批准 → 受控交付
```

逐组件对照：

| 目标架构组件 | 我们已有 | 差距 | 差距性质 |
|---|---|---|---|
| **专家角色（提取/经营/资产质量/写作）作为远端 A2A Runtime** | 五角色完整（Material/Operation/AssetQuality/Writer/Critic）+ 六工位分工 skill + a2a-wrapper 独立容器（18094，agent-card/tasks/send 真协议） | 角色**没挂到 W05 的 RDC/Gateway** 后面——现在是 checker 内进程+wrapper 演示，不是"四个真实远端 task"（yougq 验证链路已是四个远端 task） | 接入层：把角色包成独立 Runtime 进程注册进 RDC |
| **结构化执行身份 o=(T,g,node_path,iteration_path)，attempt 与执行所有权代次分离** | **无**。GatekeepRail 的 revisionRounds 只管修订轮次；agent 调用无执行身份；重试无"同输入复用原结果、冲突不覆盖"语义 | our contract 类无此字段；yougq 的 Contract-C 已实现并测试（16 项核对过），**直接消费他的共享类型即可** | 契约层：contract 包加 o/a/b 字段 + 复用规则 |
| **正式结果唯一性：PG 事务提交，checkpoint 不宣布业务成功** | 我们无 PG/无事务——GatekeepRail 在内存+文件落 verdict | W06 侧职责其实是"提供核验结论给 W04 的正式提交"，不需要自建 PG；但**核验结论必须带工件 digest 才能进他的正式记录**（见功能层 N3） | 边界明确：W06 出检查记录，W04 出正式结果 |
| **Java 唯一图推进者，Go 不建第二调度** | 我们的 root 编排走 DeepAgent Loop（enableTaskLoop+TaskPlanning）——与 BlueprintScheduler 是两个推进模型 | DeepAgent 的自由 loop vs BlueprintScheduler 的固定图推进：#16 要的是后者驱动、专家作为节点。**我们的角色应该被 his 调度，而不是自己当 root** | 形态转换：orchestrator 降级为普通专家节点（其实更简单） |
| **知识消费影响图（能力召回→适用性→选择→绑定→候选图）** | 装配引擎已有"注册→装配→工具挂载"数据层；但**图是预定义的**（六工位固定流水线） | xierongqiag2015 交付的 ALK-05（年报场景包+两份选择记录）就是给规划侧消费的——W06 侧需提供**能力卡的机器可读描述**让规划端召回（我们 SKILL_REGISTRY 的 description 已可用，但缺输入输出 schema 绑定） | 契约层：能力卡挂 inputSchema/outputSchema |
| **MCP/工具作为受控能力（白名单、版本、预算）** | MCP 5 server/12 工具真协议闭环+全限定名装配——超出要求（#16 只要求 A2A，没强制 MCP） | 无版本绑定（工具变更无版本漂移检测）——对应 AR-04 的 W06 侧"检查策略版本" | 加版本字段（小） |

## 二、功能级对照（W06 职责逐条）

| #16 要求 | 现状 | 判定 |
|---|---|---|
| 计算检查（确定性，模型不心算） | NumericChecker N1-N6 + verify_metrics.py 脚本 + MCP w06-compute（safeEval）——三形态同源 | ✅ 超配 |
| 引用检查（digest/page/excerpt + 支持度） | CitationChecker + check_citation.py + MCP w06-checker | ✅ |
| 完整性/单位/口径/期间 | N1/N2 覆盖 | ✅ |
| **"百分点 vs 百分比"**（W01 反例集明列） | **无对应规则** | ❌ N4，0.2 天 |
| **检查绑定产物摘要（κ 绑定，AR-02/E03）** | CheckResult 无 artifact digest——改正文重查出新 PASS，旧 PASS 理论上可被引用 | ❌ N3，0.3 天。E03 是 #20 明确验收用例：检查后改正文，原批准不能继续放行 |
| 检查失败阻止验收 | GatekeepRail fail-closed + 修订上限 1 + waiting_human | ✅ |
| **对 W01 材料核验**：招行年报快照（SHA 52c49dcd…，368 页）+ 8 指标参考答案 + 10 条反例全部拦截 | 演示快照是自造的三行文本页；16 反例测试基于自造数据 | ❌ N2：真实 368 页 PDF 的解析鲁棒性（findNumberAfter 对版面噪声）、W01 反例逐条拦截——都没验证过 |
| 语义灰区转人工（A08） | PARTIAL→waiting_human 链路 | ✅ |
| **报告含"口径声明/风险提示/限制说明"段**（W01 模板规格） | ReportSection 无类型字段 | ⚠️ N7，0.3 天 |
| 输出契约给 W02 展示 / W04 提交 | AnnualReportArtifacts JSON 稳定 | ✅（缺 digest 绑定见 N3） |

## 三、昨晚新增资产在目标架构中的位置

| 资产 | 在目标架构里的角色 | 对齐度 |
|---|---|---|
| MCP 5 server（JSON-RPC 2.0） | "受控能力"层的工具面——对应 W05 能力卡的调用面 | 协议真、无版本/预算字段 |
| 装配引擎（bind 即装配） | "能力声明与使用是治理和消费的过程"的直接实现 | 数据层真实；消费（loop 真调用）待 LLM key |
| Skill 五件套（scripts/assets/evals） | 专家的知识层——W01 模板+官方标准的交叉要求 | ✅ |
| 运行面板 | "界面可观测真实通信"的展示面 | 展示真实装配数据；通信内容是编排演示 |

## 四、技术整改清单（按依赖顺序，纯工程）

1. **N3 检查绑定 digest**（0.3 天）：CheckResult 加 `artifact_sha256`；GatekeepRail 校验"verdict 的 digest ≠ 当前工件 digest 则无效"；补 E03 反例测试（改正文→旧 verdict 失效）。这是 E03/E09 验收用例的 W06 侧前提。
2. **N4 百分点规则**（0.2 天）：N7 规则——指标为 % 时，正文"上升/下降 X 个百分点"必须与两期值之差一致；"X%"表述必须与相对变化一致。W01 反例"0.94% vs 0.94 个百分点"即此。
3. **N2 换 W01 真实材料**（0.5 天）：拉 PR #21 快照（368 页招行年报）进快照目录；跑通真实 PDF 版面下 findNumberAfter/摘录匹配的鲁棒性；10 条 W01 反例逐条出拦截记录（W01 交付卡明示的 W06 验收题）。
4. **N5 执行身份契约**（0.5 天）：消费 yougq 的 Contract-C 共享类型（`o=(T,g,node_path,iteration_path)`）：contract 包加字段；角色调用声明"同 o 重交复用、改内容冲突"；与 W04 对账接口语义对齐（无结果≠允许重发）。
5. **N6 角色挂 RDC**（1-2 天，需 W05 环境）：五角色从 checker 内进程改为独立 Runtime 进程（a2a-wrapper 模式已验证），注册进 RDC/Gateway；yougq 的 W03 已在等"四个真实远端 task"消费。
6. **N7 ReportSection 类型**（0.3 天）：五段模板（口径声明/指标表/分析/风险提示/限制说明）+ 指标表百分比/百分点标注。
7. **N8 LLM key**（阻塞上报）：运行轨迹的 agent 决策与 token/成本从估算变实测。

合计约 3.1 人日 + 1-2 天联调。1-4 可在 D1 前完成（2.5 天窗口内），5 在 D2 窗口，6 顺手，7 是外部依赖。

## 五、架构层面的一个判断

yougq 的实现把"图执行+正式结果"全包了（BlueprintScheduler+Go/PG，105 项测试），这意味着 **W06 的正确形态是"专家节点+核验器"，不是"自研编排"**——我们的 OrchestratorAgentFactory（DeepAgent root+五角色）应重构为：角色全部独立成远端 expert Runtime，GatekeepRail/Checker 作为 W04 提交前的检查组件。好消息：角色/核验/skill/MCP 四类资产全部可复用，重构是"拆开挂上去"，不是重写。
