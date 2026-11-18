#!/usr/bin/env python3
"""D1 变异电池（review-loop 步骤 1，变异先行）。

对 runner/闸 逐发注入缺陷，跑全量探针矩阵：KILLED=探针红=守卫活；
SURVIVED=存活变异体=盲区坐标（按消费者门定级入台账，不默认进修复环）。
正副本 results/ 全程只读校验（跑完必须仍 OK——守恒守卫）。

跑法：python3 -I probes/run-mutants.py
"""
import os
import pathlib
import shutil
import subprocess
import sys

SB = pathlib.Path(__file__).resolve().parent.parent
RUNNER_SRC = (SB / "d1-runner-full.py").read_text(encoding="utf-8")
GATE_SRC = (SB / "gates" / "claim-gate-v2.7.1.py").read_text(encoding="utf-8")
OUT = SB / "_mut"

MUTANTS = [
    # id, needle, replacement, target
    ("A1-strip-off", "（事实[:：]", "（事实ZZZ", "runner"),
    ("A2-rule-mut", '"rule": rule,', '"rule": "MUT",', "runner"),
    ("A3-no-pathfilter", 'if not m or pathlib.Path(m.group("path")).name != TMP.name:', "if not m:", "runner"),
    ("A4-sha-1space", 'f"{digest}  {SCORES_NAME}\\n"', 'f"{digest} {SCORES_NAME}\\n"', "runner"),
    ("A5-d2-expect-flip", '"段D2": (False,', '"段D2": (True,', "runner"),
    ("A7-scores-rename", '"d1-scores-v1.json"', '"d1-scores-vX.json"', "runner"),
    ("A8-hits-oldmetric", "got = len(claims) > 0", 'got = stdout.count("count claim") > 0', "runner"),
    ("A6-gate-escape-revert", "alternative's " + chr(92) * 2 + "s*", "alternative's " + chr(92) + "s*", "gate"),
    # R1 外环审核位三条未跑成注入的补环（裁决 f-R1-a/b/c 盲区是否真盲）
    ("A9-recall-double", 'round(rec_, 4)', 'round(rec_ * 2, 4)', "runner"),
    ("A10-prec-round3", 'round(prec, 4)', 'round(prec, 3)', "runner"),
    ("A11-hits-inflate", 'claims.append({"line": lineno, "rule": rule, "token": token})',
     'claims.append({"line": lineno, "rule": rule, "token": token}); claims.append({"line": lineno, "rule": rule, "token": token})', "runner"),
    # R2 minor 修复的守恒变异：TMP mkdtemp 化 / OTHER 判定真源 / 派生 totals 三钉的存在性证明
    ("A12-tmp-moved", 'tempfile.mkdtemp(prefix="d1_seg_")) / "_seg_tmp.md"',
     'tempfile.mkdtemp(prefix="d1_seg_")) / "_moved_tmp.md"', "runner"),
]

if OUT.exists():
    shutil.rmtree(OUT)


def run_probe(env, timeout=600):
    return subprocess.run([sys.executable, "-I", str(SB / "probes" / "d1_probes.py")],
                          env=env, capture_output=True, text=True, cwd=str(SB), timeout=timeout)


rows = []
for mid, needle, repl, target in MUTANTS:
    src = RUNNER_SRC if target == "runner" else GATE_SRC
    n = src.count(needle)
    assert n == 1, f"{mid}: 锚出现 {n} 次（应恰 1）——变异定义失效"
    mdir = OUT / mid
    (mdir / "results").mkdir(parents=True)
    env = dict(os.environ)
    if target == "runner":
        p = mdir / "runner.py"
        p.write_text(src.replace(needle, repl, 1), encoding="utf-8")
        (mdir / "seeds").symlink_to(SB / "seeds")
        (mdir / "gates").symlink_to(SB / "gates")
        env.update(D1_RUNNER=str(p), D1_RESULTS=str(mdir / "results"))
    else:
        p = mdir / "gate.py"
        p.write_text(src.replace(needle, repl, 1), encoding="utf-8")
        env.update(D1_GATE_WARN=str(p), D1_RESULTS=str(mdir / "results"))
    r = run_probe(env)
    red = [l for l in r.stdout.splitlines() if l.startswith("FAIL")]
    verdict = "KILLED" if r.returncode != 0 or "PROBES" not in r.stdout else "SURVIVED"
    rows.append((mid, target, verdict, len(red), red[0][:66] if red else ""))
    print(f"{mid:<24} {target:<7} {verdict:<9} red面×{rows[-1][3]:<3} {rows[-1][4]}", flush=True)

gt = subprocess.run(["sha256sum", "-c", "d1-scores-v1.json.sha256"], cwd=str(SB / "results"),
                    capture_output=True, text=True)
print("\n正副本完好性: " + ("OK（未被电池污染）" if gt.returncode == 0 else "被污染! " + gt.stdout))
killed = sum(1 for x in rows if x[2] == "KILLED")
print(f"\n杀灭率 {killed}/{len(rows)}；存活：{[x[0] for x in rows if x[2] == 'SURVIVED'] or '无'}")
