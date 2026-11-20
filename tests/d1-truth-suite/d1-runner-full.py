#!/usr/bin/env python3
"""D1 十事件真值套件 · v1.2 完整版（R3 收编后）
（声明级粒度 · 注记剥离 · 白名单分裂探针 · 真值持久化）

用法：python3 -I d1-runner-full.py [gate_path ...]
缺省测四代（v1.8 / v2.6 / v2.7.1 / v2.7.2）；传 gate 路径可测任意版本。

v1.1 交付五件（对应 2026-11-15 指令）：
  1. 语料段D2：带 `ls`/`find` 反引号证据命令的合法段——簇① CMD 白名单分裂探针。
     源码实证：v1.8 / v2.6 反引号分支缺 ls|find（claim-gate-v1.8.py:153-156、
     claim-gate-v2.6.py:739-742），该段应命中（=FP，病活着）；
     v2.7.1 已 _CMD_TOOLS 单源修复（claim-gate-v2.7.1.py:859-862），应放行。
  2. 注记剥离：以"（事实："起的行喂闸前滤除。前版实测注记渗透进判定
     （v1.8 的 E08 命中源自注记"2 处"而非正文"3/5 已改"——v1.8 无 PROGRESS
     档；剥离后该段预期翻 FN，属诚实修正非退化）。
  3. 声明级粒度：解析每条 `path:行号: count claim 'token' without ...` finding，
     规则归属直接调该闸模块自己的 find_claim/TRIPLE/PROGRESS/UNITS
     （importlib 载入，零镜像漂移——闸自己的 P-Ⅰ 单源律）。
     规则取值：COUNTING/<单位词> | TRIPLE | PROGRESS | REDCOUNT | OTHER | COUNTING/?（兜底位，哨兵钉死）。
  4. 持久化：results/d1-scores-v1.json（版本、时间戳、闸与语料 sha256、
     逐声明明细、逐段判定、混淆矩阵+精确率/召回率/F1）
     + results/d1-scores-v1.json.sha256（sha256sum -c 兼容格式）。
  5. （闸侧随交付）v2.7.1 docstring :557 `\\s` 转义修复，消 SyntaxWarning，
     文档渲染语义不变。
"""
import hashlib
import importlib.util
import json
import os
import pathlib
import re
import subprocess
import sys
import tempfile
from datetime import datetime, timezone

SB = pathlib.Path(__file__).parent
CORPUS = SB / "seeds" / "d1-corpus-full.md"
TMP = pathlib.Path(tempfile.mkdtemp(prefix="d1_seg_")) / "_seg_tmp.md"
RESULTS_DIR = pathlib.Path(os.environ.get("D1_RESULTS", SB / "results"))
SCORES_NAME = "d1-scores-v1.json"
RUNNER_VERSION = "d1-runner-full.py v1.2（R3 收编：四代×48 格＋REDCOUNT 归属档）"

# 期望判定表：段 ID → (应命中?, 说明)
# True = 该段含病文本，闸应命中；False = 合法或域外，闸应放行
EXPECT = {
    "E01": (False, "错误码抄串——坐标/码失实，claim_gate 域外（按设计）"),
    "E02": (False, "抛点机制写错——语义层，claim_gate 域外"),
    "E03": (False, "行号硬错——坐标失实，claim_gate 域外"),
    "E04": (True,  "旗舰形——声明-实态差（计数声明缺证据），应命中"),
    "E05": (False, "脚本静默空转——过程性失实，非文本计数声明"),
    "E06": (True,  "变异红集虚高——计数声明'红 5'，应命中"),
    "E07": (True,  "闸回归位假闭合——真载体'12 攻击样本'为隔词闭合形不在现文法（E07 裁决票 2026-11-16："
                   "现 TP 由叙事'三类'搭便车，CN 泛化收紧落地即改判 FN；闭合形票 v2.7.3 以真 token 反制）"),
    "E08": (True,  "虚报进度'3/5 已改'——PROGRESS 完成态形，应命中"),
    "E09": (True,  "陈旧计数'9 个回归样本'——应命中"),
    "E10": (True,  "已 POST 计数无证据——'1826 用例'等，应命中"),
    "段C": (False, "合法段（grep 反引号复算命令），应放行"),
    "段D2": (False, "合法段（ls/find 反引号证据）——簇①探针：命中＝该版白名单分裂病活着"),
}

DEFAULT_GATES = [
    ("v1.8",   SB / "gates" / "claim-gate-v1.8.py"),
    ("v2.6",   SB / "gates" / "claim-gate-v2.6.py"),
    ("v2.7.1", SB / "gates" / "claim-gate-v2.7.1.py"),
    ("v2.7.2", SB / "gates" / "claim-gate-v2.7.2.py"),
]

ANNOT_RE = re.compile(r"^\s*（事实[:：]")
FINDING_RE = re.compile(r"^(?P<path>.+?):(?P<no>\d+): (?P<rest>.+)$")
CLAIM_RE = re.compile(
    r"^count claim (?P<tok>.+?) without a command in the ±(?P<win>\d+)-line window: (?P<ctx>.*)$"
)


def _rel(p):
    """出站卫生（m1 扩展，148号复审 2026-11-18）：包内路径一律记相对形态，
    防沙箱绝对路径泄漏进 tracked GT。包外路径降为首文件名。"""
    try:
        return str(pathlib.Path(p).resolve().relative_to(SB.resolve()))
    except ValueError:
        return pathlib.Path(p).name


def sha256(path: pathlib.Path) -> str:
    return hashlib.sha256(path.read_bytes()).hexdigest()


def segments():
    """按 ## 切段；剥离（事实：注记行。yield (seg_id, title, body_lines, n_dropped)。"""
    src = CORPUS.read_text(encoding="utf-8")
    parts = re.split(r"^## ", src, flags=re.M)
    for p in parts[1:]:
        lines = p.splitlines()
        title = lines[0].strip()
        kept, dropped = [], 0
        for ln in lines[1:]:
            if ANNOT_RE.match(ln):
                dropped += 1
            elif ln.strip() or kept:          # 保留正文与空行占位（行号对齐闸输出）
                kept.append(ln)
        seg_id = title.split("·")[0].strip().split()[0]
        yield seg_id, title, kept, dropped


def load_gate_module(gate_path: pathlib.Path):
    """importlib 载入闸模块，取它自己的 find_claim/TRIPLE/PROGRESS/UNITS 做归属。"""
    spec = importlib.util.spec_from_file_location(f"d1gate_{gate_path.stem}", gate_path)
    mod = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(mod)
    return mod


def strip_repr(tok: str) -> str:
    if len(tok) >= 2 and tok[0] in "'\"" and tok[-1] == tok[0]:
        return tok[1:-1]
    return tok


def attribute_rule(mod, body_lines, lineno, token):
    """把 (行号, token) 归到闸的规则档：TRIPLE / PROGRESS / REDCOUNT / COUNTING/<单位词>（兜底 /?）。"""
    idx = lineno - 1
    text = body_lines[idx] if 0 <= idx < len(body_lines) else ""
    triple = getattr(mod, "TRIPLE", None)
    if triple is not None:
        m = triple.search(text)
        if m and m.group(0) == token:
            return "TRIPLE"
    prog = getattr(mod, "PROGRESS", None)
    if prog is not None:
        m = prog.search(text)
        if m and m.group(0) == token:
            return "PROGRESS"
    red = getattr(mod, "REDCOUNT", None)
    if red is not None:
        m = red.search(text)
        if m and m.group(0) == token:
            return "REDCOUNT"
    for u in sorted(getattr(mod, "UNITS", []), key=len, reverse=True):
        if token.endswith(u):
            return f"COUNTING/{u}"
    return "COUNTING/?"


def run_gate_raw(gate_path: pathlib.Path, body_lines):
    TMP.write_text("\n".join(body_lines) + "\n", encoding="utf-8")
    r = subprocess.run(
        [sys.executable, "-I", str(gate_path), "--check", str(TMP)],
        capture_output=True, text=True, timeout=20,
    )
    return r.stdout


def parse_claims(stdout, mod, body_lines):
    """stdout → 声明级列表。非 count-claim 形态的 path:NN: finding 记 OTHER，不静默吞。"""
    claims = []
    for ln in stdout.splitlines():
        m = FINDING_RE.match(ln)
        if not m or pathlib.Path(m.group("path")).name != TMP.name:
            continue
        lineno = int(m.group("no"))
        cc = CLAIM_RE.match(m.group("rest"))
        if cc:
            token = strip_repr(cc.group("tok"))
            rule = attribute_rule(mod, body_lines, lineno, token) if mod else "UNATTRIBUTED"
        else:
            token = m.group("rest")[:80]
            rule = "OTHER"
        claims.append({"line": lineno, "rule": rule, "token": token})
    return claims


def score_gate(gname, gpath):
    rec = {"name": gname, "path": _rel(gpath), "exists": gpath.exists(), "segments": []}
    if not rec["exists"]:
        return rec
    rec["sha256"] = sha256(gpath)
    try:
        mod = load_gate_module(gpath)
        rec["module_loaded"] = True
        rec["window"] = getattr(mod, "WINDOW", None)
        rec["has_progress"] = hasattr(mod, "PROGRESS")
    except Exception as e:  # 归属退化为 UNATTRIBUTED，判定主链不受影响
        mod, rec["module_loaded"] = None, False
        rec["import_error"] = f"{type(e).__name__}: {e}"
    for seg_id, title, body_lines, dropped in segments():
        expected, note = EXPECT.get(seg_id, (None, "未登记"))
        stdout = run_gate_raw(gpath, body_lines)
        claims = parse_claims(stdout, mod, body_lines)
        got = len(claims) > 0
        if expected is None:
            verdict = "UNREGISTERED"
        elif expected and got:
            verdict = "TP"
        elif expected and not got:
            verdict = "FN"
        elif not expected and got:
            verdict = "FP"
        else:
            verdict = "TN"
        rec["segments"].append({
            "id": seg_id, "title": title, "expected_hit": expected, "note": note,
            "annotations_dropped": dropped, "claims": claims,
            "hits": len(claims), "verdict": verdict,
        })
    c = {k: sum(1 for s in rec["segments"] if s["verdict"] == k) for k in ("TP", "FP", "TN", "FN")}
    tp, fp, fn = c["TP"], c["FP"], c["FN"]
    prec = tp / (tp + fp) if tp + fp else 0.0
    rec_ = tp / (tp + fn) if tp + fn else 0.0
    rec["totals"] = {
        "tp": tp, "fp": fp, "tn": c["TN"], "fn": fn,
        "precision": round(prec, 4), "recall": round(rec_, 4),
        "f1": round(2 * prec * rec_ / (prec + rec_), 4) if prec + rec_ else 0.0,
    }
    return rec


def main():
    args = [a for a in sys.argv[1:] if not a.startswith("--")]
    write_results = "--write-results" in sys.argv[1:]
    gates = DEFAULT_GATES
    if args:
        gates = [(pathlib.Path(a).stem, pathlib.Path(a)) for a in args]

    print("=" * 78)
    print(f"D1 十事件真值套件 v1.2 · 声明级跑分（四代×48 格；注记剥离＋段D2 白名单探针）")
    print("=" * 78)

    gate_recs = []
    for gname, gpath in gates:
        if not gpath.exists():
            print(f"\n### {gname}: 文件不存在，跳过")
            continue
        r = score_gate(gname, gpath)
        gate_recs.append(r)
        t = r["totals"]
        print(f"\n### {gname} ({gpath.name})  window=±{r.get('window')} "
              f"PROGRESS={'有' if r.get('has_progress') else '无'}  sha256={r['sha256'][:12]}…")
        for s in r["segments"]:
            exp = "命中" if s["expected_hit"] else ("?" if s["expected_hit"] is None else "放行")
            mark = "✓" if s["verdict"] in ("TP", "TN") else ("✗" if s["verdict"] in ("FP", "FN") else " ? ")
            print(f"{s['id']:<6} 期望{exp:<3} 实得{s['hits']:<3} {mark} {s['verdict']:<3} "
                  f"{s['note'][:36]}{'｜注记-{}'.format(s['annotations_dropped']) if s['annotations_dropped'] else ''}")
            for cl in s["claims"]:
                print(f"        L{cl['line']:<4} [{cl['rule']}] {cl['token']!r}")
        print("-" * 78)
        print(f"TP={t['tp']} FP={t['fp']} TN={t['tn']} FN={t['fn']}  |  "
              f"精确率={t['precision']:.2f} 召回率={t['recall']:.2f} F1={t['f1']:.2f}")

    doc = {
        "suite": "D1-ten-event-v1",
        "runner": RUNNER_VERSION,
        "generated_utc": datetime.now(timezone.utc).isoformat(timespec="seconds"),
        "python": sys.version,
        "corpus": {"path": _rel(CORPUS), "sha256": sha256(CORPUS), "n_segments": len(list(segments()))},
        "expect": {k: {"should_hit": v[0], "note": v[1]} for k, v in EXPECT.items()},
        "semantics": "段级二值判定（hits>0），声明级明细见 segments[].claims；"
                     "E01-E03/E05 域外事件期望放行非漏检，归坐标闸管辖；"
                     "其中 E01 在 v2.6/v2.7.1 各产 1 FP（CN 叙事数词『两个』，簇⑤形），已计入 precision。",
        "gates": gate_recs,
    }
    # 持久化目标（m1 修复，148号复审 2026-11-18）：默认写 scratch，防克隆形态
    # 一跑就把 tracked 的 results/*.json 弄脏（主线 worktree 起点即脏即此故）。
    # 显式更新正副本：--write-results；调用方也可用 D1_RESULTS 指定目录
    # （d1_probes / run-mutants / final-review-battery 均如此，互不影响）。
    if "D1_RESULTS" in os.environ:
        res_dir = pathlib.Path(os.environ["D1_RESULTS"])
    elif write_results:
        res_dir = RESULTS_DIR
    else:
        res_dir = SB / "_probe_results"
        print("（m1：结果写 scratch _probe_results/；更新 tracked 正副本请加 --write-results）")
    res_dir.mkdir(exist_ok=True, parents=True)
    spath = res_dir / SCORES_NAME
    spath.write_text(json.dumps(doc, ensure_ascii=False, indent=1) + "\n", encoding="utf-8")
    digest = sha256(spath)
    cpath = res_dir / (SCORES_NAME + ".sha256")
    cpath.write_text(f"{digest}  {SCORES_NAME}\n", encoding="utf-8")

    print("\n" + "=" * 78)
    print(f"真值已持久化：{spath}")
    print(f"校验文件　　：{cpath}")
    print(f"验证：cd {res_dir} && sha256sum -c {SCORES_NAME}.sha256")
    return 0


if __name__ == "__main__":
    sys.exit(main())
