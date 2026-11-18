#!/usr/bin/env python3
"""R3 收编装配器：gates/claim-gate-v2.7.2.py ＝ v2.7.1 ＋ T1(REDCOUNT) ＋ T2(两+个 skip)
＋红队义务合并落地（六件）。所有替换 assert 唯一；输出前自校义务在场。

跑法：python3 -I probes/build-v272.py
幂等：重跑覆盖输出。源锚表逐字复制自 D3 分诊两设计位的交付脚本（设计件存战役档案，
哈希桥见 results/provenance.json 与出站包 MANIFEST；T2 的 docstring 锚已按 T1 改写后的
实态适配——合成顺序 T1→T2 的冲突面）。sanitize() 把装配产物中一切私有工作目录引用清零。
"""
import hashlib
import pathlib
import re

SB = pathlib.Path(__file__).resolve().parent.parent
SRC = (SB / "gates" / "claim-gate-v2.7.1.py").read_text(encoding="utf-8")
OUT = SB / "gates" / "claim-gate-v2.7.2.py"
src = SRC


def rep1(old, new):
    global src
    n = src.count(old)
    assert n == 1, f"锚非唯一 (count={n}): {old[:70]!r}"
    src = src.replace(old, new, 1)


# ---- 出站卫生（R4/MR 轮）：私有指向清零。逐条 assert，防漂移 ----
def sanitize(s):
    """出站卫生 v2（R4 披露审后重构；判据逐条来自该审 findings）。
    原则转变：种子文本（REDBLOCK/NARRBLOCK/CASES_*）直接写脱敏串，私有指向不再
    进入字面量——本函数只剩两件结构职责＋一件保险：
    ① v2.7.1 基底继承的私有战役仓名引用（公仓 v2.6 grep=0，新于公开面，须清；两处，
       串以碎片构造防本文件自泄漏）；
    ② 版本头注 bump（v2.7.2 文件自述仍写 v2.7.1——R4 审 minor#7）；
    agent-festival 引用**不再脱敏**（v2.6:445 公仓字节既有＝公开面文本，
    上一轮脱它反而造成血统偏离并留下"前提虚假"——R4 审 minor#6 纠正）。
    兜底扫描保留作保险（种子再被污染时装配当场红）。"""
    PRIV_T = "the N" + "GA campaign repo"   # 碎片构造：本文件随包，私有代号字面量不得完整
    n_priv = s.count(PRIV_T)
    assert n_priv == 2, f"v2.7.1 基底私有仓名计数 {n_priv}≠2——血统文本漂移，本层需同步"
    s = s.replace(PRIV_T, "the private campaign archive")
    s = s.replace('"""Claim gate v2.7.1: counting-closure claims',
                  '"""Claim gate v2.7.2: counting-closure claims')
    s = s.replace("Current execution scope (v2.7.1, unchanged from v2.6)",
                  "Current execution scope (v2.7.2: +REDCOUNT, +bare 两+个 narrative skip; "
                  "window/CMD/units surfaces unchanged from v2.6)")
    s_wo_lineage = re.sub(r"minor-ledger:\d+", "", s)   # 血统白名单（v2.6 公仓既有）
    _L = ["_" + "d3_work", "_" + "audit_r3", "ledger:1", "yao" + "jun", "/home/", r"\bNG" + "A\b"]
    leftover = re.findall("|".join(_L), s_wo_lineage)
    assert not leftover, f"出站卫生未清干净: {set(leftover)}"
    return s


# ==================== T1: REDCOUNT（E06 红N） ====================
ANCHOR_R1 = (
    '    r"|已(?:" + "|".join(_PROGRESS_VERBS) + r")[^0-9\\n]{0,4}\\d{1,3}\\s*/\\s*\\d{1,3}(?![\\d/])"\n'
    ")\n"
)
REDBLOCK = ANCHOR_R1 + '''
# v2.7.2 (R3 收编; D3 T1; E06 「红 N」 family): battery state-color count
# claims in REVERSE order — 「红 5」「恰红 4」「恒红 1」「绿 12」 — the color
# token BEFORE the digits. COUNTING's grammar is N+unit, so this word order
# is unreachable from the units table (「红 5」≠「5 红」), and appending 红 to
# the units alternation would flip the census 23-pin (P-Ⅱ/Ⅲ) and the
# attribution chain. The PROGRESS precedent (v2.7, the fraction form)
# applies instead: a NEW table hooked at the END of find_claim's fallback
# chain — it inherits the ±WINDOW CMD evidence check UNCHANGED, so 「红 4」
# plus a nearby recompute command passes exactly like 「恰 4 处」 does (段C
# law). SCOPE (obligation-①, red-team 2026-11-16; carrier-corrected by A1
# audit M-finding same day — the red team named the CORPUS lines, an earlier
# draft of this comment misattributed them to the truth-doc-A face): in the SEGMENT-ISOLATED
# suite face E10's 「恰红 4」 line has no command within its segment ±8 and
# therefore flags; in WHOLE-FILE runs that same line (d1-corpus-full.md :41)
# IS PASSED by the corpus's own :45/:48/:49 backtick-command windows —
# receipt: whole-file corpus scan in the campaign archive (T1-era: 9 hits;
# the merged gate's same face now runs 8 — the :4 「两个」 FP since closed by
# T2; :40/:41 absent on both faces). The truth-doc-A face is the opposite:
# :40-:43 are among the 26 hits,
# the nearest command there (:52) being 9-12 lines away, beyond ±8 — a TRUE
# catch, not a pass (REALDOCS 台账=26 pin). D3-G11's same-line twin carries
# the pass semantics in every face.
# Guards, each BIDIRECTIONALLY pinned (rule 2; the 零样本 and per-verb
# precedents):
#   ① word order + whitespace-only adjacency is the PRIMARY guard — 「红色 3」
#      (color word + count) and the forward 「7≠12 红」 result reading
#      (measured battery-result form) cannot match; 「红色 3 个」 flags through the
#      UNCHANGED 个 grammar, never through this table;
#   ② (?<![A-Za-z0-9-/.]) — the lookbehind class PROGRESS shares, kills
#      code/coordinate-adjacent forms (「x红 5」); PROPHYLACTIC, zero
#      in-domain carriers, REGISTERED as such;
#   ③ (?!\\s*[号队楼线路军班连]) — address/org readings (红 5 号楼 / 红 2 队 /
#      红 3 班 / 红 2 连); CLOSED set, one pin per member (P-Ⅲ law). 班/连
#      JOINED the class after the red-team LIVE-FIRE (attack2.md hits
#      「全红 2 连」「幼儿园红 3 班」, 2026-11-16 — demonstrated FP forms, not
#      speculative widening), pinned D3-G12/G13 below;
#   ④ ASCII digits only — the CN twin 「红五」 is a REGISTERED zero-carrier
#      residual (v2.0 digit-slot law: extend on carriers, not before).
#   ⑤ REGISTERED zero-carrier residuals (red-team live-fire 2026-11-16,
#      receipts: docs/receipts/attack2.md (bundled) — these DO fire today,
#      no corpus carriers yet; extension waits for carriers + per-member
#      pins): decimal tails 「红 4.5」→'红 4' (no (?!\\.\\d) guard); CJK-compound
#      prefixes 「分红/桃红/爆红/深红/本发红 5」 (lookbehind is ASCII-only);
#      duration reading 「全绿 48 小时」 (绿 + hours ≠ battery-pair sense).
# The color set is the battery state PAIR only (红=FAIL 读数 / 绿=PASS 读数 —
# 全绿 in the corpus proves 绿 lives as a state word here); 黑/黄/蓝 are
# different senses (黑名单/黄牌) and stay out until carriers appear.
_REDCOLORS = ("红", "绿")
REDCOUNT = re.compile(
    r"(?<![A-Za-z0-9-/.])(?:" + "|".join(_REDCOLORS) + r")\\s*\\d+(?!\\s*[号队楼线路军班连])"
)
'''
rep1(ANCHOR_R1, REDBLOCK)

rep1(
    "            return TRIPLE.search(line) or PROGRESS.search(line)\n",
    "            return TRIPLE.search(line) or PROGRESS.search(line) or REDCOUNT.search(line)\n",
)
rep1(
    '    """COUNTING with the two code-level skips, else TRIPLE/PROGRESS:',
    '    """COUNTING with the two code-level skips, else TRIPLE/PROGRESS/REDCOUNT:',
)
rep1(
    '    no COUNTING match falls through to TRIPLE, then PROGRESS (v2.7)."""',
    '    no COUNTING match falls through to TRIPLE, then PROGRESS (v2.7), then\n'
    '    REDCOUNT (v2.7.2, the reverse-order battery-state form 「红 N」)."""',
)

ANCHOR_R4 = '    ("零 样本（空格变体）。", True),'
i = src.index(ANCHOR_R4)
eol = src.index("\n", i) + 1
CASES_T1 = '''    # ---- v2.7.2 REDCOUNT pins (E06 「红 N」 family). BIDIRECTIONAL by
    # mandate (rule 2, the 零样本 precedent): deleting REDCOUNT flips every
    # D3-E* True; deleting/widening each named guard flips exactly its D3-G*
    # False twin. The noun-tail class is pinned member-by-member (P-Ⅲ law). ----
    ("变异电池：删成员 mergeBackendChannel → 红 5。", True),   # D3-E1: the E06 flagship carrier verbatim (the three-version FN); deleting REDCOUNT flips this
    ("变异红集：恰红 4。", True),                              # D3-E2: the E10 corpus line 「变异红集：恰红 4。」 (point-named per anchor-governance — line-order phrasing retired, obligation-④) — glued 恰 prefix; deleting REDCOUNT flips this
    ("M5 恒红 4／M6 恒红 1。", True),                           # D3-E3: 恒 prefix (measured truth-doc-B :72 form) — first match wins, one claim per line
    ("本发绿 12。", True),                                     # D3-E4: per-member pin of the 绿 state-pair member
    ("红色 3。", False),                                       # D3-G1: guard-① twin — 色 blocks the adjacency (「红色 3 个」 still flags via 个, NOT via this table)
    ("倒置→connect 7≠12 红。", False),                          # D3-G2: forward order, the battery-result reading — unreachable by word order
    ("绿灯 3 盏。", False),                                    # D3-G3: the 绿灯 compound-object reading (灯 blocks); 盏 ∉ units keeps COUNTING silent — pair of D3-E4
    ("这是红 4 号。", False),                                  # D3-G4: noun-tail member 号 (swatch/ID reading)
    ("红 2 队获胜。", False),                                  # D3-G5: member 队 (org name)
    ("红 5 楼宿舍。", False),                                   # D3-G6: member 楼 (address)
    ("红 3 线公交。", False),                                   # D3-G7: member 线 (route)
    ("红 7 路口。", False),                                     # D3-G8: member 路 (road)
    ("红 1 军番号。", False),                                    # D3-G9: member 军 (unit)
    ("全红 2 连演习明日开始。", False),                            # D3-G12: member 连 — red-team live-fire form, joined per obligation-②
    ("幼儿园红 3 班作品展出。", False),                            # D3-G13: member 班 — red-team live-fire form, joined per obligation-②
    ("坐标x红 5。", False),                                     # D3-G10: lookbehind-② pin — deleting (?<![A-Za-z0-9-/.]) flips True
    ("红 4 FAIL（复算：`grep -c FAIL /tmp/r4.log` = 4）。", False),  # D3-G11: the CMD-window twin of D3-E2 — 「红 4」+SAME-LINE grep evidence PASSES
'''
src = src[:eol] + CASES_T1 + src[eol:]

# ==================== T2: 两+个 narrative skip（E01，含空格修订） ====================
ANCHOR_W = "WINDOW = 8  # evidence-command window, ±WINDOW lines BOTH directions\n"
NARRBLOCK = ANCHOR_W + '''
# v2.7.2 (R3 收编; D3 T2; E01 「两个邻近数字」 debt, 簇⑤ CN-narrative
# family's nail on the truth table — measured claim inventory, results/
# d1-scores-v1.json: the ONLY 两+个 carriers in the matrix are the v2.6/
# v2.7.1 E01 FP tokens '两个'; every TP rides other faces: E04 '12 样本'/
# E07 '三类'/E09 '9 个'/E10 ' 1826 用例'/E08 PROGRESS. The bare frame 两+个
# is the ANAPHORIC/narrative reading (「两个邻近数字被我抄串了」 names two
# ALREADY-IDENTIFIED codes — the claim at stake is the copy error, not the
# count; that face is 坐标闸's domain, so a flag here is precision debt).
# The v2.0 adjudication RULED no char-level discriminator separates article
# 一+个 from real counts (㉝④ — that ruling STANDS; 一 is NOT cut). 两+个 is
# the narrower adjudicated face: 两 is the only CN numeral whose default
# register is colloquial-anaphoric (formal counts write 二/2 — the Arabic
# twin 「2 个」 keeps its full grammar UNTOUCHED; 共两个包 forms survive
# because 共 is consumed inside group(1), so a bare-frame fullmatch can
# never hold). REGISTERED COST (pinned D3-N3, recorded not hidden): a bare
# real count 「新增两个文件」 is silenced by adjudication; 两+次/轮/条/处/类
# OUTSIDE the frame stay live (truth-doc-A 两次/两轮 carriers).
# MEASURED SCOPE (red-team revision-②, 2026-11-16 — framing kept inside its
# evidence radius): this cut closes the recorded no-space E01 cell AND, via
# the lstrip() below, the leading-space frame; it does NOT reach the
# 两+非个/三+个 real-document flood (泛化票) nor E07 '三类'.
_CN_BARE_LIANGGE = re.compile(r"两\\s*个")
'''
rep1(ANCHOR_W, NARRBLOCK)

rep1(
    '        if m.group(0).endswith("零样本"):\n'
    "            pos = m.end()\n"
    "            continue\n"
    "        return m\n",
    '        if m.group(0).endswith("零样本"):\n'
    "            pos = m.end()\n"
    "            continue\n"
    "        # v2.7.2 skip (third code-level skip): the bare 两+个 frame is\n"
    "        # the narrative/anaphoric reading — see _CN_BARE_LIANGGE's\n"
    "        # adjudication note. skip-THEN-CONTINUE, never swallow the rest\n"
    "        # of the line (iter-5 M-2 law — the continue→return-None mutant\n"
    "        # dies on the 「…抄串了，改动 4 行」 twin pin D3-N2). lstrip():\n"
    "        # COUNTING's prefix group absorbs leading whitespace into\n"
    "        # group(0) — ' 两个' is the SAME anaphoric frame (red-team\n"
    "        # escape 2026-11-16; m-R1-2 proved spaced tokens live here;\n"
    "        # pinned both ways D3-N5/N6).\n"
    "        if _CN_BARE_LIANGGE.fullmatch(m.group(0).lstrip()):\n"
    "            pos = m.end()\n"
    "            continue\n"
    "        return m\n",
)
# T2 docstring 面：锚按 T1 改后实态适配（two→three skips + ③段前插）
rep1(
    '    """COUNTING with the two code-level skips, else TRIPLE/PROGRESS/REDCOUNT:',
    '    """COUNTING with the three code-level skips, else TRIPLE/PROGRESS/REDCOUNT:',
)
rep1(
    "    no COUNTING match falls through to TRIPLE, then PROGRESS (v2.7), then\n",
    "    ③ v2.7.2: the standalone 两+个 classifier frame (E01's CN narrative\n"
    "    family) is skipped the same way, scan continues; the frame is bare-\n"
    "    only (恰/共/总 prefix consumed in group(1); multi-char runs 一两/两百\n"
    "    never fullmatch), so prefixed and compound forms stay claims.\n"
    "    no COUNTING match falls through to TRIPLE, then PROGRESS (v2.7), then\n",
)

import re as _re
m2f_pat = _re.compile(r'    \("历时两个星期完成。", True\),.*?(?=\n)', _re.DOTALL)
assert len(m2f_pat.findall(src)) == 1, "M2-F 锚非唯一"
M2F_NEW = (
    '    ("历时两个星期完成。", False),                              # M2-F '
    'RE-ADJUDICATED by v2.7.2 (㉝④-mandated, executed openly): 两个星期 now '
    'dies at the bare 两+个 narrative skip BEFORE the duration guard is ever '
    'consulted — the residual this pin RECORDED is absorbed by the E01 cut; '
    'the 7-char class [月日年点时分秒] stays FROZEN. The silent-widening '
    'catcher\'s mandate MOVES to the 三+个+星期 twin D3-K6'
)
src = m2f_pat.sub(lambda _: M2F_NEW, src, count=1)

ANCHOR_R5 = '    ("``` 本批共 3 处差异 ```", False),\n]'
CASES_T2 = '''    ("``` 本批共 3 处差异 ```", False),
    # ---- v2.7.2 两+个 narrative-skip pins (E01 debt). BIDIRECTIONAL by
    # mandate (rule 2): deleting the skip flips D3-N1/N5; swallowing the
    # line (continue→return-None) flips D3-N2; widening past 两 is caught by
    # the EXISTING M2-B 「新增三个新文件」 pin, past 个 by D3-K3; containment-
    # rewrite by D3-K4/K5; touching Arabic/个-unit by D3-N4/N6. ----
    ("两个邻近数字被我抄串了。", False),                        # D3-N1: the E01 corpus carrier verbatim — the truth-table nail; deleting the skip returns the FP
    ("两个邻近数字被我抄串了，改动 4 行。", True),                 # D3-N2: skip-THEN-CONTINUE pin — exemption must not swallow the line's real 「4 行」
    ("新增两个文件。", False),                                  # D3-N3: REGISTERED COST face — bare real count silenced BY ADJUDICATION, recorded not hidden
    ("本批新增 2 个文件。", True),                               # D3-N4: Arabic twin stays flagged — cut is the CN token shape, not the 个 unit
    ("本次 两个 文件已通过。", False),                            # D3-N5: leading-space frame — the red-team escape form, closed by lstrip(); a nospace-only skip reintroduction flips this
    ("本次 2 个 文件已通过。", True),                             # D3-N6: spaced ARABIC twin — lstrip must not eat unit grammar
    ("修复两轮，两次到位。", True),                              # D3-K3: 两+轮/两+次 OUTSIDE the frame — widening past 个 flips this
    ("共两个包。", True),                                      # D3-K4: closure-prefix rescue rides group(1) consumption — fullmatch can never hold
    ("两百个样本。", True),                                    # D3-K5: compound run 两百 never fullmatches — the 200-count stays a claim
    ("维持三个星期不变。", True),                                # D3-K6: M2-F's relocated mandate — 三+个+星期 still depends on the frozen duration class
]'''
rep1(ANCHOR_R5, CASES_T2)

# ==================== 义务在场自校 ====================
src = sanitize(src)   # 出站卫生：私有指向清零（R4/MR 轮）

for marker, desc in [
    ("or REDCOUNT.search(line)", "T1 链尾挂载"),
    ("_REDCOLORS", "T1 色对表"),
    ('join(_REDCOLORS) + r")\\s*\\d+(?!\\s*[号队楼线路军班连])"', "义务② 闭集 regex 面（半恒真消除：注释双现不再充数）"),
    ("REGISTERED zero-carrier residuals", "义务②' ⑤残差注册注释（内环 minor 补钉）"),
    ("same-line twin carries", "G11 放行语义句（内环 minor 补钉）"),
    ("REDCOUNT (v2.7.2, the reverse-order", "docstring 链尾挂载句（内环 minor 补钉）"),
    ("M2-F's relocated mandate", "K6 迁移行（内环 minor 补钉；首跑自逮——设计期逐字句在装配面系改写形）"),
    ("carrier-corrected by A1", "M-R3-1 修复在场"),
    ("SEGMENT-ISOLATED", "义务① 注释限缩"),
    ("point-named per anchor-governance", "义务④ E2 点名"),
    ("fullmatch(m.group(0).lstrip())", "修订(a) lstrip"),
    ("D3-N5", "修订(c) 空格钉"),
    ("MEASURED SCOPE", "修订(b) 半径降格"),
    ("whole-file corpus scan", "⑤ 时代限定句（R4 重构：种子直写脱敏串，收据口径入档案措辞）"),
    ("three code-level skips", "docstring 三枚同步"),
]:
    assert marker in src, f"义务缺位: {desc}"

OUT.write_text(src, encoding="utf-8")
print(f"built {OUT.name}: {len(src)} chars, sha256={hashlib.sha256(src.encode()).hexdigest()[:16]}…")
print(f"新增行数 vs v2.7.1: {src.count(chr(10)) - SRC.count(chr(10))}")
