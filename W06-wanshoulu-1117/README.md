# W06 · 银行年报智能体参考实现（agent-hub-alpha 场景切片）

> 状态：**可运行的参考实现**，跑通过完整链路并产出成品，非生产系统。
> 评审请求见 [REVIEW-REQUEST.md](REVIEW-REQUEST.md)。

## 这是什么

W06 把 `agent-hub-alpha` 的 S1/S2 智能体架构，落到「银行年报编制」一条端到端链路上：从落盘的银行原始材料出发，经取数、指标、核验、成稿、主编审批，产出**可溯源、可核验、可具名批准**的结构化年报。

它验证的不是模型能力，而是三件事：**决策可回放**（工件+引用可追到来源页）、**执行可核验**（数值/引用两道确定性检查）、**结果可审计**（具名批准+指纹防篡改）。措辞与保证范围对齐本仓 README 的纪律：一致不等于正确，重复得到错误答案也不符合目标。

## 名词（首次出现就地解释）

- **artifact / 工件**：一次编制的结构化结果，含 `metrics` / `sections` / `citations` / `tables` / `charts`。
- **双检**：NumericChecker（N1–N7 数值规则）+ CitationChecker（C1–C2 引用规则）。
- **writer**：openJiuwen deep agent，通过 `skill-app-agent` 容器调用，负责长文成稿。
- **工具**：系统对外暴露的能力，`web_search` / `web_fetch` / `pdf_extract` / `snapshot_commit` / `table_extract` / `table_analyze`。
- **批准链**：approver 对工件指纹签名；工件被改则指纹失配、拒绝批准。

## 交付物

| 目录 | 内容 |
| --- | --- |
| `source/w06-annual-report-agent/` | Maven 工程：agent-orchestrator / agents / checker 三模块，23 个权威 Java 类（线上实跑版已覆盖） |
| `source/config/` | `tools.json`（6 工具注册）、`report-structure.json`（28 节骨架） |
| `skills/` | `annual-report`（成稿方法论与核验夹具）、`orchestrator-skill`（主编知识包） |
| `frontend/workbench.html` | 工作台：对话 / 资产库 / 报告库三页（Agent Hub 已并入资产库，演示链降级为二级入口） |
| `reviews/` | 架构全景、实现审视、技能合理性、差距分析等 6 篇 |
| `artifacts/` | 工行 2025 年报成品 HTML（27 章/50 表/81 引用）+ `gen_report_v2.py` 渲染脚本 |

## 运行方式（最小）

系统以容器方式部署在远端，核心是 `w06-checker`（Java HTTP 服务 + 工具网关 + 核验器）与 `skill-app-agent`（writer）两个容器。

```bash
# 构建（见 source/w06-annual-report-agent/deploy/checker.Dockerfile + compose.yaml）
docker compose -f source/w06-annual-report-agent/deploy/compose.yaml up -d

# 健康
curl http://<host>:18083/health

# 跑一次编制（brief 快验 / long_form 完整长文）
curl -X POST http://<host>:18083/v1/pipeline/run \
  -H 'Content-Type: application/json' \
  -d '{"bank":"icbc","mode":"brief"}'

# 直接调数据工具
curl -X POST http://<host>:18083/v1/tools/call \
  -H 'Content-Type: application/json' \
  -d '{"tool_name":"table_extract","arguments":{"bank":"live"}}'
```

核验与批准：

```bash
curl -X POST http://<host>:18083/v1/checks/numeric  -H 'Content-Type: application/json' -d '<artifact JSON>'
curl -X POST http://<host>:18083/v1/checks/citation -H 'Content-Type: application/json' -d '<artifact JSON>'
curl -X POST 'http://<host>:18083/v1/approvals?approver=Audit' -H 'Content-Type: application/json' -d '<artifact JSON>'
```

## 状态与已知限制

成品效果与测量口径见 [REVIEW-REQUEST.md §3](REVIEW-REQUEST.md)。已知限制摘要：writer 长文生成有波动（靠超时+续写兜底）；表格不进双检规则；招行/建行快照仅节选、无可抽表格。完整清单见 REVIEW-REQUEST §5。

本实现对照本仓架构 v3 仍有差距（如未接 PostgreSQL 权威账本），详见 `reviews/w06-gap-vs-github-1117.md` 与 `reviews/w06-implementation-review-vs-v3-1117.md`。

## 评审

请 chaosxingxc-orion 按 [REVIEW-REQUEST.md](REVIEW-REQUEST.md) 评审架构 / 代码 / 效果；不正确之处将及时修正回传。
