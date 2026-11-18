# W06 Annual Report Agent — 招商银行 2025 年报多 Agent 报告生成与核验

> 基于 openJiuwen `agent-runtime-java`（固定快照 `bf702bc`）构建的 Issue #16 W06「报告 Agent 与核验」参考实现。
> 架构完全复用 `multi-deep-research-demo` 的两层模式：**库层无 Spring、wrapper 层做 Spring 装配**，远端子 agent 经 A2A 注册。

## 设计映射（W06 验收标准 → 本实现）

| W06 要求（#16 11-14 交接表） | 实现位置 |
|---|---|
| 资料、经营分析、资产质量分析、综合报告角色 | 5 个 A2A 远端子 agent（material/metrics/operation/asset-quality/writer）+ 1 个 root 编排 agent |
| 确定性/限定语义检查 | `checker` 模块：NumericChecker（确定性复算）+ CitationChecker（逐条引用核验，语义支持度判定） |
| 从原始快照产出 | `MaterialAgent` 只读挂载的固定年报快照（PDF/文本 + manifest），无任何网络搜索 |
| 来源可打开 | 每个数值携带 `source_document_digest + page + evidence_excerpt_digest`，checker 逐项校验 |
| 完整性/正反例达标 | `ContractVerdict` JSON 结构化输出 + 反例测试（错误单位/伪造引用/检查后篡改） |
| 检查失败阻止验收 | `GatekeepRail`：verify FAIL → 打回一次定向修订 → 再 FAIL → `waiting_human`（不是 best-effort） |

## 模块结构

```
w06-annual-report-agent/
├── agent-orchestrator/     # root DeepAgent + GatekeepRail + 产物落盘（对应 demo 的 agent-deep-research）
├── agents/                 # 5 个专业角色子 agent（对应 demo 的 agent-search/agent-verify）
├── checker/                # 确定性核验器（独立进程，不依赖模型——数字必须从工具来）
├── deploy/                 # compose + versions.lock + 年报快照挂载
└── pom.xml
```

## 与上游示例的三处关键差异（W06 特化）

1. **verify 从「best-effort 质量闸」改造为「交付门禁」**：上游 `verify-agent` 明确「失败不阻塞交付」；本实现 `GatekeepRail` 相反——Critic FAIL 阻断交付、触发一次定向修订、再失败进 `waiting_human`（对应 #20 对「检查只是建议面」的质疑）。
2. **数字纪律**：八指标值只能来自 `MetricAgent` 的结构化抽取 + `NumericChecker` 的确定性复算；writer 口述的数字若与 metrics artifact 不一致，`NumericChecker` 一律 FAIL（数字只从工具来，参考 SEC-EDGAR 的 verifier 模式）。
3. **无搜索路径**：上游 search-agent 的 web_search 整体移除，MaterialAgent 只读快照卷（对应 GPT-Researcher local 模式的取舍）。

## 快速开始

```bash
# 1. 锁定上游版本（versions.lock 已固定 bf702bc）
# 2. 构建镜像并启动（compose 内部网络，不映射宿主端口，经 Agent Bus RDC/Gateway 调用）
cd deploy && docker compose up -d --wait
# 3. 反例回归（错误单位、伪造引用、检查后篡改、无依据数字）
mvn -pl checker test
```

## 状态

**参考实现骨架（W06 认领前的草案）**——尚未接入真实 Agent Bus RDC、未经 W01 材料独立核验、未跑 A01–A15。按 #16 口径，本仓库任何内容不构成工作包已完成或验收通过的声明。
