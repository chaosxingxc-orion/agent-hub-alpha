#!/usr/bin/env python3
"""run_probes.py instrument red-probe battery (G12 conservation law, 2026-11-27).

Precedent form: xs3/test_xs3_validate.py (F7 battery — pure additive file,
runner touched zero). The fake `go` stub makes the battery hermetic: no
toolchain, deterministic, seconds-fast. The live overlay path stays covered
by the adjudication record (R0-R3 in INSTRUMENT-ADJUDICATION-G12.md).

Run:  python3 -I agent-festival/review/2026-11-08/test_run_probes.py
All PASS -> rc=0; any red -> rc=1.

Mutation coupling (kill list lives in the adjudication record):
  drop the missing-probe guard   -> S1/S6 red
  drop the zero-total guard      -> S2/S7 red (message assertions are load-
                                    bearing: both branches would else exit 3)
  drop the empty-set guard       -> S10 red
  invert the top-level filter    -> S3/S7 red
  neutralize the instrument exit -> S1/S2/S6/S7/S10 red (rc 3 -> 0)
  unconditional exit rewrite     -> S4/S5 red (real rc must survive)
"""
import json
import os
import re
import subprocess
import sys
import tempfile
from pathlib import Path

HERE = Path(__file__).resolve().parent
RUNNER = HERE / "run_probes.py"
GO_ROOT = HERE.parents[1] / "go"
PROBE_FILE = HERE / "system_contract_probes.go.txt"

# README.md evidence-table names, transcribed independently of the runner's
# regex (instrument-side anchor: file drift vs README table must go red here).
FAMILY = [
    "TestNGAReview_RecentShortOverdueMustNotPassUpperGrant",
    "TestNGAReview_SupersededDecisionMustNotDispatch",
    "TestNGAReview_UnadoptedResultMustNotBecomeCurrent",
    "TestNGAReview_VerifierRefusalNeedsDurableEvidence",
    "TestNGAReview_OriginalLowBandRule",
    "TestNGAReview_RequiredWindowCannotDisappear",
    "TestNGAReview_ExpectedEpochIsNotDecisionAuthority",
]

STUB = """#!/usr/bin/env python3
import os, shutil, sys
dump = os.environ["G12_STUB_DUMP"]
ov = next(a for a in sys.argv if a.startswith("-overlay="))
shutil.copy(ov.split("=", 1)[1], os.path.join(dump, "overlay.json"))
open(os.path.join(dump, "cwd.txt"), "w").write(os.getcwd())
open(os.path.join(dump, "argv.txt"), "w").write("\\n".join(sys.argv))
sys.stdout.write(os.environ["G12_STUB_OUT"])
sys.exit(int(os.environ["G12_STUB_RC"]))
"""


def stream(names, extra_lines=()):
    out = []
    for name in names:
        out.append("=== RUN   %s" % name)
        out.append("--- PASS: %s (0.00s)" % name)
    out.extend(extra_lines)
    out.append("PASS")
    out.append("ok  \tgithub.com/openjiuwen/agent-festival/go/internal/e2e\t0.02s")
    return "\n".join(out) + "\n"


def run_probe_case(lines, stub_rc, runner=RUNNER, probe_text=None):
    """Invoke run_probes.py with a fake go; return (proc, dump_path)."""
    with tempfile.TemporaryDirectory(prefix="g12-battery-") as dump:
        stub = Path(dump) / "go"
        stub.write_text(STUB)
        stub.chmod(0o755)
        env = dict(os.environ)
        env.update(G12_STUB_DUMP=dump, G12_STUB_OUT=lines, G12_STUB_RC=str(stub_rc))
        target = runner
        if probe_text is not None:  # synthetic tree with a custom probe file
            root = Path(dump) / "agent-festival"
            leaf = root / "review" / "syn"
            leaf.mkdir(parents=True)
            (root / "go").mkdir()
            (leaf / "run_probes.py").write_text(Path(runner).read_text())
            (leaf / "system_contract_probes.go.txt").write_text(probe_text)
            target = leaf / "run_probes.py"
        proc = subprocess.run(
            [sys.executable, "-I", str(target), str(stub)],
            capture_output=True, text=True, env=env)
        files = {p.name: Path(dump, p.name).read_text()
                 for p in Path(dump).iterdir() if p.name.endswith(".txt")
                 or p.name == "overlay.json"}
        proc.dump = files
        return proc


def check(results, name, cond, detail=""):
    results.append((name, cond, detail))
    print("%s %s%s" % ("PASS" if cond else "FAIL", name,
                       "" if cond else " -- " + detail))


def main():
    results = []
    full = stream(FAMILY + ["TestP0A_ScanRegenerated", "TestXS2_MachineInternalFormImmune"])

    p3 = run_probe_case(full, 0)
    check(results, "S3 full-probe stream passes (positive control, rc=0)",
          p3.returncode == 0 and "instrument failure" not in p3.stderr,
          "rc=%s stderr=%r" % (p3.returncode, p3.stderr[:200]))

    p1 = run_probe_case(stream(["TestP0A_ScanRegenerated", "TestP0A_SixCellMatrix",
                                "TestXS2_MachineInternalFormImmune",
                                "TestXS2_ParameterFormTrustHole"]), 0)
    check(results, "S1 overlay-ineffective shape (record R2): rc=0 must become 3",
          p1.returncode == 3 and "never ran" in p1.stderr,
          "rc=%s stderr=%r" % (p1.returncode, p1.stderr[:200]))

    p2 = run_probe_case("testing: warning: no tests to run\n"
                        "PASS\nok  \tgithub.com/openjiuwen/agent-festival/go"
                        "/internal/e2e\t0.01s [no tests to run]\n", 0)
    check(results, "S2 zero-tests-matched shape (record R1): rc=0 must become 3",
          p2.returncode == 3 and "zero tests matched" in p2.stderr,
          "rc=%s stderr=%r" % (p2.returncode, p2.stderr[:200]))

    p4 = run_probe_case(stream(FAMILY, ["--- FAIL: "
                                        "TestNGAReview_SupersededDecisionMustNotDispatch (0.00s)"]),
                        1)
    check(results, "S4 genuine probe failure propagates rc=1 (never rewritten)",
          p4.returncode == 1 and "instrument failure" not in p4.stderr,
          "rc=%s" % p4.returncode)

    p5 = run_probe_case("", 1)
    check(results, "S5 build-failure shape: nonzero rc passes through, no mask",
          p5.returncode == 1, "rc=%s" % p5.returncode)

    six = [n for n in FAMILY if n != "TestNGAReview_OriginalLowBandRule"]
    p6 = run_probe_case(stream(six), 0)
    check(results, "S6 one injected probe missing from run -> red with name",
          p6.returncode == 3 and "TestNGAReview_OriginalLowBandRule" in p6.stderr,
          "rc=%s stderr=%r" % (p6.returncode, p6.stderr[:200]))

    p7 = run_probe_case("=== RUN   TestP0A_SixCellMatrix/cell1\n"
                        "--- PASS: TestP0A_SixCellMatrix/cell1 (0.00s)\n"
                        "PASS\n", 0)
    check(results, "S7 subtest-only stream is not evidence of top-level runs",
          p7.returncode == 3 and "zero tests matched" in p7.stderr,
          "rc=%s stderr=%r" % (p7.returncode, p7.stderr[:200]))

    defs = set(re.findall(r"^func (TestNGAReview_\w+)\(",
                          PROBE_FILE.read_text(encoding="utf-8"), re.MULTILINE))
    check(results, "S8 probe file equals README evidence-table family",
          defs == set(FAMILY), "diff=%r" % (defs ^ set(FAMILY)))

    argv = p3.dump["argv.txt"].splitlines()
    overlay_key = list(json.loads(p3.dump["overlay.json"])["Replace"])
    cwd = p3.dump["cwd.txt"]
    check(results, "S9 go-test call semantics + overlay key/cwd share one root",
          argv[1:] == ["test", argv[2], "./internal/e2e", "-run",
                       "TestNGAReview_|TestP0A|TestXS2", "-count=1", "-v"]
          and overlay_key == [str(GO_ROOT / "internal/e2e/nga_external_review_test.go")]
          and cwd == str(GO_ROOT) and overlay_key[0].startswith(cwd),
          "argv=%r key=%r cwd=%r" % (argv, overlay_key, cwd))

    p10 = run_probe_case(full, 0, probe_text="package e2e\n")
    check(results, "S10 injected file defining no probes is instrument failure",
          p10.returncode == 3 and "no TestNGAReview_ functions" in p10.stderr,
          "rc=%s stderr=%r" % (p10.returncode, p10.stderr[:200]))

    failed = [n for n, ok, _ in results if not ok]
    print("%d/%d PASS" % (len(results) - len(failed), len(results)))
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
