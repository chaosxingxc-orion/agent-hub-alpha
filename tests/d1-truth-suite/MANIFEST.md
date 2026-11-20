# MANIFEST：出站包收录判决（export-package.py 生成，实树即账）

判决原则：与复现相关且私有指向已清者收；运行 scratch 不收；战役叙事与私档原文不收、
以哈希桥保留可追溯。卫生自校 pattern 见导出器（失败即中止，不在导出层打补丁）。

## 收录（25 文件，与磁盘实树逐一对应）
- `.gitignore`
- `README.md`
- `d1-runner-full.py`
- `docs/architecture.md`
- `docs/mr-description.md`
- `docs/receipts/attack2.md`
- `docs/rules.md`
- `gates/claim-gate-v1.8.py`
- `gates/claim-gate-v2.6.py`
- `gates/claim-gate-v2.7.1.py`
- `gates/claim-gate-v2.7.2.py`
- `probes/build-v272.py`
- `probes/d1_probes.py`
- `probes/export-package.py`
- `probes/final-review-battery.py`
- `probes/r3-audit-battery.py`
- `probes/run-mutants.py`
- `probes/stub_gate_other.py`
- `results/battery-receipt.log`
- `results/d1-scores-v1.json`
- `results/d1-scores-v1.json.sha256`
- `results/package-hashes.json`
- `results/provenance.json`
- `seeds/d1-corpus-full.md`
- `seeds/drift-replay.md`

## 不收录（声明性排除）
- 真载体文档原文（包内以别名 truth-doc-A/B 记钉值；缺席时仪器可见跳过，在场跑满）
- 设计期工作目录（私有）中的设计期闸副本与装配收据原件——发布态哈希桥见 provenance.json
- 运行 scratch 目录若干（包内 .gitignore 复证；导出器每次全清重建，scratch 无处随包）
- workflow 编排脚本与 git 基础设施脚本（战役过程面，非复现所需）
- 跑分器 v0.1（被 v1.2 取代）；战役叙事类文档（私档，哈希可桥）

## 包内变体勘误（非逐字节标本处，逐条）
- `gates/claim-gate-v2.7.1.py`：包内副本非逐字节标本：私有战役仓名引用×2 → 通用短语（公仓在仓 v2.6 不含该词，属新于公开面的指向）；检测行为零影响（纯注释），发布态哈希由 provenance 桥记录。

## 待外部补全
- provenance.json public_repo_anchor_v2.6：代码托管平台在仓 commit 哈希（v2.6/v1.8 同源核验的外部锚）——
  本机无该平台凭证不可达（418），由持有仓权限侧补。

## 血统保留判决（非遗漏，逐条）
- `agent-festival/tools vs root`（v2.6:445 / v2.7.1:450 / v2.7.2:450）：v2.6 为公仓在仓字节，
  该引用已属公开面历史文本；R4 披露槽判"脱它＝对血统的新偏移＋前提虚假"——保留。
- v2.7.x 血统头注中的工单号/事件编号（56号/64号类）：同上，v2.6 既有文面；裁决记录本身
  是规则的一部分（用户裁定：诚实深刻优先）。
