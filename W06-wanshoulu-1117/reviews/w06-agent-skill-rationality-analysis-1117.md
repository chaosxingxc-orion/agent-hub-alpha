# Agent 与 Skill 设置合理性系统分析（2026-11-17 07:00）

> 数据基线：线上运行时 10 agent + 市场目录 3 + 技能注册表 7 + 骨架配置 14 节（report-structure.json 实读）+ 六工位分工 skill。
> 评估框架：①目标架构对位（#16 图执行架构）②职责边界 ③skill 与 agent 的匹配度 ④工位/骨架/注册表三层一致性 ⑤发现的具体问题。

## 一、Agent 清单合理性与问题（逐个）

### 1.1 按角色定位逐一审查

| Agent | kind | 定位 | 问题 |
|---|---|---|---|
| editor-in-chief 主编 | ORCHESTRATOR | 对话流主编+分派 | **角色错位**：骨架配置已把分派写死（14 节→agent 映射），主编实际职能只剩"受理+澄清+汇总"。按图执行架构，分派权在图/调度器，主编应定位为**任务接口人**而非编排者。当前注册表却让它背"多agent分派"能力——与骨架配置矛盾（两个分派真相源） |
| material-agent 资料 | ANALYST | 快照装载唯一入口 | ✅ 职责清晰。但 kind=ANALYST 名不副实——它不做分析，是**工具型**角色（读快照），更接近 TOOL/WORKER。六工位把它和 finance-analyst 并为"查数工位"说明定义时就模糊 |
| finance-analyst | ANALYST | 经营分析 | **职责过载**：骨架里负责 6 个节（S-1/S-3/S-4/S-6/S-7 + 上年同期），risk/capital 各只有 1 节。负载 6:1:1 失衡——不是业务性（银行年报经营类确实占大头），但"上年同期对比"（S-7）给了它而非 risk，分类依据不明 |
| risk-analyst | ANALYST | 资产质量 | ✅ 单节（S-2 贷款质量）清晰。但它没绑 metric-dictionary skill——**不良率/拨备覆盖率的口径复算规则它看不到**，而它恰好是这些指标的主要分析者（skill 绑定缺位） |
| capital-analyst | ANALYST | 资本充足 | ✅ 单节（S-5）清晰，绑 metric-dictionary 合理 |
| disclosure-officer | ANALYST | **职责三合一** | 最大问题：骨架里承担 7 个 static 节（S-8/S-11/S-12/S-9/S-10/S-13/S-14）=**静态正文填充**；六工位里它是"写作工位"（产出 AnnualReportArtifacts）；注册表描述却是"披露规范审查"。三个身份（静态填充器/主笔/合规审查）职责漂移，骨架把 7 节 static 全压给它使其成为事实上的瓶颈节点 |
| chart-agent | ANALYST | 图表 | ✅ 定位清晰（Trae 新增，"只渲染不改数"纪律好）。但六工位表格里写作工位输入含"图表规格"——图表产物应该流入写作，当前骨架 14 节没有任何节分派给 chart-agent（**它没有活干**：section_order 里无图表节，chart 产物无处挂载） |
| numeric-checker | CHECKER | N1-N7 | ⚠️ 版本 0.2.1 与 seedAgents 描述里的"N1–N7"——**N7 百分点规则还没实现**（对照结论 N4），描述超前于实现 |
| citation-checker | CHECKER | 引用溯源 | ✅。描述"摘录指纹比对"与实现一致 |
| approver | APPROVER | 具名批准 | ✅ 定位对。但它在"主编团队"里——**批准者不应是编排者的下属**（批准链独立性：批准人不能被主编分派控制，否则 fail-closed 的独立性受损。#20 E01/E02 的精神：批准凭据必须来自受信边界而非编排链） |

**市场目录 3 个（peer-benchmark / macro-review / layout-reviewer）**：合理（未完成能力不入运行时，版本 0.9.0 标注清楚）；但 peer-benchmark"同业排名论断"与"快照之外的信息不进正文"铁律冲突——安装后若真生成同业论断会违反 S-10 的静态声明。**安装前的治理审查缺位**（能力卡没有"会违反哪条铁律"的标注）。

### 1.2 结构性问题汇总

1. **分派真相源分裂**：team[]（注册表）vs report-structure.json（骨架）vs agent-division.md（六工位）三处各定义一次"谁干什么"，已经开始互相矛盾（见 editor-in-chief 与 chart-agent）。
2. **writer 角色缺失**：六工位定义了"写作工位=disclosure-officer"，但系统里没有独立 writer/主笔 agent——disclosure-officer 被迫兼任。而我们的 Java 侧明明有完整 WriterAgent（年度报告工件的生产者）——**server 注册表与 Java 实现脱节**（orchestrator 里五角色含 writer，前端注册表却没有）。
3. **kind 体系不完整**：ORCHESTRATOR/ANALYST/CHECKER/APPROVER 四类覆盖不了实际形态——material-agent 是工具型（应 TOOL），disclosure-officer 的静态填充是文档组装型（应 WRITER），chart-agent 是渲染型。kind 与实际行为不符会让规划侧（若接 W02/Builder）按错误的能力语义召回。

## 二、Skill 清单合理性与问题

### 2.1 逐个审查

| Skill | 绑定 | 问题 |
|---|---|---|
| annual-report 主技能 | 全部 10 agent | **过度绑定**：主技能是流程总纲（五步核验协议），核验器/批准官需要的是"消费核验结论"部分，不需要全文。全员绑定=没有重点，也使主技能失去"谁该读什么"的指向性。**官方渐进披露原则：reference 按需加载，主技能应只给主编+写作** |
| report-skeleton 骨架 | 仅 chart-agent | **严重错绑**：骨架是全系统的分派真相源，最该读的是**主编**（分派）和 **disclosure-officer**（静态节生产者），却只绑了 chart-agent（骨架里根本没有图表节） |
| agent-division 六工位 | 仅 chart-agent | 同上错绑：分工纪律是主编和全工位的契约，只绑 chart 说不通 |
| banking-domain | finance/risk/capital | ✅ 合理（分析师绑领域知识） |
| citation-style | disclosure-officer/citation-checker | ✅ 合理。但 finance/risk/capital **没绑**——它们产出带 [C-n] 引用的草稿，引用格式规范恰恰是分析工位最需要的（铁律"每论断恰好一条引用"写在 citation-style 里，分析师看不到） |
| metric-dictionary | finance/capital/numeric-checker | **risk-analyst 缺绑**（见 1.1）；writer（若存在）也该绑——"metrics 原样转抄"的口径规则在此 |
| source-library | 仅 material-agent | 基本合理（快照装载唯一入口）；但 citation-checker 做信源注册表校验（SOURCE_UNREGISTERED 规则）也依赖此文档，缺绑 |

### 2.2 结构性问题

4. **skill 粒度混淆**：7 个"技能"里 1 个是主技能（方法论）、6 个是 reference（领域文档）。官方语义里 reference 是 skill 的**附件**，不是平级技能——现在注册表把 reference 注册成了独立技能，导致"绑定"语义混乱（绑 banking-domain 和绑 annual-report 是两种完全不同的操作）。前端资产库把它们并排展示也强化了这个混淆。
5. **绑定没有跟随消费关系**：正确模型应该是"**谁执行什么工位→绑那个工位必读的 reference**"，现在的绑定是手工拍的，已出现四处错绑/漏绑（skeleton/division/citation-style/source-library）。

## 三、骨架配置与注册表的交叉验证

report-structure.json 实读发现的深层问题：

6. **静态节占比过高**：14 节里 7 节是 static（50%）——真实年报的叙事主体（董事长致辞、管理层讨论、ESG、治理）全部退化为模板填空。对"简报"定位合理（#16 demo 边界），但静态节的"责任 agent"标注意义存疑——静态正文不经过 agent 生成，标 disclosure-officer 只是"填表人"。
7. **S-9"公司治理"节的静态正文是核验说明**——把交付治理元数据（N1-N7 清单）写进报告正文，混淆了"报告内容"与"系统行为"两层。真实年报不存在这种章节。
8. **工位-节-指标三层没有一致性校验**：S-2 贷款质量（risk-analyst）需要的不良率/拨备指标在 metric-dictionary 的分析子集里，但骨架没有声明该节"requires 哪些指标"（report-skeleton.md 的示例结构有 metric_keys 字段，实际 report-structure.json 没有）——分析工位的输入注入**无配置依据**，靠隐式约定。

## 四、合理性总评

**设计意图正确、执行一致性差**。工位分工（六工位）、知识分层（主技能+reference）、骨架驱动（配置即分派）三个设计决策都是对的；但三层之间（agent 注册表 / skill 绑定 / 骨架配置）**没有单一事实源约束，已经产生 5 处矛盾**（分派双真相源、chart 无节可干、writer 缺位、4 处 skill 错绑漏绑、kind 失真）。

## 五、整改建议（按结构收益排序）

| # | 建议 | 性质 |
|---|---|---|
| R1 | **注册表重构为骨架派生**：agent 的工位/节分派不再在 seedAgents 硬编码，改为启动时从 report-structure.json 计算（谁负责哪些节→公开什么能力→必读哪些 reference）——消灭分派双真相源 | 结构（0.5 天） |
| R2 | **补 writer-agent**：Java 侧 WriterAgent 上注册表，disclosure-officer 回归纯合规审查；静态节填充独立成 assembler 角色（或明确 static 节不经 agent） | 结构（0.3 天） |
| R3 | **skill 绑定改为消费关系驱动**：按"工位必读表"自动绑定（分析→citation-style+metric-dictionary；核验→source-library+各自规则）；主技能从全员绑定收缩为主编+writer | 结构（0.2 天） |
| R4 | **kind 体系补全**：加 TOOL（material）/WRITER（writer/assembler）/RENDERER（chart），注册表描述与实现同步（numeric-checker 描述去掉未实现的 N7 或实现 N7） | 一致性（0.2 天） |
| R5 | **approver 移出主编 team**：批准链独立（主编可请求批准，不可分派/控制批准官） | 治理（0.1 天） |
| R6 | **骨架加 metric_keys 声明**：每 analysis 节声明消费指标子集，作为输入注入的配置依据 + 一致性校验 | 一致性（0.3 天） |
| R7 | **注册表区分"主技能/reference"两级**：前端与 API 都标注类型，绑定为"绑技能包"而非"绑文档" | 概念修正（0.2 天） |

合计约 1.8 人日。R1-R3 是根（消灭矛盾源头），R4-R7 是表面一致性。

## 六、与 #16 要求的映射

这些不一致若带到集成验收会直接踩 #16 的审查点：分派双真相源违背"契约冻结"（D1 要求单一事实源）；approver 从属主编违背 E02（伪造批准链）；skill 错绑会让规划侧（W02 消费能力卡）召回错误能力——**建议在角色接入 W05 RDC（N6）之前完成 R1-R3**，否则错绑会被固化到 A2A 能力卡里。
