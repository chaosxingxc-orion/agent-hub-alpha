#!/usr/bin/env python3
"""check_citation.py — 引用核验（W06 skill 确定性脚本）

用法：
    python check_citation.py <artifact.json> <snapshot_dir>

规则（与 Java CitationChecker 同源）：
  L1 确定性层：digest/page/excerpt 三元组逐条比对——
    FABRICATED_CITATION：digest/page 定位不到快照页（BLOCKING）
    EXCERPT_DIGEST_MISMATCH：摘录与声明 digest 不符（检查后篡改，BLOCKING）
    EXCERPT_NOT_ON_PAGE：摘录未出现在声明页（BLOCKING）
    SOURCE_UNREGISTERED：引用页不在信源注册表覆盖内（BLOCKING；manifest 有 sources 时启用）
    NO_PRIMARY_SOURCE：指标页无官方一手信源覆盖（BLOCKING）
  支持度三级（确定性可判部分）：
    SUPPORTED：数字交集+实体交集+期间覆盖
    PARTIAL：相关但不完整（WARNING，建议人工）
    UNSUPPORTED：数字与实体全不匹配（BLOCKING）

退出码：0=PASSED；1=FAILED；2=WAITING_HUMAN。
"""

import json
import re
import sys
import hashlib
from pathlib import Path

ENTITIES = ("招商银行", "本集团", "营业收入", "净利润", "不良", "拨备", "净息差", "资本充足")
NUMBER_RE = re.compile(r"\d{1,3}(?:,\d{3})*(?:\.\d+)?|\d+(?:\.\d+)?")
YEAR_RE = re.compile(r"(?:19|20)\d{2}")


def load_pages(snapshot_dir: Path) -> dict:
    pages = {}
    pages_dir = snapshot_dir / "pages"
    if pages_dir.is_dir():
        for p in sorted(pages_dir.glob("page-*.txt")):
            num = int(re.sub(r"\D", "", p.name))
            pages[num] = p.read_text(encoding="utf-8")
    return pages


def load_sources(snapshot_dir: Path):
    manifest = snapshot_dir / "manifest.json"
    if not manifest.is_file():
        return []
    try:
        return json.loads(manifest.read_text(encoding="utf-8")).get("sources") or []
    except json.JSONDecodeError:
        return []


def sha256(s: str) -> str:
    return hashlib.sha256(s.encode("utf-8")).hexdigest()


def numbers_of(text: str):
    return {m.group().replace(",", "") for m in NUMBER_RE.finditer(text)}


def assess_support(claim: str, excerpt: str) -> str:
    claim_nums, excerpt_nums = numbers_of(claim), numbers_of(excerpt)
    num_overlap = bool(claim_nums) and bool(claim_nums & excerpt_nums)
    entity_overlap = (any(e in claim for e in ENTITIES)) and (any(e in excerpt for e in ENTITIES))
    if not num_overlap and not entity_overlap:
        return "UNSUPPORTED"
    claim_years = set(YEAR_RE.findall(claim))
    if claim_years and not any(y in excerpt for y in claim_years):
        return "PARTIAL"
    return "SUPPORTED"


def check(artifact: dict, pages: dict, sources: list) -> dict:
    violations = []
    citations = artifact.get("citations") or []
    seen = set()

    for c in citations:
        cid = c.get("claim_id", "")
        if cid in seen:
            violations.append({"rule": "CLAIM_ID_DUPLICATED", "target": cid,
                               "detail": "引用编号重复", "severity": "BLOCKING"})
        seen.add(cid)

    # 论断引用覆盖
    cited = set(seen)
    for sec in artifact.get("sections") or []:
        for claim_id in sec.get("claim_ids") or []:
            if claim_id not in cited:
                violations.append({"rule": "UNCITED_CLAIM", "target": claim_id,
                                   "detail": f"章节 {sec.get('section_id')} 的论断无引用记录", "severity": "BLOCKING"})

    source_pages = set()
    for src in sources:
        source_pages.update(int(p) for p in str(src.get("pages", "")).split(",") if str(p).strip().isdigit())

    for c in citations:
        cid = c.get("claim_id", "")
        page = int(c.get("page", 0))
        excerpt = c.get("evidence_excerpt") or ""
        page_text = pages.get(page)
        if page_text is None:
            violations.append({"rule": "FABRICATED_CITATION", "target": cid,
                               "detail": f"第 {page} 页无法定位快照页（伪造来源）", "severity": "BLOCKING"})
            if sources and page not in source_pages:
                violations.append({"rule": "SOURCE_UNREGISTERED", "target": cid,
                                   "detail": f"第 {page} 页未在任何已注册信源内", "severity": "BLOCKING"})
            continue
        if sha256(excerpt) != c.get("evidence_excerpt_digest"):
            violations.append({"rule": "EXCERPT_DIGEST_MISMATCH", "target": cid,
                               "detail": "证据摘录与声明 digest 不符（检查后篡改）", "severity": "BLOCKING"})
            continue
        if excerpt not in page_text:
            violations.append({"rule": "EXCERPT_NOT_ON_PAGE", "target": cid,
                               "detail": f"摘录未出现在声明的第 {page} 页", "severity": "BLOCKING"})
            continue
        # 支持度三级
        support = assess_support(c.get("claim_text") or "", excerpt)
        if support == "PARTIAL":
            violations.append({"rule": "PARTIAL_SUPPORT", "target": cid,
                               "detail": "摘录与论断相关但支撑不完整（数字命中但口径/期间未同时覆盖）→ 建议人工复核",
                               "severity": "WARNING"})
        elif support == "UNSUPPORTED":
            violations.append({"rule": "UNSUPPORTED_CLAIM", "target": cid,
                               "detail": "摘录不含论断中的任何数字或关键实体 → 不支撑", "severity": "BLOCKING"})

    # 信源注册表校验
    if sources:
        for c in citations:
            page = int(c.get("page", 0))
            if page in pages and page not in source_pages:
                violations.append({"rule": "SOURCE_UNREGISTERED", "target": c.get("claim_id", ""),
                                   "detail": f"第 {page} 页未在任何已注册信源内", "severity": "BLOCKING"})
        has_primary = any(s.get("authority") == "PRIMARY_OFFICIAL" and str(s.get("pages", "")).strip()
                          for s in sources)
        if not has_primary:
            violations.append({"rule": "NO_PRIMARY_SOURCE", "target": "metrics",
                               "detail": "信源注册表中无官方一手信源覆盖指标页；指标值禁止取自二手/聚合信源",
                               "severity": "BLOCKING"})

    blocking = [v for v in violations if v["severity"] == "BLOCKING"]
    warning = [v for v in violations if v["severity"] == "WARNING"]
    status = "FAILED" if blocking else ("WAITING_HUMAN" if warning else "PASSED")
    return {"checker_id": "citation-checker(skill)", "checker_version": "1.0.0",
            "status": status, "violations": violations}


def main():
    if len(sys.argv) != 3:
        print(__doc__)
        sys.exit(3)
    artifact = json.loads(Path(sys.argv[1]).read_text(encoding="utf-8"))
    snapshot_dir = Path(sys.argv[2])
    result = check(artifact, load_pages(snapshot_dir), load_sources(snapshot_dir))
    print(json.dumps(result, ensure_ascii=False, indent=2))
    sys.exit(1 if result["status"] == "FAILED" else (2 if result["status"] == "WAITING_HUMAN" else 0))


if __name__ == "__main__":
    main()
