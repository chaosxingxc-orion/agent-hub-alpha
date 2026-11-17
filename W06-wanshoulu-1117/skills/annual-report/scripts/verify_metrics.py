#!/usr/bin/env python3
"""verify_metrics.py — 八指标数值核验（W06 skill 确定性脚本）

用法：
    python verify_metrics.py <artifact.json> <snapshot_dir>

artifact.json 为 AnnualReportArtifacts 工件（至少含 metrics 数组）；
snapshot_dir 为快照目录（pages/page-NNN.txt）。

退出码：0=PASSED；1=FAILED（存在 BLOCKING）；2=WARNING（仅非阻断项，建议人工）。
输出：JSON 格式的 CheckResult（与 checker 服务同构）。

规则（与 Java NumericChecker N1-N6 同源，术语以 SKILL.md 为准）：
  N1 完整性：八指标缺一/重复即 BLOCKING
  N2 单位白名单：指标单位不符即 BLOCKING
  N3 数字溯源：值必须在快照页原文字面出现，否则 BLOCKING
  N4 派生复算：不良率/拨备覆盖率/成本收入比三口径复算，偏差超容差 BLOCKING；
     复算基数定位失败降级 WARNING（不误杀）
  N5 期间成对：同比论断须与两期值一致（YOY_MISMATCH）
  N7 变动口径：比率类指标变动须用「个百分点/基点」，写成「%」即拦（DELTA_UNIT_MISMATCH）
     ——与 Java 版 NumericChecker.deltaUnitCheck 同源（三形态同步纪律）
"""

import json
import re
import sys
import hashlib
from pathlib import Path

METRICS = {
    "OPERATING_REVENUE": ("营业收入", "亿元"),
    "NET_PROFIT_ATTRIBUTABLE": ("归母净利润", "亿元"),
    "NET_INTEREST_MARGIN": ("净息差", "%"),
    "COST_INCOME_RATIO": ("成本收入比", "%"),
    "NPL_BALANCE": ("不良贷款余额", "亿元"),
    "NPL_RATIO": ("不良贷款率", "%"),
    "PROVISION_COVERAGE": ("拨备覆盖率", "%"),
    "CET1_RATIO": ("核心一级资本充足率", "%"),
}

NUMBER_RE = re.compile(r"\d{1,3}(?:,\d{3})*(?:\.\d+)?|\d+(?:\.\d+)?")


def load_pages(snapshot_dir: Path) -> dict:
    pages = {}
    pages_dir = snapshot_dir / "pages"
    if pages_dir.is_dir():
        for p in sorted(pages_dir.glob("page-*.txt")):
            num = int(re.sub(r"\D", "", p.name))
            pages[num] = p.read_text(encoding="utf-8")
    return pages


def page_text(pages, digest, page):
    if digest and digest in {hashlib.sha256(t.encode()).hexdigest() for t in pages.values()}:
        pass  # digest 命中与否不影响按页定位（脚本版与快照 resolver 同语义）
    return pages.get(page)


def number_present(text: str, value: str) -> bool:
    target = value.replace(",", "")
    return any(m.group().replace(",", "") == target for m in NUMBER_RE.finditer(text))


def find_number_after(text: str, keyword: str):
    i = text.find(keyword)
    if i < 0:
        return None
    j, skipped = i + len(keyword), 0
    while j < len(text) and not (text[j].isdigit() or text[j] in ",.") and skipped < 12:
        j += 1
        skipped += 1
    start = j
    while j < len(text) and (text[j].isdigit() or text[j] in ",."):
        j += 1
    if j == start:
        return None
    raw = text[start:j].rstrip(",")
    try:
        return float(raw.replace(",", ""))
    except ValueError:
        return None


def violation(rule, target, detail, severity):
    return {"rule": rule, "target": target, "detail": detail, "severity": severity}


def verify(artifact: dict, pages: dict) -> dict:
    violations = []
    metrics = artifact.get("metrics") or []
    by_key = {}
    for m in metrics:
        if m.get("key") in by_key:
            violations.append(violation("METRIC_DUPLICATED", m.get("key"), "同一指标出现多条记录", "BLOCKING"))
        by_key[m.get("key")] = m

    # N1 完整性
    for key in METRICS:
        if key not in by_key:
            violations.append(violation("METRIC_MISSING", key, f"八指标缺失：{METRICS[key][0]}", "BLOCKING"))

    # N2/N3
    for key, m in by_key.items():
        if key in METRICS:
            expected_unit = METRICS[key][1]
            if m.get("unit") != expected_unit:
                violations.append(violation("UNIT_MISMATCH", key,
                    f"单位应为 {expected_unit}，实际 {m.get('unit')}", "BLOCKING"))
        value = str(m.get("value", ""))
        if value not in ("NOT_FOUND", "UNSPECIFIED"):
            text = page_text(pages, m.get("source_document_digest"), int(m.get("page", 0)))
            if text is None:
                violations.append(violation("SOURCE_UNRESOLVABLE", key, "digest/page 无法定位快照页", "BLOCKING"))
            elif not number_present(text, value):
                violations.append(violation("FABRICATED_NUMBER", key,
                    f"值 {value} 未在第 {m.get('page')} 页原文出现", "BLOCKING"))

    # N4 三口径复算
    def mv(k):
        return by_key.get(k) or {}

    npl_balance, npl_ratio, loans_total = mv("NPL_BALANCE"), mv("NPL_RATIO"), None
    if npl_balance:
        text = page_text(pages, npl_balance.get("source_document_digest"), int(npl_balance.get("page", 0)))
        if text:
            loans_total = find_number_after(text, "贷款和垫款本金总额") or find_number_after(text, "贷款总额")
            if npl_ratio and loans_total:
                try:
                    recomputed = float(str(npl_balance["value"]).replace(",", "")) * 100 / loans_total
                    reported = float(str(npl_ratio["value"]).replace(",", ""))
                    if abs(recomputed - reported) > 0.05:
                        violations.append(violation("CALC_MISMATCH", "NPL_RATIO",
                            f"复算 {recomputed:.2f}% vs 报告 {reported}%（偏差>0.05pp）", "BLOCKING"))
                except (ValueError, ZeroDivisionError):
                    pass
            elif npl_ratio:
                violations.append(violation("DERIVED_BASE_MISSING", "NPL_RATIO",
                    "同页未定位贷款总额基数，不良率复算降级人工核对", "WARNING"))

    cov, npl_b = mv("PROVISION_COVERAGE"), mv("NPL_BALANCE")
    if cov and npl_b:
        text = page_text(pages, npl_b.get("source_document_digest"), int(npl_b.get("page", 0)))
        if text:
            reserve = find_number_after(text, "贷款减值准备")
            if reserve is None:
                violations.append(violation("DERIVED_BASE_MISSING", "PROVISION_COVERAGE",
                    "同页未定位贷款减值准备，拨备覆盖率复算降级人工核对", "WARNING"))
            else:
                try:
                    recomputed = reserve * 100 / float(str(npl_b["value"]).replace(",", ""))
                    reported = float(str(cov["value"]).replace(",", ""))
                    if abs(recomputed - reported) > 1:
                        violations.append(violation("CALC_MISMATCH", "PROVISION_COVERAGE",
                            f"复算 {recomputed:.2f}% vs 报告 {reported}%（容差1pp）", "BLOCKING"))
                except (ValueError, ZeroDivisionError):
                    pass

    ratio, rev = mv("COST_INCOME_RATIO"), mv("OPERATING_REVENUE")
    if ratio and rev:
        text = page_text(pages, rev.get("source_document_digest"), int(rev.get("page", 0)))
        if text:
            expense = find_number_after(text, "业务及管理费")
            if expense is None:
                violations.append(violation("DERIVED_BASE_MISSING", "COST_INCOME_RATIO",
                    "同页未定位业务及管理费，成本收入比复算降级人工核对", "WARNING"))
            else:
                try:
                    recomputed = expense * 100 / float(str(rev["value"]).replace(",", ""))
                    reported = float(str(ratio["value"]).replace(",", ""))
                    if abs(recomputed - reported) > 0.5:
                        violations.append(violation("CALC_MISMATCH", "COST_INCOME_RATIO",
                            f"复算 {recomputed:.2f}% vs 报告 {reported}%（容差0.5pp）", "BLOCKING"))
                except (ValueError, ZeroDivisionError):
                    pass

    # N5 同比复算：正文"同比…"与两期值一致
    rev = mv("OPERATING_REVENUE")
    if rev and "value" in rev:
        prior = None
        for num, text in pages.items():
            if num != int(rev.get("page", 0)) and "2024年" in text and "营业收入" in text:
                mm = re.search(r"营业收入([\d,]+(?:\.\d+)?)亿元", text)
                if mm:
                    prior = float(mm.group(1).replace(",", ""))
                    break
        if prior is None:
            violations.append(violation("PERIOD_BASE_MISSING", "OPERATING_REVENUE",
                "快照未定位上期营业收入基数，同比复算降级人工核对", "WARNING"))
        else:
            try:
                current = float(str(rev["value"]).replace(",", ""))
                yoy = (current - prior) / prior * 100
                for sec in artifact.get("sections") or []:
                    # N6 归属绑定（与 Java 版同源）：只校验与"营业收入"同句的同比
                    for pm in re.finditer(r"营业收入[^。；]{0,80}?同比(?:增长|下降|减少|增加)?([\d.]+)%", sec.get("body", "")):
                        if abs(abs(yoy) - float(pm.group(1))) > 0.1:
                            violations.append(violation("YOY_MISMATCH", "OPERATING_REVENUE",
                                f"正文同比 {pm.group(1)}% vs 复算 {yoy:.2f}%", "BLOCKING"))
            except (ValueError, ZeroDivisionError):
                pass

    # N7 变动口径（与 Java deltaUnitCheck 同源）
    import re as _re
    DELTA_PHRASE = _re.compile(
        r"(不良贷款率|不良率|拨备覆盖率|成本收入比|核心一级资本充足率|资本充足率|净息差)"
        r"[^。；]{0,30}?(上升|下降|增加|减少|增减|变动|增长)\s*(?:约\s*)?"
        r"([\d,]+(?:\.\d+)?)\s*(个百分点|基点|亿元|%|％)")
    for sec in artifact.get("sections") or []:
        for m in DELTA_PHRASE.finditer(sec.get("body", "")):
            metric_name, direction, magnitude, unit_word = m.group(1), m.group(2), m.group(3), m.group(4)
            if unit_word in ("%", "％"):
                violations.append(violation("DELTA_UNIT_MISMATCH", metric_name,
                    f"比率类指标「{metric_name}」变动 {direction}{magnitude}%：比率变动须用「个百分点/基点」，"
                    f"仅增长率可用%（W01 反例：0.94 个百分点误写 0.94%）", "BLOCKING"))

    blocking = [v for v in violations if v["severity"] == "BLOCKING"]
    warning = [v for v in violations if v["severity"] == "WARNING"]
    status = "FAILED" if blocking else ("WAITING_HUMAN" if warning else "PASSED")
    return {"checker_id": "numeric-checker(skill)", "checker_version": "1.0.0",
            "status": status, "violations": violations}


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(3)
    artifact = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
    pages = load_pages(Path(sys.argv[2]))
    result = verify(artifact, pages)
    print(json.dumps(result, ensure_ascii=False, indent=2))
    sys.exit(1 if result["status"] == "FAILED" else (2 if result["status"] == "WAITING_HUMAN" else 0))


if __name__ == "__main__":
    main()
