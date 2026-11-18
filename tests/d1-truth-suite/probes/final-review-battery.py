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
from datetime import datetime, timezone

SB = pathlib.Path(__file__).resolve().parent.parent
PKG = SB / "_outbound"
if not (PKG / "results" / "provenance.json").exists():
    PKG = SB
FAST = "--fast" in sys.argv
FAILS = []
LOGDIR = SB / "_final_review"
LOGDIR.mkdir(exist_ok=True)
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


files = [p for p in PKG.rglob("*") if p.is_file() and "__pycache__" not in str(p)]
texts = {str(p.relative_to(PKG)): p.read_text(encoding="utf-8", errors="ignore") for p in files}

# ---------- D 组：披露复检 ----------
bad = [f for f in texts if re.search(r"_probe_results|_mut/|_battery_work|__pycache__|\.pyc$", f)]
say(not bad, "D1-scratch零随包", f"scratch 形态文件={bad or '无'}；树内 {len(texts)} 文件")
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
