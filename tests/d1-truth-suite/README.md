# claim-gate v2.7.2 ＋ D1 真值套件

计数闭合声明的机械检查闸，与其回归真值底座。本包克隆即可复现全部声明数字。

## 它解决什么问题

台账/评论/出站稿里的**计数声明**（样本数、全量用例数、变异红读数一类）与实态不符——陈旧计数、
虚报进度、声明-实态差——是本仓治理的真病灶族。claim_gate 是纯白盒词法闸：计数声明的
±8-row 窗口内必须存在可执行的复算命令，否则 FAIL。零 LLM、零网络、单文件。

## 快速复现

```
python3 -I d1-runner-full.py            # 四代闸 × 12 段真值语料 = 48 格判定矩阵
python3 -I probes/d1_probes.py          # 仪器全量自检（矩阵/声明/持久化/守恒，含派生锚）
python3 -I probes/run-mutants.py        # A 系列守恒变异 12 发（预期 12/12 杀灭）
python3 -I probes/r3-audit-battery.py   # V 系列行为审计 11 腿（战役档案 11/11；包形态含可见跳过腿，V0 为基线腿）
```

`results/battery-receipt.log` 由战役档案环境生成（V7 腿含真载体对差）；本包独立运行时
真载体缺席，仪器在探针与电池两处均**可见跳过**并说明补测方式——缺席腿跳过可见且不计入
通过声称（不假装跑满），在场环境跑满格（不放水）。

## 效果（本包内即可复核）

| 代 | F1 | 关键差 |
|---|---|---|
| v1.8 | 0.25 | 存在形态基线：精确率满但只认窄单位集 |
| v2.6 | 0.40 | 中文数词面展开伴随洪峰；白名单分裂病在 |
| v2.7.1 | 0.83 | 样本/用例/PROGRESS 入面；`ls`/`find` 反引号证据修复（v2.6.1 血统） |
| **v2.7.2** | **1.00** | REDCOUNT 反序读数形＋裸"两+个"叙事帧剔除；十事件＋双控制段全闭 |

对差复核命令（读 `results/d1-scores-v1.json`）：
```
python3 -c "import json; [print(g['name'], g['totals']) for g in json.load(open('results/d1-scores-v1.json'))['gates']]"
```

## 文件地图

- `gates/` 四代闸本体（血统标本与当前版）；v2.7.2 由 `probes/build-v272.py` 自 v2.7.1 幂等装配
- `seeds/d1-corpus-full.md` 十事件语料（真实事件的段级复原，含双控制段）
- `probes/` 跑分器、探针矩阵、两台变异电池、装配器、打包导出器、stub 闸、出厂终审电池
  （D/R/W 十四腿，可在克隆上复跑披露与复现自检）
- `results/` 跑分 JSON＋校验＋审计收据＋provenance.json（打码桥哈希）
- `docs/` 架构、规则、MR 正文、收据；`MANIFEST.md` 由导出器按实树生成

## 边界（1.00 不说的话）

- F1 1.00 覆盖**十事件套件面**，不外推：真载体文档洪峰（另一域，见 architecture §4）与
  "12 攻击样本"类隔词闭合形（已知文法外，另案）不在其内
- 闸不辖：错误码/坐标/语义失实（归坐标类检查器）
- 已注册零载体残差（色名小数尾读数、复合前缀、时长误读形）按"载体先于扩展"家法挂账
- E07 现行命中由句尾叙事量词承载；其本体声明形（隔词闭合形）在现文法外，反制另案。
  复核（应显示 token='三类'）：
  `python3 -c "import json; d=json.load(open('results/d1-scores-v1.json')); print([s['claims'] for s in [g for g in d['gates'] if g['name']=='v2.7.2'][0]['segments'] if s['id']=='E07'][0])"`
