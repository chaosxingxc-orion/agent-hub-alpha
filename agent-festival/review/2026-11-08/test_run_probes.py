#!/usr/bin/env python3
"""run_probes.py instrument red-probe battery (G12 conservation law, 2026-11-27).

Precedent form: xs3/test_xs3_validate.py (F7 battery — pure additive file,
runner touched zero). The fake `go` stub makes the battery hermetic: no
toolchain, deterministic, seconds-fast. The live overlay path stays covered
by the adjudication record (R0-R3b in INSTRUMENT-ADJUDICATION-G12.md).

Run:  python3 -I agent-festival/review/2026-11-08/test_run_probes.py
All PASS -> rc=0; any red -> rc=1.

Mutation coupling (kill list, full table in the adjudication record):
  drop the missing-probe guard        -> S1/S6 red
  drop the zero-total guard           -> S2/S7 red (message assertions are
                                         load-bearing: the rc would else stay 3)
  drop the empty-set guard            -> S10 red
  invert the top-level filter         -> S1/S3/S4/S6/S7/S10 red
  instrument exit 3 -> 0              -> S1/S2/S6/S7/S10 red
  instrument branch masks rc!=0       -> S5 red
  final pass-through exit -> 3        -> S3/S4 red
  drop stream replay (stdout swallow) -> S3 red
  drop go_root .resolve()             -> S11 red (key/cwd split under a
                                         symlinked agent-festival/go)
  narrow PROBE_DEF parse breadth      -> S12 red (expected set must cover the
                                         whole injected family, not one probe)
  replay only RUN lines (swallow PASS/FAIL transcript) -> S3 red (byte-exact)
  drop stderr merge (stderr=DEVNULL)  -> S13 red (notice must reach the stream)
  drop a README evidence-table row    -> S8 red (table is the runtime anchor)
Full kill matrix incl. logs: record §4.
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
GO_ROOT = (HERE.parents[1] / "go").resolve()
PROBE_FILE = HERE / "system_contract_probes.go.txt"

README = HERE / "README.md"


def probe_family_from_readme():
    """Names straight from the README evidence table, parsed at runtime —
    no transcription list to drift (review R1/R2 finding family closed by
    structure: the table itself is the anchor)."""
    rows = re.findall(r"^\| ([A-Za-z][A-Za-z0-9]*) \|",
                      README.read_text(encoding="utf-8"), re.MULTILINE)
    return {"TestNGAReview_" + n for n in rows}

STUB = """#!/usr/bin/env python3
import os, shutil, sys
dump = os.environ["G12_STUB_DUMP"]
ov = next(a for a in sys.argv if a.startswith("-overlay="))
shutil.copy(ov.split("=", 1)[1], os.path.join(dump, "overlay.json"))
open(os.path.join(dump, "cwd.txt"), "w").write(os.getcwd())
open(os.path.join(dump, "argv.txt"), "w").write("\\n".join(sys.argv))
sys.stdout.write(os.environ["G12_STUB_OUT"])
if os.environ.get("G12_STUB_ERR"):
    sys.stderr.write(os.environ["G12_STUB_ERR"])
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


def run_probe_case(lines, stub_rc, runner=RUNNER, probe_text=None, go_symlink=False,
                   stub_stderr=None):
    """Invoke run_probes.py with a fake go; return (proc, expected_overlay_root).

    probe_text/go_symlink build a synthetic same-depth tree in /tmp (the repo
    is never written): go_symlink replaces agent-festival/go with a symlink,
    reproducing the review-loop R3 topology where an unresolved go_root splits
    the overlay key (lexical) from the child cwd (physical).
    """
    with tempfile.TemporaryDirectory(prefix="g12-battery-") as dump:
        stub = Path(dump) / "go-stub"
        stub.write_text(STUB)
        stub.chmod(0o755)
        env = dict(os.environ)
        env.update(G12_STUB_DUMP=dump, G12_STUB_OUT=lines, G12_STUB_RC=str(stub_rc))
        if stub_stderr:
            env["G12_STUB_ERR"] = stub_stderr
        target, expected_root = runner, GO_ROOT
        if probe_text is not None or go_symlink:
            root = Path(dump) / "agent-festival"
            leaf = root / "review" / "syn"
            leaf.mkdir(parents=True)
            expected_root = root / "go"
            if go_symlink:
                phys = Path(dump) / "phys-go"
                phys.mkdir()
                expected_root.symlink_to(phys)
                expected_root = phys
            else:
                expected_root.mkdir()
            (leaf / "run_probes.py").write_text(Path(runner).read_text())
            (leaf / "system_contract_probes.go.txt").write_text(
                probe_text if probe_text is not None else PROBE_FILE.read_text())
            target = leaf / "run_probes.py"
        proc = subprocess.run(
            [sys.executable, "-I", str(target), str(stub)],
            capture_output=True, text=True, env=env)
        files = {p.name: Path(dump, p.name).read_text()
                 for p in Path(dump).iterdir()
                 if p.name.endswith(".txt") or p.name == "overlay.json"}
        proc.dump = files
        proc.expected_root = expected_root
        return proc


def check(results, name, cond, detail=""):
    results.append((name, cond, detail))
    print("%s %s%s" % ("PASS" if cond else "FAIL", name,
                       "" if cond else " -- " + detail))


def main():
    results = []
    extra = ["TestP0A_ScanRegenerated", "TestXS2_MachineInternalFormImmune"]
    file_defs = set(re.findall(
        r"^func (TestNGAReview_\w+)\(",
        PROBE_FILE.read_text(encoding="utf-8"), re.MULTILINE))
    FAMILY = sorted(file_defs)
    full = stream(FAMILY + extra)

    p3 = run_probe_case(full, 0)
    check(results, "S3 full-probe stream passes (positive control, rc=0) and the "
                   "go stream is replayed to stdout byte-exactly",
          p3.returncode == 0 and "instrument failure" not in p3.stderr
          and p3.stdout == full,
          "rc=%s stderr=%r stdout_len=%d vs %d" % (p3.returncode, p3.stderr[:200],
                                                   len(p3.stdout), len(full)))

    p13 = run_probe_case(stream(FAMILY + extra), 0, stub_stderr="go build notice\n")
    check(results, "S13 child stderr is merged into the replayed stream (evidence "
                   "transcript survives as one merged stdout; no swallow)",
          p13.returncode == 0 and "go build notice" in p13.stdout
          and "go build notice" not in p13.stderr,
          "rc=%s stdout_tail=%r stderr=%r" % (p13.returncode, p13.stdout[-120:],
                                              p13.stderr[:120]))

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

    absent = FAMILY[0]
    six = FAMILY[1:]
    p6 = run_probe_case(stream(six), 0)
    check(results, "S6 one injected probe missing from run -> red with name",
          p6.returncode == 3 and absent in p6.stderr,
          "rc=%s stderr=%r" % (p6.returncode, p6.stderr[:200]))

    p7 = run_probe_case("=== RUN   TestP0A_SixCellMatrix/cell1\n"
                        "--- PASS: TestP0A_SixCellMatrix/cell1 (0.00s)\n"
                        "PASS\n", 0)
    check(results, "S7 subtest-only stream is not evidence of top-level runs",
          p7.returncode == 3 and "zero tests matched" in p7.stderr,
          "rc=%s stderr=%r" % (p7.returncode, p7.stderr[:200]))

    readme_defs = probe_family_from_readme()
    check(results, "S8 probe file equals README evidence-table family (README parsed "
                   "at runtime; transcription list abolished)",
          file_defs == readme_defs and len(file_defs) >= 1,
          "diff=%r readme_n=%d file_n=%d" % (file_defs ^ readme_defs,
                                             len(readme_defs), len(file_defs)))

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

    p11 = run_probe_case(full, 0, go_symlink=True)
    key11 = list(json.loads(p11.dump["overlay.json"])["Replace"])
    check(results, "S11 symlinked agent-festival/go: key and child cwd land on the "
                   "physical root (go_root .resolve() normalization)",
          p11.returncode == 0
          and key11 == [str(p11.expected_root
                            / "internal/e2e/nga_external_review_test.go")]
          and p11.dump["cwd.txt"] == str(p11.expected_root),
          "rc=%s key=%r cwd=%r phys=%r" % (p11.returncode, key11,
                                           p11.dump["cwd.txt"],
                                           str(p11.expected_root)))

    p11b = run_probe_case(stream(extra), 0, go_symlink=True)
    check(results, "S11b injection-dead stream under go-symlink topology still reds "
                   "via the family guard",
          p11b.returncode == 3 and "never ran" in p11b.stderr,
          "rc=%s stderr=%r" % (p11b.returncode, p11b.stderr[:200]))

    one_only = stream([FAMILY[0]] + extra)
    p12 = run_probe_case(one_only, 0)
    missing_names = FAMILY[1:]
    check(results, "S12 all-but-one probes absent while one runs -> red listing them "
                   "(pins PROBE_DEF parse breadth)",
          p12.returncode == 3
          and all(n in p12.stderr for n in missing_names)
          and p12.stderr.count("TestNGAReview_") >= len(missing_names),
          "rc=%s stderr=%r" % (p12.returncode, p12.stderr[:220]))

    failed = [n for n, ok, _ in results if not ok]
    print("%d/%d PASS" % (len(results) - len(failed), len(results)))
    return 1 if failed else 0


if __name__ == "__main__":
    sys.exit(main())
