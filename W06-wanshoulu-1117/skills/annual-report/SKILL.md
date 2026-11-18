---
name: annual-report
description: 从固定年报快照生成可核验的银行年报简报，并对报告工件做数值与引用核验。凡是涉及银行/上市公司年报分析、年报指标抽取、生成带引用的财务简报、核验报告数值或引用、或用户提到"年报""财务指标""不良率""拨备覆盖率"等年报场景的任务，都应使用本技能——即使用户没有明确说"年报简报"。
version: 1.1.1
---

# 年报简报生成与核验指南

## 概述

本指南覆盖从固定年报快照生成简报的完整流程：五角色协作、结构化工件成稿、双层核验、修订与人工门禁。口径细则见 references/metric-dictionary.md；引用格式见 references/citation-style.md；需要填写引用前，先读 references/citation-style.md 并遵循其规范。

**核心理念**：确定性操作运行脚本，模型的判断力只用于内容。数字一律照抄原文，不心算、不推算；所有派生值与核验结论来自脚本输出。

## 快速开始

```bash
# 数值核验（完整性/单位/溯源/三口径复算/同比）
python scripts/verify_metrics.py <artifact.json> <snapshot_dir>
# 退出码：0=通过 1=FAILED(含BLOCKING) 2=仅WARNING(建议人工)

# 引用核验（三元组比对/信源注册表/支持度三级）
python scripts/check_citation.py <artifact.json> <snapshot_dir>
```

`<artifact.json>` 按 `assets/report-template.json` 骨架填充；`<snapshot_dir>` 含 `manifest.json`（信源注册表）与 `pages/page-NNN.txt`。

## 工作流程

按以下顺序推进，禁止跳步：

1. **资料装载**：读快照 manifest 的 `metric_index` 定位指标页；读取页原文，记录 digest/page/excerpt 三元组。缺材料时如实报 MATERIAL_MISSING，绝不编造。
2. **指标抽取**：按 `assets/metric-registry.json` 的 expected_unit 填八指标。值照抄原文（含千分位逗号）；单位用原文单位；找不到输出 NOT_FOUND。
3. **双分支分析**：经营分析只可消费营业收入/归母净利润/净息差/成本收入比四项；资产质量分析只可消费不良余额/不良率/拨备覆盖率/核心一级资本充足率四项。两个集合互斥——越界引用会被核验脚本拦截。
4. **成稿**：按 `assets/report-template.json` 填充。metrics 数组原样转抄（一字不改，改动会被核验失败）；每条论断恰好一条 [C-n] 引用（格式见 references/citation-style.md）。
5. **核验**：运行上述两个脚本。两者均 PASSED → 可交付；任一 FAILED → 按违规清单定向修订一次（只修违规项，其余原样）；再失败或仅 WARNING → 转人工，绝不放行。

## 常见陷阱

**百分比与百分点**：资本充足率从 14.86% 降到 14.16% 是"下降 0.70 个百分点"，不是"下降 0.70%"。写作与核验都区分二者。

**同比必须可复算**：正文写"同比增长 X%"时，X 必须与两期值一致（快照的上期信源页）。写不出算式的判断不写。

**派生值标注**：任何非照抄的计算值必须写 `基于 [C-k] 计算：<算式>`，算式中的每个输入数字必须出现在被引证据中。

**跨期数据**：上期值必须来自上年度独立信源（快照的 2024 信源页），禁止用本期报告的"上年同期"栏做唯一依据（防排版错位）。

## 输出格式

报告工件 ALWAYS 使用 assets/report-template.json 的结构：

```
task_id / company / report_period
metrics[8]   —— key/value/unit/period/source_document_digest/page/evidence_excerpt_digest
citations[]  —— claim_id/claim_text/source_document_digest/page/evidence_excerpt/evidence_excerpt_digest
sections[]   —— section_id/title/body/claim_ids
```

**Example 1（指标条目）:**
Input: 快照第 12 页原文"实现营业收入3,374.88亿元"
Output: `{"key":"OPERATING_REVENUE","value":"3,374.88","unit":"亿元","period":"2025年度","page":12,...}`

**Example 2（核验结论消费）:**
Input: `verify_metrics.py` 退出码 1，violations 含 `FABRICATED_NUMBER: CET1_RATIO 值 999.99 未在第 45 页原文出现`
Output: 将 CET1_RATIO 修订为原文值 14.86 后重跑核验；不得调低阈值或跳过该指标

## 后续步骤

- 指标口径、复算公式与容差：references/metric-dictionary.md（规则源为 assets/metric-registry.json）
- 监管口径与会计口径差异、分析视角陷阱：references/banking-domain.md
- 信源分级标准与引用谱系纪律：references/source-library.md
- [C-n] 协议与摘录要求：references/citation-style.md
- 评估用例（含迷你快照，可直接跑脚本）：evals/，期望见 evals/expected.json

## 已知限制（如实声明）

- 指标集为 W01 冻结的八项，不含分部收入/流动性/表外等全量科目；
- 支持度判定的语义灰区由 LLM Critic 兜底但最终转人工，不由 judge 单独批准；
- 快照固定，无实时联网（设计约束，非缺陷）；展望章节不在本技能范围（demo 禁止预测）；
- 本技能不使用 MCP 工具；A 股结构化数据 MCP 在路线图（见 references/source-library.md）。
