"""Run external contract probes without adding failing tests to the normal suite.

Nonzero is intentionally preserved while system acceptance items remain open.
Use an exact source checkout with LF fixtures; no model or business API is used.

Instrument self-check (G12 adjudication 2026-11-27): the -v stream is counted
and a passing rc=0 is accepted only when every TestNGAReview_* probe defined in
the injected file actually ran. Zero tests matched, an empty injected probe set,
or missing probes exit 3 (instrument failure). Any nonzero go test exit code is
still propagated unchanged: open acceptance items stay red for their own reason.
go_root is resolved: a symlinked agent-festival/go otherwise keeps the overlay
key on the lexical path while the child cwd is physical, silently dropping the
injection (review-loop R3 finding; the family guard alone would only mask that
topology as a misleading rc=3).
"""
import json
from pathlib import Path
import re
import subprocess
import sys
import tempfile

here = Path(__file__).resolve().parent
go_root = (here.parents[1] / "go").resolve()
go = sys.argv[1] if len(sys.argv) > 1 else "go"
PROBE_DEF = re.compile(r"^func (TestNGAReview_\w+)\(", re.MULTILINE)
RUN_LINE = re.compile(r"^=== RUN\s+(\S+)")
with tempfile.TemporaryDirectory(prefix="nga-contract-review-") as tmp:
    probe_bytes = (here / "system_contract_probes.go.txt").read_bytes()
    source = Path(tmp) / "review_test.go"
    source.write_bytes(probe_bytes)
    expected = set(PROBE_DEF.findall(probe_bytes.decode("utf-8")))
    overlay = Path(tmp) / "overlay.json"
    overlay.write_text(json.dumps({"Replace": {
        str(go_root / "internal/e2e/nga_external_review_test.go"): str(source)
    }}), encoding="utf-8")
    result = subprocess.run([
        go, "test", "-overlay=" + str(overlay), "./internal/e2e",
        "-run", "TestNGAReview_|TestP0A|TestXS2", "-count=1", "-v"
    ], cwd=go_root, stdout=subprocess.PIPE, stderr=subprocess.STDOUT, text=True)
    total = 0
    seen = set()
    for line in (result.stdout or "").splitlines(True):
        sys.stdout.write(line)
        sys.stdout.flush()
        match = RUN_LINE.match(line)
        if match and "/" not in match.group(1):
            total += 1
            seen.add(match.group(1))
    reason = None
    if total == 0:
        reason = "zero tests matched the run set"
    elif not expected:
        reason = "injected probe file defines no TestNGAReview_ functions"
    elif expected - seen:
        reason = "injected probes never ran: " + ", ".join(sorted(expected - seen))
    if reason is not None:
        sys.stderr.write(
            "run_probes: instrument failure: %s -- rc=0 without executed "
            "probes is a false green, not an acceptance pass\n" % reason)
        raise SystemExit(result.returncode if result.returncode != 0 else 3)
    raise SystemExit(result.returncode)
