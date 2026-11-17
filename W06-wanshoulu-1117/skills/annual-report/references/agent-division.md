# 六工位分工（多 Agent 协作纪律）

> 查数 / 分析 / 图表 / 写作 / 核验 / 主编六工位的输入-输出-产物契约。**分派来自骨架配置**（report-skeleton.md），不由主编临场指定。

## 工位契约表

| 工位 | agent | 输入 | 输出（产物） | 纪律 |
|---|---|---|---|---|
| 查数 | material-agent / finance-analyst | 快照 manifest + 骨架 metric_keys | MetricValue[8]（含 digest/page/excerpt 三元组） | 值照抄原文；缺材料报 MATERIAL_MISSING；禁推算 |
| 分析 | finance-analyst / risk-analyst | 指标子集（按骨架 requires 注入） | 带 [C-n] 引用的分节草稿 | 只消费注入指标；派生值标注"基于 [C-k] 计算"+算式 |
| 图表 | chart-agent | 指标工件 + 图表规格 | 图表产物（结构化描述+数据表） | 图表数据必须来自指标工件；禁自行取数 |
| 写作 | disclosure-officer | 骨架 + 各工位产物 | AnnualReportArtifacts 工件 | metrics 原样转抄；论断只来自上游草稿 |
| 核验 | numeric-checker / citation-checker | 完整工件 + 快照 | CheckResult（双核验） | 运行脚本出结论；FAILED 打回一次；再失败转人工 |
| 主编 | editor-in-chief | 任务 + 骨架配置 | 分派指令 + 修订指令 + 交付决定 | 只按骨架分派；不放行未过核验的工件 |

## 分派纪律

1. **来源唯一**：每个节点的 owner 由 report-structure.json 决定；主编不越权改派（owner 不可用时任务 waiting，换人须人工确认）；
2. **输入隔离**：分析工位只可见骨架注入的指标子集——经营类与资产质量类互斥（越界引用被 citation 核验拦截）；
3. **产物不可变**：下游工位不得修改上游产物，只能产出新工件（修订 = 新 digest，不覆盖）；
4. **核验先行**：写作产物未经双核验 PASSED 不得进入交付决定；主编的"交付"动作绑定核验结论与工件 digest。

## 与其他 reference 的关系

工位执行细节见各自铁律：指标口径 `metric-dictionary.md`、引用 `citation-style.md`、领域陷阱 `banking-domain.md`、骨架分派源 `report-skeleton.md`。
