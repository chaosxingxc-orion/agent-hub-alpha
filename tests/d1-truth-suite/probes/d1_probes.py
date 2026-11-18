#!/usr/bin/env python3
"""D1 v1.2 探针矩阵（review-loop R1–R3；四代×48 格；v1.1=R2 minor 修复轮）——确定性、仅 stdlib。

跑法：python3 -I probes/d1_probes.py
默认探针跑动写 _probe_results/，**绝不触碰 results/ 正副本**（守恒读对象）。
env 注入点（变异电池用）：
  D1_RUNNER=<runner.py>   被测跑分器副本（默认 d1-runner-full.py）
  D1_RESULTS=<dir>        跑分结果目录（默认 _probe_results）
  D1_GATE_WARN=<gate.py>  警告门目标闸（默认 gates/claim-gate-v2.7.1.py）

族键与战役台账（campaign ledger）对齐（FAM-*）。任一 FAIL＝红＝内环评审不开。
矩阵值来源：2026-11-15 R1 前实测跑分（预注册 36/36 命中后的冻结）＋
真载体实测（truth-doc-A 上 v1.8=3 命中与战役预审记录逐字一致；其余当轮跑定）；
R3（2026-11-16/17）追加 v2.7.2 列 12 格、CLAIMS 7 枚、REALDOCS 26/16、L182 四代——
均经 D3 设计/红队与 R3 审计电池多重实跑后冻结。
"""
import hashlib
import importlib.util
import json
import os
import pathlib
import re
import subprocess
import sys

SB = pathlib.Path(__file__).resolve().parent.parent
RUNNER = pathlib.Path(os.environ.get("D1_RUNNER", SB / "d1-runner-full.py"))
RESULTS = pathlib.Path(os.environ.get("D1_RESULTS", SB / "_probe_results"))
GT_RESULTS = SB / "results"
GATE_WARN = pathlib.Path(os.environ.get("D1_GATE_WARN", SB / "gates" / "claim-gate-v2.7.1.py"))
SCORES = "d1-scores-v1.json"
GATE_V271 = SB / "gates" / "claim-gate-v2.7.1.py"
TMPX = str(SB / "_seg_tmp.md")

MATRIX = {
    "v1.8":   {"E01": "TN", "E02": "TN", "E03": "TN", "E04": "FN", "E05": "TN",
               "E06": "FN", "E07": "FN", "E08": "FN", "E09": "TP", "E10": "FN",
               "段C": "TN", "段D2": "FP"},
    "v2.6":   {"E01": "FP", "E02": "TN", "E03": "TN", "E04": "FN", "E05": "TN",
               "E06": "FN", "E07": "TP", "E08": "FN", "E09": "TP", "E10": "FN",
               "段C": "TN", "段D2": "FP"},
    "v2.7.1": {"E01": "FP", "E02": "TN", "E03": "TN", "E04": "TP", "E05": "TN",
               "E06": "FN", "E07": "TP", "E08": "TP", "E09": "TP", "E10": "TP",
               "段C": "TN", "段D2": "TN"},
    "v2.7.2": {"E01": "TN", "E02": "TN", "E03": "TN", "E04": "TP", "E05": "TN",
               "E06": "TP", "E07": "TP", "E08": "TP", "E09": "TP", "E10": "TP",
               "段C": "TN", "段D2": "TN"},  # R3 收编列：E01 闭(skip)＋E06 闭(REDCOUNT)
}
def metrics_from(cells):
    """FAM-METRIC-UNPIN 换结构钉（R2，族×3 存活变异体触发）：混淆计数与三个比率
    全部由探针从 MATRIX verdict 独立重算（仪器侧锚——判别参考不由 SUT 自产），
    字段集合相等校验让未来新增 totals 字段无法静默逃逸出钉面。"""
    c = {k: sum(1 for v in cells.values() if v == k) for k in ("TP", "FP", "TN", "FN")}
    tp, fp, fn = c["TP"], c["FP"], c["FN"]
    prec = tp / (tp + fp) if tp + fp else 0.0
    rec = tp / (tp + fn) if tp + fn else 0.0
    return {"tp": tp, "fp": fp, "tn": c["TN"], "fn": fn,
            "precision": round(prec, 4), "recall": round(rec, 4),
            "f1": round(2 * prec * rec / (prec + rec), 4) if prec + rec else 0.0}
# (gate, seg, rule, token)——声明级冻结
CLAIMS = [
    ("v2.7.1", "E04", "COUNTING/样本", "12 样本"),
    ("v2.6", "E07", "COUNTING/类", "三类"),
    ("v2.7.1", "E07", "COUNTING/类", "三类"),
    ("v2.7.1", "E08", "PROGRESS", "3/5 已改"),
    ("v1.8", "E09", "COUNTING/个", "9 个"),
    ("v2.6", "E09", "COUNTING/个", "9 个"),
    ("v2.7.1", "E09", "COUNTING/个", "9 个"),
    ("v2.6", "E01", "COUNTING/个", "两个"),
    ("v2.7.1", "E01", "COUNTING/个", "两个"),
    ("v2.7.1", "E10", "COUNTING/用例", " 1826 用例"),
    ("v1.8", "段D2", "COUNTING/个", " 17 个"),
    ("v1.8", "段D2", "COUNTING/个", " 4 个"),
    ("v2.6", "段D2", "COUNTING/个", " 17 个"),
    ("v2.6", "段D2", "COUNTING/个", " 4 个"),
    # R3 收编列 v2.7.2 的声明级冻结（REDCOUNT 新档＋继承面逐点钉）
    ("v2.7.2", "E06", "REDCOUNT", "红 5"),
    ("v2.7.2", "E10", "REDCOUNT", "红 4"),
    ("v2.7.2", "E10", "COUNTING/用例", " 1826 用例"),
    ("v2.7.2", "E04", "COUNTING/样本", "12 样本"),
    ("v2.7.2", "E07", "COUNTING/类", "三类"),
    ("v2.7.2", "E08", "PROGRESS", "3/5 已改"),
    ("v2.7.2", "E09", "COUNTING/个", "9 个"),
]
DROPS = {"E02": 1, "E03": 1, "E04": 1, "E05": 1, "E06": 1, "E07": 1, "E08": 1, "E09": 1}
# truth-doc-A/B＝战役档案两份真载体的包内别名（战役树以 seeds/ 符号链接落地，
# 公开包不含真名文件——缺席时本检查可见跳过，见下）。
REALDOCS = {
    "seeds/truth-doc-A.md": {"v1.8": 3, "v2.6": 23, "v2.7.1": 22, "v2.7.2": 26},
    "seeds/truth-doc-B.md": {"v1.8": 7, "v2.6": 13, "v2.7.1": 15, "v2.7.2": 16},
}
# truth-doc-A 的 `ls gate-attack-samples/` 载体行（簇①真载体）。True=该行有 count claim。
# 四代值均 R1-R3 实测钉死（v2.6 于 2026-11-15 实测 True 转正；v2.7.2 经 V7/内环双证 False）。
L182 = {"v1.8": True, "v2.6": True, "v2.7.1": False, "v2.7.2": False}

N_CHECK = 0
FAILS = []


def check(fam, cond, msg):
    global N_CHECK
    N_CHECK += 1
    tag = "PASS" if cond else "FAIL"
    print(f"{tag} [{fam}] {msg}")
    if not cond:
        FAILS.append((fam, msg))


def load(name, path):
    spec = importlib.util.spec_from_file_location(name, path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def gate_check(gate_path, target):
    r = subprocess.run([sys.executable, "-I", str(gate_path), "--check", str(target)],
                       capture_output=True, text=True, timeout=30)
    return r.stdout


def sha(p):
    return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()


def check_persist(fam, d):
    j, c = d / SCORES, d / (SCORES + ".sha256")
    ok = j.exists() and c.exists()
    check(fam, ok, f"{d.name}/ 两文件存在")
    if not ok:
        return
    txt = c.read_text(encoding="utf-8")
    check(fam, re.fullmatch(r"[0-9a-f]{64}  " + re.escape(SCORES) + r"\n", txt) is not None,
          "校验行格式=64hex+双空格+名+\\n（sha256sum -c 兼容）")
    check(fam, txt.split("  ")[0] == sha(j), "digest 与本体一致")
    r = subprocess.run(["sha256sum", "-c", SCORES + ".sha256"], cwd=str(d),
                       capture_output=True, text=True)
    check(fam, r.returncode == 0, f"sha256sum -c rc=0（{r.stdout.strip()[:40]}）")


def main():
    # —— 跑被测 runner（隔离输出）——
    env = dict(os.environ, D1_RESULTS=str(RESULTS))
    r = subprocess.run([sys.executable, "-I", str(RUNNER)], env=env,
                       capture_output=True, text=True, cwd=str(SB), timeout=300)
    check("FAM-RUN", r.returncode == 0, f"runner rc=0（mut={RUNNER.name}）")
    check("FAM-RUN", r.stderr == "", f"runner stderr 空（实得 {r.stderr[:60]!r}）")
    try:
        doc = json.loads((RESULTS / SCORES).read_text(encoding="utf-8"))
    except Exception as e:
        check("FAM-RUN", False, f"结果 JSON 不可读: {e}")
        return finish()
    gates = {g["name"]: g for g in doc["gates"]}

    # —— FAM-MATRIX 48 格（四代）+ 总量（派生锚 metrics_from）——
    for gname, cells in MATRIX.items():
        segs = {s["id"]: s for s in gates.get(gname, {}).get("segments", [])}
        check("FAM-MATRIX", set(segs) == set(cells), f"{gname} 段集恰 12")
        for sid, want in cells.items():
            got = segs.get(sid, {}).get("verdict")
            check("FAM-MATRIX", got == want, f"{gname}/{sid}={want}")
        t = gates.get(gname, {}).get("totals", {})
        want = metrics_from(cells)
        check("FAM-TOTALS", set(t) == set(want), f"{gname} totals 字段集恰等={sorted(want)}（新字段红）")
        for k, v in sorted(want.items()):
            check("FAM-TOTALS", t.get(k) == v,
                  f"{gname}.totals.{k}=={v}（实得 {t.get(k)}，仪器侧锚派生）")

    # —— FAM-ATTRIB 声明级 token ——
    for gname, sid, rule, tok in CLAIMS:
        found = any(c["rule"] == rule and c["token"] == tok
                    for s in gates.get(gname, {}).get("segments", []) if s["id"] == sid
                    for c in s["claims"])
        check("FAM-ATTRIB", found, f"{gname}/{sid} 含 [{rule}] {tok!r}")

    # —— FAM-STRIP 注记剥离 ——
    for s in gates["v2.7.1"]["segments"]:
        check("FAM-STRIP", s["annotations_dropped"] == DROPS.get(s["id"], 0),
              f"{s['id']} dropped={DROPS.get(s['id'], 0)}")

    # —— FAM-NO-SILENT 静默通道 ——
    all_segs = [s for g in gates.values() for s in g.get("segments", [])]
    check("FAM-NO-SILENT", all(s["verdict"] != "UNREGISTERED" for s in all_segs), "无 UNREGISTERED")
    all_claims = [c for s in all_segs for c in s["claims"]]
    check("FAM-NO-SILENT", not any(c["rule"].startswith("UNATTRIB") for c in all_claims), "无 UNATTRIBUTED")
    check("FAM-NO-SILENT", not any(c["rule"] == "COUNTING/?" for c in all_claims),
          "无兜底漂移值 COUNTING/?（m-R1-4 修复：哨兵覆盖兜底位）")
    check("FAM-NO-SILENT", not any(c["rule"] == "OTHER" for c in all_claims),
          "现语料零 OTHER 通道（新 finding 形出现即红→必须复审归属）")

    # —— FAM-CORPUS ——
    rmod = load("d1run_under_probe", RUNNER)
    segs_list = list(rmod.segments())
    check("FAM-CORPUS", len(segs_list) == 12, f"段数=12（实得 {len(segs_list)}）")
    check("FAM-CORPUS", set(i for i, *_ in segs_list) == set(rmod.EXPECT), "段 ID 与 EXPECT 全交")
    check("FAM-CORPUS", all(not l.lstrip().startswith("（事实") for _, _, b, _ in segs_list for l in b),
          "喂闸正文零注记残留")
    d2 = [b for i, _, b, _ in segs_list if i == "段D2"][0]
    joined = "\n".join(d2)
    check("FAM-CORPUS", "`ls " in joined and "`find " in joined, "段D2 含 ls/find 反引号证据形")

    # —— FAM-UNITS 单元（合成 stdout 直接打 parse_claims/ANNOT_RE）——
    gmod = load("d1gate_v271", GATE_V271)
    body = ["9 个回归样本"]
    cl = rmod.parse_claims(f"{TMPX}:1: count claim '9 个' without a command in the ±8-line window: x", gmod, body)
    check("FAM-UNITS", cl == [{"line": 1, "rule": "COUNTING/个", "token": "9 个"}], f"归属单元 {cl}")
    cl2 = rmod.parse_claims(f"{TMPX}:2: ledger entry: weird finding", gmod, body)
    check("FAM-UNITS", len(cl2) == 1 and cl2[0]["rule"] == "OTHER", "非 count-claim 形→OTHER 不静默")
    cl3 = rmod.parse_claims("/other/notes.md:1: count claim '9 个' without a command in the ±8-line window: x", gmod, body)
    check("FAM-UNITS", cl3 == [], "异路径 finding 行被滤（防窗口外污染）")
    ar = rmod.ANNOT_RE
    check("FAM-UNITS", ar.match("（事实：x）") and ar.match("（事实: x）") and ar.match("  （事实：x）"),
          "ANNOT 全/半角冒号+前导空格收")
    check("FAM-UNITS", not ar.match("（注记：x）") and not ar.match("事实：x"), "ANNOT 不收相邻形（防过度剥离）")
    cl4 = rmod.parse_claims(f"{TMPX}:1: count claim '9 节' without a command in the ±8-line window: 9 节",
                            gmod, ["9 节"])
    check("FAM-UNITS", cl4 and cl4[0]["rule"] == "COUNTING/?",
          f"兜底位可达性钉（单位表漂移→/? 非静默）实得={cl4 and cl4[0]['rule']}")
    check("FAM-TMP", rmod.TMP.name == "_seg_tmp.md" and rmod.TMP.parent.name.startswith("d1_seg_"),
          f"TMP mkdtemp 每运行隔离（m-R1-6 修复：树脏/并发覆写灭）实得={rmod.TMP.parent.name[:12]}")
    # FAM-OTHERVER（m-R1-1/A8 杀钉）：stub 闸只产 OTHER 形 finding——
    # 判定真源必须是 parse_claims 产物；got 退回 stdout 计数字符串口径的变异在此现红。
    srec = rmod.score_gate("stub", SB / "probes" / "stub_gate_other.py")
    check("FAM-OTHERVER", srec.get("module_loaded") is True, "stub 模块载入（import 无副作用面）")
    ssegs = srec["segments"]
    check("FAM-OTHERVER", len(ssegs) == 12 and all(
        s["claims"] == [{"line": 1, "rule": "OTHER", "token": "ledger audit: orphan claim"}]
        and s["verdict"] in ("TP", "FP") for s in ssegs),
        f"stub×全段：OTHER 捕获且 got=True（FN/TN 出现=got 非 claims 真源=A8 复发）")

    # —— FAM-EQUIV 声明计数 vs 旧口径等价钉 ——
    for gname, gpath in rmod.DEFAULT_GATES:
        for sid, _, body_lines, _ in segs_list:
            stdout = rmod.run_gate_raw(gpath, body_lines)
            n_claims = len(rmod.parse_claims(stdout, None, body_lines))
            check("FAM-EQUIV", n_claims == stdout.count("count claim"),
                  f"{gname}/{sid} claims==count('count claim')")

    # —— FAM-REALDOCS 真载体 e2e 面（私有语料缺席时可见跳过，不崩不静默）——
    gpaths = {g[0]: g[1] for g in rmod.DEFAULT_GATES}
    ledger_rel = "seeds/truth-doc-A.md"
    for rel, pins in REALDOCS.items():
        if not (SB / rel).exists():
            print(f"SKIP [FAM-REALDOCS] {rel.split('/')[-1]} 不在场（私有语料未随包），"
                  f"跳过 {len(pins)}+{'182面' if rel == ledger_rel else ''}检查——"
                  "复现者请在含真载体的战役档案环境跑满")
            continue
        for gv, want in pins.items():
            n = gate_check(gpaths[gv], SB / rel).count("count claim")
            check("FAM-REALDOCS", n == want, f"{gv}×{rel.split('/')[-1]}={want}（实得 {n}）")
        if rel == ledger_rel and (SB / ledger_rel).exists():
            for gv, want in L182.items():
                if want is None:
                    continue
                hit = ":182: count claim" in gate_check(gpaths[gv], SB / ledger_rel)
                check("FAM-REALDOCS", hit == want, f"{gv} 台账:182（`ls`真载体）claim={want}")

    # —— 持久化面 ——
    check_persist("FAM-PERSIST", RESULTS)
    check_persist("FAM-PERSIST-GT", GT_RESULTS)
    gt_doc = json.loads((GT_RESULTS / SCORES).read_text(encoding="utf-8"))
    check("FAM-PERSIST-GT", gt_doc["corpus"]["sha256"] == sha(SB / "seeds" / "d1-corpus-full.md"),
          "正副本记录之语料 hash 与盘上一致")

    # —— FAM-WARN 闸转义门 ——
    rc = subprocess.run([sys.executable, "-W", "error::SyntaxWarning", "-m", "py_compile", str(GATE_WARN)],
                        capture_output=True, text=True)
    check("FAM-WARN", rc.returncode == 0, f"{GATE_WARN.name} -W error 编译净")
    for gv in ("claim-gate-v2.7.1.py", "claim-gate-v2.7.2.py"):  # R3 内环守恒补口：现行血统编译钉全覆盖
        r2 = subprocess.run([sys.executable, "-W", "error::SyntaxWarning", "-m", "py_compile",
                             str(SB / "gates" / gv)], capture_output=True, text=True)
        check("FAM-WARN", r2.returncode == 0, f"{gv} -W error 编译净（永久钉，非一次性冒烟）")

    return finish()


def finish():
    print(f"\nPROBES: checks={N_CHECK} fails={len(FAILS)}")
    for fam, msg in FAILS:
        print(f"  RED [{fam}] {msg}")
    print(f"test-surface={N_CHECK}")
    return 1 if FAILS else 0


if __name__ == "__main__":
    sys.exit(main())
