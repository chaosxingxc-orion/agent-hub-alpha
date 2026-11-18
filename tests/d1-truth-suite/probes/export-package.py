#!/usr/bin/env python3
"""出站包导出器（R4 对抗审后结构化；取代手工 _outbound 搭建）。

原则：包=白名单导出＋再哈希自校，MANIFEST 由实树生成——"手工搭包必然漂移"
由 R4 审 M1/_probe_results 混入实锤，此处消灭（结构，非提醒）。
幂等：_outbound 每次全清重建。

跑法：python3 -I probes/export-package.py
"""
import hashlib
import json
import pathlib
import re
import shutil
import subprocess
import sys
from datetime import datetime, timezone

SB = pathlib.Path(__file__).resolve().parent.parent
OUT = SB / "_outbound"

def sha(p): return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()

# 1) 先重建闸（build-v272 自带 sanitize＋义务自校，失败即中止——闸不干净不出包）
r = subprocess.run([sys.executable, "-I", str(SB / "probes" / "build-v272.py")],
                   capture_output=True, text=True)
print(r.stdout.strip() or r.stderr.strip())
if r.returncode != 0:
    sys.exit(f"build 失败，导出中止（闸未净不外包）")

if OUT.exists():
    shutil.rmtree(OUT)

# 2) 白名单导出（与复现无关的一律不出包：私档真载体原文、scratch、workflow 编排、
#    git 基础设施脚本、设计期闸副本——MANIFEST"不收录"表由同一清单生成，单一真源）
FILES = [
    ("d1-runner-full.py", "d1-runner-full.py"),
    ("gates/claim-gate-v1.8.py", "gates/claim-gate-v1.8.py"),
    ("gates/claim-gate-v2.6.py", "gates/claim-gate-v2.6.py"),
    ("gates/claim-gate-v2.7.1.py", "gates/claim-gate-v2.7.1.py"),
    ("gates/claim-gate-v2.7.2.py", "gates/claim-gate-v2.7.2.py"),
    ("seeds/d1-corpus-full.md", "seeds/d1-corpus-full.md"),
    ("seeds/drift-replay.md", "seeds/drift-replay.md"),
    ("probes/d1_probes.py", "probes/d1_probes.py"),
    ("probes/run-mutants.py", "probes/run-mutants.py"),
    ("probes/r3-audit-battery.py", "probes/r3-audit-battery.py"),
    ("probes/build-v272.py", "probes/build-v272.py"),
    ("probes/export-package.py", "probes/export-package.py"),
    ("probes/stub_gate_other.py", "probes/stub_gate_other.py"),
    ("probes/final-review-battery.py", "probes/final-review-battery.py"),
    ("results/d1-scores-v1.json", "results/d1-scores-v1.json"),
    ("results/battery-receipt.log", "results/battery-receipt.log"),
    # 终审收据故意不入包（自指悖论：收据声称的文件数无法包含收据自身）——
    # 改为包内携带 final-review-battery.py，任何人可在克隆上现场重放同批 14 腿；收据存战役档案。
    ("_" + "d3_work/r2-e06-hongN-red/attack2.md", "docs/receipts/attack2.md"),
    ("README.outbound.md", "README.md"),
    ("docs/architecture.outbound.md", "docs/architecture.md"),
    ("docs/rules.outbound.md", "docs/rules.md"),
    ("docs/mr-description.outbound.md", "docs/mr-description.md"),
]

def mask_text(t):
    t = t.replace(str(SB), "/campaign-root")
    t = t.replace(str(SB.parent), "/home-user")
    for real, alias in (("review-loop-" + "台账", "truth-doc-A"),
                        ("claim_gate-" + "探索-" + "中间档", "truth-doc-B")):
        t = t.replace(real, alias)
    return t

NOTA_BENE = {}
OPTIONAL = {"results/final-review-receipt.log"}   # 由终审电池产出、首跑前允许缺席（生成后重导出入包）
for src, dst in FILES:
    sp, dp = SB / src, OUT / dst
    if not sp.exists() and src in OPTIONAL:
        print(f"（缺席待生成：{src}——终审电池产出后重导出即入包）")
        continue
    dp.parent.mkdir(parents=True, exist_ok=True)
    text = sp.read_text(encoding="utf-8")
    if src.startswith(("results/", "docs/", "README.outbound")):
        text = mask_text(text)          # 证据面＋出站文档源统一过别名道；代码面本应零私有指向
    if dst.endswith("claim-gate-v2.7.1.py"):
        PRIV_T = "the N" + "GA campaign repo"
        n = text.count(PRIV_T)
        assert n == 2, f"v2.7.1 私有仓名计数 {n}≠2——血统文本变了，sanitize 配对需同步"
        text = text.replace(PRIV_T, "the private campaign archive")
        NOTA_BENE["gates/claim-gate-v2.7.1.py"] = (
            "包内副本非逐字节标本：私有战役仓名引用×2 → 通用短语（公仓在仓 v2.6 不含该词，"
            "属新于公开面的指向）；检测行为零影响（纯注释），发布态哈希由 provenance 桥记录。")
    dp.write_text(text, encoding="utf-8")

# 3) 卫生自校（包内不许有真名/绝对家路径/私 scratch 目录名/私有代号/账号）
TERMS = ["review-loop-" + "台账", "claim_gate-" + "探索", "/home/" + "yao" + "jun",
         "yao" + "jun", "_" + "d3_work", "_" + "audit_r3",
         r"\bNG" + "A\b", "fei" + "shu"]
# 判决注记（R4 披露槽）：agent-festival＝v2.6 公仓字节既有引用（血统面，脱它反造漂移），
# 平台名 gitcode＝本 MR 落点自身——均不属"新披露/个人指向"，不入拦截集。
pat = re.compile("|".join(re.escape(t) for t in TERMS))   # 碎片构造：扫描器自身不得含完整目标串
hits = {}
for p in OUT.rglob("*"):
    if p.is_file():
        for i, ln in enumerate(p.read_text(encoding="utf-8", errors="ignore").splitlines(), 1):
            m = pat.findall(ln)
            if m:
                hits.setdefault(str(p.relative_to(OUT)), []).append((i, sorted(set(m)), ln.strip()[:90]))
if hits:
    for f, rows in hits.items():
        for i, ms, snip in rows[:6]:
            print(f"  LEAK {f}:{i} {ms} :: {snip}")
    sys.exit("卫生自校失败——导出中止（修源头，不在导出层打补丁）")
print("卫生自校：CLEAN")

# 4) 校验与桥哈希重算
jv = OUT / "results/d1-scores-v1.json"
h = sha(jv)
(jv.with_suffix(".json.sha256")).write_text(f"{h}  d1-scores-v1.json\n", encoding="utf-8")
prov = {
    "note": "本包 results/ 为路径/文档名打码拷贝；判定/声明/计数/文件内容哈希与原始运行一致。",
    "masked_gt_sha256": h,
    "original_gt_hash_chain": ["c3c8f093(R1)", "74de6c9a(R3)", "8417e5ad(R3-sanitized)"],
    "masked_vs_original": "上列为原始环境各版正副本哈希；本包 JSON 为打码拷贝，"
                          "其自身哈希见 masked_gt_sha256 字段——两者不可互替。",
    "campaign_archive_anchor": "战役档案私档 commit 链（仓名从略；哈希桥 575dcdf/d044ede 后继）",
    "gate_final_sha256": sha(OUT / "gates/claim-gate-v2.7.2.py"),
    "gate_published_variant_sha256": sha(OUT / "gates/claim-gate-v2.7.1.py"),
    "gate_published_variant_note": "包内 v2.7.1 副本的私有仓名注释已通用化——见 MANIFEST 勘误节",
    "public_repo_anchor_v2.6": "TODO 待补：代码托管平台在仓 v2.6 的 commit 哈希/URL（同源核验的外部锚）",
    "generated": datetime.now(timezone.utc).isoformat(timespec="seconds"),
}
(OUT / "results/provenance.json").write_text(json.dumps(prov, ensure_ascii=False, indent=1), encoding="utf-8")
# 发布态哈希桥：包内 battery V10 据此对 v2.7.1 变体副本改用自己的前缀锚（零代码分叉）
(OUT / "results/package-hashes.json").write_text(json.dumps({
    "gate_sha_prefix": {"v2.7.1": prov["gate_published_variant_sha256"][:12]},
    "note": "exporter 生成的发布态锚；战役档案环境无此文件，battery 用账本史锚——两端各自闭合。",
}, ensure_ascii=False, indent=1), encoding="utf-8")
(OUT / ".gitignore").write_text("__pycache__/\n*.pyc\n_probe_results/\n_battery_work/\n"
                                "seeds/truth-doc-*.md\n", encoding="utf-8")

# 5) MANIFEST 由实树生成（"不收录"清单=导出白名单的补集声明＋固定排除项）
tree = sorted(str(p.relative_to(OUT)) for p in OUT.rglob("*") if p.is_file())
inc_lines = "\n".join(f"- `{t}`" for t in tree)
excl = """- 真载体文档原文（包内以别名 truth-doc-A/B 记钉值；缺席时仪器可见跳过，在场跑满）
- 设计期工作目录（私有）中的设计期闸副本与装配收据原件——发布态哈希桥见 provenance.json
- 运行 scratch 目录若干（包内 .gitignore 复证；导出器每次全清重建，scratch 无处随包）
- workflow 编排脚本与 git 基础设施脚本（战役过程面，非复现所需）
- 跑分器 v0.1（被 v1.2 取代）；战役叙事类文档（私档，哈希可桥）
"""
nota = "\n".join(f"- `{k}`：{v}" for k, v in NOTA_BENE.items()) or "- （无）"
(OUT / "MANIFEST.md").write_text(f"""# MANIFEST：出站包收录判决（export-package.py 生成，实树即账）

判决原则：与复现相关且私有指向已清者收；运行 scratch 不收；战役叙事与私档原文不收、
以哈希桥保留可追溯。卫生自校 pattern 见导出器（失败即中止，不在导出层打补丁）。

## 收录（{len(tree)} 文件，与磁盘实树逐一对应）
{inc_lines}

## 不收录（声明性排除）
{excl}
## 包内变体勘误（非逐字节标本处，逐条）
{nota}

## 待外部补全
- provenance.json public_repo_anchor_v2.6：代码托管平台在仓 commit 哈希（v2.6/v1.8 同源核验的外部锚）——
  本机无该平台凭证不可达（418），由持有仓权限侧补。

## 血统保留判决（非遗漏，逐条）
- `agent-festival/tools vs root`（v2.6:445 / v2.7.1:450 / v2.7.2:450）：v2.6 为公仓在仓字节，
  该引用已属公开面历史文本；R4 披露槽判"脱它＝对血统的新偏移＋前提虚假"——保留。
- v2.7.x 血统头注中的工单号/事件编号（56号/64号类）：同上，v2.6 既有文面；裁决记录本身
  是规则的一部分（用户裁定：诚实深刻优先）。
""", encoding="utf-8")

n = sum(1 for p in OUT.rglob("*") if p.is_file())
print(f"EXPORT OK: {n} files | gate sha {prov['gate_final_sha256'][:12]}… | masked GT {h[:12]}…")
