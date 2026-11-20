#!/usr/bin/env python3
"""出厂终审电池：对本包做披露卫生（D 组）、文档承诺复现（R 组）、措辞机检（W 组）共 14 腿。

维护者可对自己的克隆直接复跑——全部判据只读本包与自建 scratch，无外部依赖。
--fast 跳过最慢的 R5-R7（包内全探针/全电池/A 系列）；默认全跑。
跑法：python3 -I probes/final-review-battery.py [--fast]
"""
import hashlib
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile
from datetime import datetime, timezone

SB = pathlib.Path(__file__).resolve().parent.parent
PKG = SB / "_outbound"
if not (PKG / "results" / "provenance.json").exists():
    PKG = SB
FAST = "--fast" in sys.argv
FAILS = []
# m1 同族修复（148号复审 2026-11-18）：收据日志默认写系统临时目录，不落包内——
# 否则包形态直跑时 D4「MANIFEST 账实等值」会把电池自己的日志数成第 26 个文件。
# 战役归档需要持久收据时用 D1_REVIEW_LOG 显式指定目录。
LOGDIR = pathlib.Path(os.environ.get(
    "D1_REVIEW_LOG", pathlib.Path(tempfile.gettempdir()) / "_d1_final_review"))
LOGDIR.mkdir(parents=True, exist_ok=True)
LOG = open(LOGDIR / "final-receipt.log", "a", encoding="utf-8")
LOG.write(f"\n===== final-review-battery {'FAST' if FAST else 'FULL'} @ "
          f"{datetime.now(timezone.utc).isoformat(timespec='seconds')}\n")


def say(ok, vid, msg):
    line = ("PASS " if ok else "FAIL ") + f"[{vid}] {msg}"
    print(line)
    LOG.write(line + "\n")
    LOG.flush()
    if not ok:
        FAILS.append(vid)


def sha(p):
    return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()[:12]


def run(cmd, cwd=PKG, timeout=900):
    return subprocess.run(cmd, shell=isinstance(cmd, str), capture_output=True,
                          text=True, cwd=str(cwd), timeout=timeout,
                          env=dict(os.environ, D1_RESULTS=str(LOGDIR / "res")))


# M-a 修复（终轮补跑 CONFIRMED MAJOR，2026-11-19）：scratch 名单单源。
# 包文档化的快速复现（README:14/16/17）会生成 _probe_results/ _mut/ _battery_work/，
# 旧 D1 正则与 D4 全树对账把它们判为「随包污染」——假红拒绝合法工作树。
# 名单与 .gitignore 保持同源（此处集中一份，三处使用：files 过滤、D1 正则、D7 copytree）。
SCRATCH_DIRS = ("_probe_results/", "_mut/", "_mut_r2/", "_battery_work/", "_final_review/")
SCRATCH_NAMES = ("_probe_results", "_mut", "_mut_r2", "_battery_work", "_final_review")


def _is_scratch(rel: str) -> bool:
    return (rel.startswith(SCRATCH_DIRS) or "__pycache__/" in (rel + "/")
            or rel.endswith(".pyc") or rel.startswith("seeds/truth-doc-"))


# 文件域＝**被跟踪的文件**（终审 M-1 修复，2026-11-19）：发布物 = git tracked 集合。
# 旧实现把 scratch 从全树里过滤后，D1 的断言按构造恒为空（恒真探针），且 D2/D3/W1
# 共用同一 filtered 集合 → 对 scratch 路径下的私有指向/代号/谱系字样全部失明
# （实测：在 scratch 目录写入含本地账号路径的内容 → 9/9 全绿）。改为 tracked 域后：
# scratch 天然不在域内（D1 语义「发布物不含 scratch」成立且非恒真——被跟踪的 scratch
# 仍会红），而 D2/D3/W1 对 tracked 全量扫描、不失明。
_tracked = subprocess.run(["git", "-C", str(PKG), "ls-files"],
                          capture_output=True, text=True).stdout.split()
_has_tracked = bool(_tracked)
if _has_tracked:
    # 发布物域＝tracked（scratch untracked → 天然域外，不发布）
    files = [PKG / rel for rel in _tracked if (PKG / rel).is_file()]
else:
    # 未提交工作树：无发布物域可判。域＝全树减 scratch（scratch 不发布）。
    files = [p for p in PKG.rglob("*")
             if p.is_file() and not _is_scratch(str(p.relative_to(PKG)))]
texts = {str(p.relative_to(PKG)): p.read_text(encoding="utf-8", errors="ignore") for p in files}

# ---------- D 组：披露复检 ----------
# D1 在 tracked 域上仍会抓「被跟踪的 scratch」（真发布污染）；scratch 未被跟踪则不在域内。
bad = [f for f in texts if _is_scratch(f)]
say(not bad, "D1-scratch零随包",
    (f"scratch 形态文件={bad or '无'}；tracked 域 {len(texts)} 文件" if _has_tracked
     else f"工作树未提交（无 tracked 域）——D1 按设计无从判，域={len(texts)} 文件（已减 scratch）"))
dang = [str(p.relative_to(PKG)) for p in PKG.rglob("*") if p.is_symlink() and not p.exists()]
say(not dang, "D1b-悬空软链", f"悬空链接={dang or '无'}")

LIT = [r"battle" + r"-ledger", "ledger:1" + "15", "_" + "audit_r3",
       "review-loop-" + "台账", "探索-" + "中间档", "yao" + "jun"]
LEAK_PAT = "|".join([re.escape(x) for x in LIT] + [r"/home/[A-Za-z]"])  # 泛路径 /home/ 不算泄漏，须跟账号名
leaks = {}
for f, t in texts.items():
    hits = [ln.strip()[:90] for ln in t.splitlines() if re.search(LEAK_PAT, ln)]
    if hits:
        leaks[f] = hits[:3]
say(not leaks, "D2-私有指向零", f"leaks={json.dumps(leaks, ensure_ascii=False)[:300] or '无'}")
ml_docs = {f: t.count("minor-ledger:") for f, t in texts.items()
           if not f.startswith(("gates/", "probes/")) and "minor-ledger:" in t}
# probes 侧白名单清洗正则与 gates 血统文本同性质（功能必需字面量），限 md 面为零
ml_gates = sum(t.count("minor-ledger:") for f, t in texts.items() if f.startswith("gates/"))
say(not ml_docs and ml_gates > 0, "D2b-minor-ledger限闸体",
    f"白名单引用只许在 gates/ 血统文本（实得 gates={ml_gates}，docs 泄漏={ml_docs or '无'}）")

CODE_T = "NG" + "A"   # 碎片构造：本文件随包，代号字面量不得完整（变量名同样数）
nga = {f: t.count(CODE_T) for f, t in texts.items() if CODE_T in t}
say(not nga, "D3-私有代号零", f"完整代号出现={nga or '无'}（碎片构造形态不计）")
af = [f for f in ("gates/claim-gate-v2.6.py", "gates/claim-gate-v2.7.1.py", "gates/claim-gate-v2.7.2.py")
      if "agent-festival" in texts.get(f, "")]
say(len(af) == 3 and "agent-festival" in texts.get("MANIFEST.md", ""),
    "D6-agent-festival判决自洽",
    f"三包文件含该血统引用={len(af)}/3；MANIFEST 判决节在场={'agent-festival' in texts.get('MANIFEST.md','')}")

mf = texts.get("MANIFEST.md", "")
listed = set(re.findall(r"^- `([^`]+)`", mf.split("## 不收录")[0], re.M))
actual = set(texts) - {"MANIFEST.md"}
say(listed == actual, "D4-MANIFEST账实等值",
    f"清单{len(listed)} vs 实树{len(actual)}，差={sorted(listed ^ actual)[:6]}")

# ---- D7（review-loop 终轮 bearing-2，2026-11-19；R7 修复：unlink+负对照臂防恒真）----
# build-v272 声称"包形态可字节同重建"——真验＝复制包到 tempdir、**先删产物**、
# 跑装配器、断言重建文件存在且与出站闸逐字节同。R6 实弹：不删产物时 no-op
# builder 也 PASS（copytree 把旧产物带进临时树）——"幂等 281f99aa"声称当时
# 无任何随包探针在锁。双臂：负对照臂注入 no-op builder，同一判据必须红得出
# （守恒钉内建——删掉 unlink 或弱化判据，负臂即转绿打红本腿）。
import shutil as _sh
_ref = hashlib.sha256((PKG / "gates" / "claim-gate-v2.7.2.py").read_bytes()).hexdigest()


def _d7_arm(builder_text):
    """单臂：临时树里删产物后跑（可选注入的）装配器，返回 (判据, 完成进程)。"""
    _w = pathlib.Path(tempfile.mkdtemp(prefix="d7-"))
    try:
        _pkg = _w / "pkg"
        _sh.copytree(PKG, _pkg, ignore=_sh.ignore_patterns(
            "__pycache__", "_probe_results", "_mut", "_mut_r2", "_battery_work", "_final_review"))
        _out = _pkg / "gates" / "claim-gate-v2.7.2.py"
        _out.unlink()                            # 关键：产物先删，逼装配器真重建
        if builder_text is not None:             # 负对照臂：换 no-op 装配器
            (_pkg / "probes" / "build-v272.py").write_text(builder_text, encoding="utf-8")
        _rb = subprocess.run([sys.executable, "-I", "probes/build-v272.py"],
                             cwd=str(_pkg), capture_output=True, text=True, timeout=300)
        return (_rb.returncode == 0 and _out.exists()
                and hashlib.sha256(_out.read_bytes()).hexdigest() == _ref), _rb
    finally:
        _sh.rmtree(_w, ignore_errors=True)


_ok_pos, _rb_pos = _d7_arm(None)
_ok_neg, _rb_neg = _d7_arm('print("no-op negative control")\nimport sys\nsys.exit(0)\n')
# 错内容负臂（R7 内环 F1 修复）：产物存在、rc=0、但 sha≠原件——锁「重建等值」谓词；
# 弱化判据（删 sha 比较）时 no-op 臂转绿抓不住，此臂专堵该缝（同判据必须红）。
_ok_bad, _rb_bad = _d7_arm(
    'import pathlib, sys\n'
    'p = pathlib.Path("gates/claim-gate-v2.7.2.py")\n'
    'p.parent.mkdir(parents=True, exist_ok=True)\n'
    'p.write_bytes(b"corrupted negative control payload\\n")\n'
    'print("corrupt negative control")\nsys.exit(0)\n')
say(_ok_pos and not _ok_neg and not _ok_bad
    # 二轮勘误（r3 家法「崩溃不是红」同律）：两支负 builder 必须**成功跑完**（rc=0），
    # 其红才算有效红——负臂 builder 源码被碎（rc≠0）时判据同样 False，无此哨兵会
    # 把崩溃当有效红，F1 弱化+碎臂组合即可无声失活本腿。
    and _rb_neg.returncode == 0 and _rb_bad.returncode == 0, "D7-装配器幂等",
    f"正臂 rc={_rb_pos.returncode} 重建等值={_ok_pos}；负臂 no-op 判红={not _ok_neg}（rc={_rb_neg.returncode}）；"
    f"负臂错内容判红={not _ok_bad}（rc={_rb_bad.returncode}）")

ph = json.loads((PKG / "results/package-hashes.json").read_text()) if (PKG / "results/package-hashes.json").exists() else {}
pv = json.loads((PKG / "results/provenance.json").read_text())
ok = (ph.get("gate_sha_prefix", {}).get("v2.7.1") == sha(PKG / "gates/claim-gate-v2.7.1.py")
      and pv.get("gate_final_sha256", "").startswith(sha(PKG / "gates/claim-gate-v2.7.2.py"))
      and sha(PKG / "results/d1-scores-v1.json") == pv.get("masked_gt_sha256", "")[:12])
r = run(["sha256sum", "-c", "d1-scores-v1.json.sha256"], cwd=PKG / "results")
say(ok and r.returncode == 0, "D5-桥哈希三向闭",
    f"包hash↔盘={ph.get('gate_sha_prefix')} vs {sha(PKG/'gates/claim-gate-v2.7.1.py')}；-c rc={r.returncode}")

# ---------- R 组：包内复现 ----------
doc = json.loads((PKG / "results/d1-scores-v1.json").read_text())
g = {x["name"]: x for x in doc["gates"]}
if len(doc["gates"]) != 4:
    say(False, "R1-48格与F1表", f"缺门：gates 现值={sorted(g)}（期望 4 门）")
    say(False, "R2-E07命令语义", "缺门，无法检查")
else:
    ok = (all(len(x["segments"]) == 12 for x in doc["gates"])
          and [g[n]["totals"]["f1"] for n in ("v1.8", "v2.6", "v2.7.1", "v2.7.2")] == [0.25, 0.4, 0.8333, 1.0]
          and g["v2.7.2"]["totals"] == {"tp": 6, "fp": 0, "tn": 6, "fn": 0,
                                        "precision": 1.0, "recall": 1.0, "f1": 1.0})
    say(ok, "R1-48格与F1表", f"v2.7.2 totals={g['v2.7.2']['totals']}")
    e07 = [s for s in g["v2.7.2"]["segments"] if s["id"] == "E07"][0]
    say(any(c["token"] == "三类" for c in e07["claims"]), "R2-E07命令语义", f"claims={e07['claims']}")
say(run("python3 -I gates/claim-gate-v2.7.2.py --selftest").returncode == 0, "R3-selftest", "rc=0")
say(run("python3 -I gates/claim-gate-v2.7.2.py --census").returncode == 0, "R3b-census", "rc=0")
cnt = {f: sum(1 for l in run(f'python3 -I gates/claim-gate-v2.7.2.py --check "{f}"').stdout.splitlines()
              if "count claim" in l)
       for f in ("README.md", "MANIFEST.md", "docs/architecture.md", "docs/rules.md", "docs/mr-description.md")}
say(all(v == 0 for v in cnt.values()), "R4-五文档过闸", str(cnt))
if not FAST:
    pr = run("python3 -I probes/d1_probes.py", timeout=900).stdout
    say("fails=0" in pr and pr.count("SKIP [FAM-REALDOCS]") == 2, "R5-包内探针",
        [l for l in pr.splitlines() if "PROBES" in l][-1:] or "无行")
    bt = run("python3 -I probes/r3-audit-battery.py", timeout=900).stdout
    say("11/11 PASS" in bt and "可见跳过" in bt, "R6-包内电池",
        [l for l in bt.splitlines() if "BATTERY" in l][-1:] or "无行")
    am = run("python3 -I probes/run-mutants.py", timeout=3600).stdout
    say("杀灭率 12/12" in am, "R7-A系列12发",
        [l for l in am.splitlines() if "杀灭率" in l][-1:] or "无行")

# ---------- W 组：措辞机检 ----------
W1_PAT = r"第[5五]次发行|版本" + "谱系"   # 词表自身也碎片化（扫描器词表命中自己=第四次同族变体）
w1 = {f: [l[:60] for l in t.splitlines() if re.search(W1_PAT, l)] for f, t in texts.items()}
w1 = {f: v for f, v in w1.items() if v}
say(not w1, "W1-包外谱系字样零", f"hits={list(w1)[:3] or '无'}")
chain = json.dumps([pv.get("original_gt_hash_chain", ""), pv.get("masked_vs_original", "")],
                   ensure_ascii=False)
say("masked_gt_sha256" in chain and "当前=本文件 masked 值" not in chain,
    "W2-哈希链话术", f"chain={chain[:180]}")
rdm = texts.get("README.md", "")
say("跳过" in rdm and "不放水" in rdm and "可见且不计" in rdm, "W3-跳过语义诚实", "README 限定句在场" if "可见且不计" in rdm else "缺『跳过可见且不计』限定")
say(all("可见跳过" in texts.get(f, "") for f in ("README.md", "docs/mr-description.md")),
    "W4-双态口径一致", "README 与 MR 均含包形态跳过口径")

total = 14 if not FAST else 9
print(f"\nR5-FINAL-BATTERY: {total - len(FAILS)}/{total} PASS" + (f" | FAILS={FAILS}" if FAILS else " — 全绿"))
LOG.write(f"SUMMARY: {FAILS or '全绿'}\n")
sys.exit(1 if FAILS else 0)
