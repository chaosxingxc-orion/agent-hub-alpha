# 年报骨架方法（真实年报骨架 · 名称保留 / 深度收缩）

> `report-structure.json` 是节序/标题/静态正文/责任 agent 的**唯一事实源**。本文件说明其结构与使用纪律；新增银行零代码——只加配置，不改代码。

## 骨架配置结构（report-structure.json）

```json
{
  "bank": "招商银行",
  "period": "2025年度",
  "sections": [
    {
      "id": "s1",
      "title": "重要提示与释义",
      "static_body": "<静态正文：责任声明/审计意见模板>",
      "owner_agent": "disclosure-officer",
      "requires": []
    },
    {
      "id": "s2",
      "title": "会计数据和财务指标摘要",
      "owner_agent": "finance-analyst",
      "requires": ["metrics"],
      "metric_keys": ["OPERATING_REVENUE", "NET_PROFIT_ATTRIBUTABLE", "NET_INTEREST_MARGIN", "COST_INCOME_RATIO"]
    }
  ]
}
```

## 使用纪律

1. **节序不可变**：执行顺序以 sections 数组序为准，agent 不得自行调整；
2. **owner_agent 必须在运行时注册表内**：分派时若 owner 未注册，任务进入 waiting（不静默换 agent）；
3. **static_body 原样输出**：静态正文（责任声明/免责）一字不改，不参与核验修订；
4. **新增银行**：复制一份配置改 bank/指标页码/metric_keys 即可——任何需要改代码的需求先怀疑配置缺失；
5. **骨架版本**：配置文件带 version 字段，采纳图与工件记录引用之（AR-04 同语义）。

## 与其他 reference 的关系

- 节内指标口径：`metric-dictionary.md`；引用格式：`citation-style.md`；
- 六工位职责边界：`agent-division.md`；信源铁律：`source-library.md`。
