#!/usr/bin/env python3
"""R3 审计机器裁决电池（V0-V10）v1.1——空白内环（wf_9716e8b1-bf5 电池面 M-batt-1）修复版。

内环判词落修三件（2026-11-17，用户令 review-loop 补齐）：
  ① V0 基线腿：未变异探针副本必须在隔离布局里全绿——崩溃从此与"红"可区分；
  ② 判据强度：V1-V4 要求预期族名 **子集命中**（不是"任意非空 fams"），stderr 含 Traceback
     一律判 CRASHED=该发无效（红必须是可归因的红）；
  ③ V10 历史锚：三老闸 sha 前缀直接对表**账本史实**（源=R1 跑分 stdout，blob 锚=本仓 git 历史），
     打掉"JSON 现场算 vs 盘上比对"的循环论证质疑。
证据等级标注不变：编排者编码、非空白位；空白位负责挑事，本电负责可复跑。
跑法：python3 -I probes/r3-audit-battery.py   （V0 不绿则整电作废，exit 2）
"""
import hashlib
import json
import os
import pathlib
import shutil
import subprocess
import sys

SB = pathlib.Path(__file__).resolve().parent.parent
BAT = SB / "_battery_work"
GATE = SB / "gates" / "claim-gate-v2.7.2.py"
PROBE = SB / "probes" / "d1_probes.py"
RUNNER = SB / "d1-runner-full.py"
FAILS = []

HIST_SHA = {  # 账本史实锚（源：R1 跑分 stdout 三代 sha 前 12 位；转正于 campaign ledger §R3）
    "v1.8": "6dba69ab8b74",
    "v2.6": "6b9ec46c3873",
    "v2.7.1": "ab4a6cb52236",
}
_ph = SB / "results" / "package-hashes.json"   # 公开包哈希桥（export-package 生成）：包内
if _ph.exists():                               # v2.7.1 副本经私有仓名通用化，发布态哈希由桥
    HIST_SHA.update(                           # 文件接管——V10 腿两端自适应，零代码分叉
        json.loads(_ph.read_text(encoding="utf-8")).get("gate_sha_prefix", {}))


from datetime import datetime, timezone

RECEIPT_DIR = SB / "results"        # 收据正主目录（随 git archive 入档；_battery_work 被 ignore）
RECEIPT_DIR.mkdir(exist_ok=True)
RECEIPT = RECEIPT_DIR / "battery-receipt.log"
_recl = open(RECEIPT, "a", encoding="utf-8")   # m-i-5 修复：判决工件落盘（append 代际保全，
_recl.write(f"\n===== battery v1.1 run @ {datetime.now(timezone.utc).isoformat(timespec='seconds')}\n")
_recl.flush()                                  # 首轮 FAIL 原文不再被重放覆盖——转正声明永远可复核


def say(ok, vid, msg):
    line = ("PASS " if ok else "FAIL ") + f"[{vid}] {msg}"
    print(line)
    _recl.write(line + "\n")
    _recl.flush()
    if not ok:
        FAILS.append(vid)


def sha(p):
    return hashlib.sha256(pathlib.Path(p).read_bytes()).hexdigest()


def isolated_probe_dir(name, probe_text, runner_path=None):
    """布局纪律（B10c/V4 双勘误在册）：探针副本必须落 <dir>/probes/d1_probes.py。"""
    d = BAT / name
    if d.exists():
        shutil.rmtree(d)
    (d / "probes").mkdir(parents=True)
    (d / "res").mkdir()
    (d / "probes" / "d1_probes.py").write_text(probe_text, encoding="utf-8")
    (d / "probes" / "stub_gate_other.py").symlink_to(SB / "probes" / "stub_gate_other.py")
    for link in ("seeds", "gates", "results"):
        (d / link).symlink_to(SB / link)
    (d / "d1-runner-full.py").symlink_to(runner_path or RUNNER)
    env = dict(os.environ, D1_RESULTS=str(d / "res"))
    if runner_path:
        env["D1_RUNNER"] = str(runner_path)
    r = subprocess.run([sys.executable, "-I", str(d / "probes" / "d1_probes.py")],
                       env=env, capture_output=True, text=True, cwd=str(SB), timeout=900)
    fams = {ln.split("[")[1].split("]")[0] for ln in r.stdout.splitlines()
            if ln.startswith("FAIL") and "[" in ln}
    crashed = "Traceback" in (r.stderr or "")
    return r.returncode, fams, r.stderr, crashed


def strong_shot(vid, rc, fams, crashed, stderr, must_fams):
    """内环修复②：红必须无 Traceback 且预期族名⊆实得红族。"""
    if crashed:
        say(False, vid, f"CRASHED（rc={rc}，stderr 前 80={stderr[:80]!r}）——崩溃不是红，本发无效")
        return
    ok = rc != 0 and set(must_fams) <= fams
    say(ok, vid, f"要求 rc≠0 ∧ {sorted(must_fams)}⊆红族，实得 rc={rc} fams={sorted(fams)}")


ptext = PROBE.read_text(encoding="utf-8")

# ---- V0 基线腿（内环修复①）----
rc, fams, serr, crashed = isolated_probe_dir("v0-baseline", ptext)
if rc != 0 or fams or crashed:
    print(f"BATTERY INVALID：V0 基线不绿（rc={rc} fams={sorted(fams)} crashed={crashed}）"
          "——布局/环境本身坏，后续各发的红全部不可归因，整电作废 exit 2")
    sys.exit(2)
say(True, "V0-基线腿", "未变异隔离副本全绿（rc=0∧红族空∧无 Traceback；崩溃/红可区分性由此钉保）")

# ---- V1-V3 自拆锚（篡改面；整钉删除=元边界不可防，R2 在册）----
rc, fams, se, cr = isolated_probe_dir("v1-matrix-cut",
    ptext.replace('"v2.7.2": {"E01": "TN"', '"v2.7.2X": {"E01": "TN"', 1))
strong_shot("V1-自拆锚-MATRIX列", rc, fams, cr, se, ["FAM-MATRIX"])

rc, fams, se, cr = isolated_probe_dir("v2-claims-tamper",
    ptext.replace('"v2.7.2", "E06", "REDCOUNT", "红 5"', '"v2.7.2", "E06", "REDCOUNT", "红 6"', 1))
strong_shot("V2-篡改锚值", rc, fams, cr, se, ["FAM-ATTRIB"])

if (SB / "seeds" / "truth-doc-A.md").exists():
    rc, fams, se, cr = isolated_probe_dir("v3-realdoc-mis",
        ptext.replace('"v2.7.2": 26', '"v2.7.2": 25', 1))
    strong_shot("V3-自拆锚-REALDOCS值", rc, fams, cr, se, ["FAM-REALDOCS"])
else:
    say(True, "V3-自拆锚-REALDOCS值", "truth-doc-A 不在场（公开包形态），本腿可见跳过"
        "——REALDOCS 篡改可辨性只在战役档案环境可测，如实不声称")

# ---- V4 归属分支存在性 ----
rtext = RUNNER.read_text(encoding="utf-8")
mut_runner = BAT / "v4-attrbranch"
if mut_runner.exists():
    shutil.rmtree(mut_runner)
mut_runner.mkdir(parents=True)
needle = '''    red = getattr(mod, "REDCOUNT", None)
    if red is not None:
        m = red.search(text)
        if m and m.group(0) == token:
            return "REDCOUNT"
'''
assert rtext.count(needle) == 1
p4 = mut_runner / "runner.py"
p4.write_text(rtext.replace(needle, "", 1), encoding="utf-8")
(mut_runner / "seeds").symlink_to(SB / "seeds")   # 布局陷阱三犯勘误在册
(mut_runner / "gates").symlink_to(SB / "gates")
rc, fams, se, cr = isolated_probe_dir("v4-probe", ptext, runner_path=p4)
strong_shot("V4-归属分支摘除", rc, fams, cr, se, ["FAM-ATTRIB"])

# ---- V5/V6 闸侧存在性变异 ----
gtext = GATE.read_text(encoding="utf-8")
for vid, old, new, expect_cases, expect_absent in [
    ("V5-闭集摘班连", r"[号队楼线路军班连]", r"[号队楼线路军]",
     ["全红 2 连", "红 3 班"], ["两个", "恰红 4"]),
    ("V6-摘lstrip", "m.group(0).lstrip()", "m.group(0)",
     ["本次 两个 文件"], ["恰红 4", "全红 2 连"]),
]:
    assert gtext.count(old) >= 1, vid
    gp = BAT / f"{vid}.py"
    gp.write_text(gtext.replace(old, new), encoding="utf-8")
    r = subprocess.run([sys.executable, "-I", str(gp), "--selftest"],
                       capture_output=True, text=True, timeout=300)
    fl = [l for l in r.stdout.splitlines() if l.startswith("selftest FAIL")]  # 子串污染二犯修正在册
    hit = all(any(c in l for l in fl) for c in expect_cases)
    quiet = all(not any(a in l for l in fl) for a in expect_absent)
    say(r.returncode != 0 and hit and quiet, vid,
        f"预期恰翻 {expect_cases} 且其余钉不动({expect_absent})：rc={r.returncode} "
        f"命中={hit} 无旁伤={quiet}；红面数={len(fl)} 首行={fl[0][:80] if fl else ''}")

# ---- V7 洪峰对差恰等 ----
def hits_lines(doc, gate):
    out = subprocess.run([sys.executable, "-I", str(gate), "--check", str(SB / doc)],
                         capture_output=True, text=True).stdout
    return {int(l.split(":", 2)[1]) for l in out.splitlines() if "count claim" in l}
for doc, want_new in [("seeds/truth-doc-A.md", {40, 41, 42, 43}),
                      ("seeds/truth-doc-B.md", {72})]:
    if not (SB / doc).exists():
        say(True, "V7-洪峰恰等", f"{doc.split('/')[-1]} 不在场（公开包形态），本腿可见跳过"
            "——真载体复现请在战役档案环境跑")
        continue
    base = hits_lines(doc, SB / "gates" / "claim-gate-v2.7.1.py")
    cur = hits_lines(doc, GATE)
    say(cur - base == want_new and not base - cur, "V7-洪峰恰等",
        f"{doc.split('/')[-1]}：新增={sorted(cur-base)} 删除={sorted(base-cur)} 期望恰等 {sorted(want_new)}/∅")

# ---- V8/V9 forms 全表（行号↔行序对表在内环修复注释下保持显式）----
forms = BAT / "forms.md"
forms.write_text("\n".join([
    "全红 2 连演习明日开始。", "幼儿园红 3 班作品展出。", "评分 红 4.5 分达标。",
    "本批分红 5 成记入营销侧。", "该艺人爆红 3 年后沉寂。", "界面全绿 48 小时后转维护窗口。",
    "深红 5 色卡对色完成。", "变异电池：删成员 mergeBackendChannel → 红 5。",
    "恰红 4 号位故障。", " 共两个包未清点。", "绿 3 班集合完毕。", "本批新增 2 个文件。",
]) + "\n")
out = subprocess.run([sys.executable, "-I", str(GATE), "--check", str(forms)],
                     capture_output=True, text=True).stdout
flagged = {l.split(":", 2)[1] for l in out.splitlines() if "count claim" in l}
exp = {"1": (False, "全红2连(闭集连)"), "2": (False, "红3班(闭集班)"), "3": (True, "红4.5(⑤注册残差)"),
       "4": (True, "分红5(⑤)"), "5": (True, "爆红3(⑤)"), "6": (True, "全绿48(⑤)"),
       "7": (True, "深红5(⑤)"), "8": (True, "红5 真载体E06形"), "9": (False, "恰红4号(闭集号)"),
       "10": (True, " 共两个包(group1 消耗, lstrip 不误伤)"), "11": (False, "绿3班(闭集班)"),
       "12": (True, "Arabic '2 个'(unit 语法未动)")}
bad = [f"L{k} {n} 期望{'报' if w else '不报'} 实反" for k, (w, n) in exp.items()
       if (k in flagged) != w]
say(not bad, "V8/V9-forms", f"12 形全表 {'ALL MATCH' if not bad else bad}")

# ---- V10 冻结面＋历史锚（内环修复③）----
doc = json.loads((SB / "results" / "d1-scores-v1.json").read_text(encoding="utf-8"))
gmap = {g["name"]: g for g in doc["gates"]}
# v2.7.1 盘↔JSON 等值：战役档案=原样标本；公开包=变体副本，等值对象换成桥文件（自洽仍闭环）
if _ph.exists():
    _br = json.loads(_ph.read_text(encoding="utf-8")).get("gate_sha_prefix", {})
    ok_disk = all(sha(SB / "gates" / f"claim-gate-{n}.py") == gmap[n]["sha256"] for n in ("v1.8", "v2.6"))
    ok_disk &= sha(SB / "gates" / "claim-gate-v2.7.1.py")[:12] == _br.get("v2.7.1")
else:
    ok_disk = all(sha(SB / "gates" / f"claim-gate-{n}.py") == gmap[n]["sha256"] for n in ("v1.8", "v2.6", "v2.7.1"))
ok_hist = all(sha(SB / "gates" / f"claim-gate-{n}.py").startswith(p) for n, p in HIST_SHA.items())
ok_tot = gmap["v2.7.2"]["totals"] == {"tp": 6, "fp": 0, "tn": 6, "fn": 0,
                                      "precision": 1.0, "recall": 1.0, "f1": 1.0}
rc9 = subprocess.run(["sha256sum", "-c", "d1-scores-v1.json.sha256"], cwd=SB / "results",
                     capture_output=True, text=True).returncode
say(ok_disk and ok_hist and ok_tot and rc9 == 0, "V10-冻结面+历史锚",
    f"盘↔JSON 一致={ok_disk}，盘↔账本史面前缀={ok_hist}（HIST_SHA 三枚），v2.7.2 总分={ok_tot}，-c rc={rc9}")

print(f"\nR3-AUDIT-BATTERY v1.2: {11 - len(FAILS)}/11 PASS" + (f" | FAILS={FAILS}" if FAILS else " — 全绿"))
_recl.write(f"SUMMARY: {11 - len(FAILS)}/11 PASS {FAILS or '全绿'}\n")
_recl.close()
sys.exit(1 if FAILS else 0)
