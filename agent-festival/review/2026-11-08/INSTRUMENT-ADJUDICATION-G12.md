# G12 · run_probes.py instrument adjudication record (2026-11-27)

Scope per 169 §五 G12: RED-BEFORE adjudication first, fix only the adjudicated
disease, mutation re-verify. Machine: aws-claw (Linux), Go go1.24.0, repo
`main` = `436570b42`, branch `fix/probe-runner-fail-loud`.

## 1. Baseline (R0) — instrument works on Linux

Command: `python3 agent-festival/review/2026-11-08/run_probes.py /home/yaojun/go-dist/go/bin/go`

Measured: `rc=1`, 20 `=== RUN` lines (7 `TestNGAReview_*` probes + 3 P0A
top-level + 8 subtests + 2 XS2); the 4 FAILs match the README evidence table
(Superseded/Unadopted/VerifierRefusal/RequiredWindow). The overlay injection
itself is effective on this machine.

## 2. RED scenarios (pre-fix copies, main tree untouched)

### R1 — `-run` zero-match (CONFIRMED false green)

Copy at same tree depth, pattern `"TestNGAReview_|TestP0A|TestXS2"` replaced by
`"NoSuchPatternXYZ"`. Measured output tail:

```
testing: warning: no tests to run
PASS
ok  	github.com/openjiuwen/agent-festival/go/internal/e2e	0.007s [no tests to run]
R1_RC=0
```

Zero tests ran; the old runner propagated rc=0 — a "probe suite passed" signal
with nothing probed.

### R2 — overlay-ineffective shape (CONFIRMED false green, family-granular)

Copy with the overlay `Replace` map emptied (simulates key drift / lost
injection). Measured: 13 `=== RUN` lines, all P0A/XS2, **0** `TestNGAReview_*`,
final `PASS`, `rc=0`. The known-red acceptance probes vanished and the suite
reported green.

Consequence for fix design: a total-count-only zero check does NOT close this
form (total was 13). Detection must be probe-family-granular.

### R3 — symlinked checkout invocation (E-report key-divergence REFUTED)

`ln -sfn ~/nga-handoff/agent-hub-wt /tmp/nga-sym-repo`, then run the real
script through the symlink path. Measured: `rc=1`, 7 probes ran — byte-identical
behavior to R0. The overlay key and the `cwd` both derive from the same
`Path(__file__).resolve()` root, so the E-report's `/private/tmp` vs `/tmp`
key-divergence mechanism does not exist in this script (matches the 169 adversarial
note; re-verified independently here). **No key-normalization change was made**
— per the explicit rule: no fix for an unadjudicated etiology.

## 3. The fix (run_probes.py)

- Capture `go test -v` output (stdout+stderr merged) and replay it, counting
  top-level `=== RUN` names (no `/`) as `total`.
- `expected` = set of `func TestNGAReview_*(` names parsed from the injected
  probe file itself (instrument-side anchor; no hardcoded count — adding a
  probe keeps working, renaming within the prefix must run).
- Three instrument conditions: `total == 0` ("zero tests matched"), empty
  `expected`, or `expected - seen` non-empty (names listed). On any: stderr
  `run_probes: instrument failure: ...`, exit `3` when go returned 0.
- Any nonzero go-test code is propagated unchanged (open acceptance items stay
  red for their own reason; a real failure is never rewritten to 3).
- go test call semantics unchanged (flag list byte-identical, pinned by S9).

Post-fix retests (real go): R1-shape `rc=3` with `zero tests matched`;
R2-shape `rc=3` listing all 7 probe names; live baseline `rc=1` unchanged,
7 probes still running and still honestly red.

## 4. Mutation table (all on committed state 19afcf2, detached worktree)

| ID | Mutation | Battery cases turned red | Result |
| --- | --- | --- | --- |
| M1 | drop `expected - seen` guard | S1, S6 | killed |
| M2 | drop `total == 0` guard | S2, S7 (message assertions) | killed |
| M3 | drop empty-`expected` guard | S10 | killed |
| M4 | invert top-level `/` filter | S1, S3, S4, S6, S7, S10 | killed |
| M5 | instrument exit 3 -> 0 | S1, S2, S6, S7, S10 | killed |
| M6 | instrument branch masks rc!=0 to 3 | S5 | killed |
| M6b | final pass-through exit -> 3 | S3, S4 | killed |

Full revert (checkout of the pre-fix main version of the script) revives both
real disease shapes: R1-shape `rc=0`, R2-shape `rc=0` — the fix is load-bearing,
not decoration. Every guard in the fix has ≥1 killing battery case
(`test_run_probes.py`, hermetic fake-`go` stubs, precedent
`xs3/test_xs3_validate.py` F7 battery).

## 5. Regression

- `go test -count=1 ./...` → `FULL_RC=0`, 21/21 packages ok (none dropped).
- `go test -count=1 -race ./...` → `RACE_RC=0`.
- `go test -count=1 -shuffle=on ./...` → `SHUF_RC=0`.
- `python3 -I agent-festival/review/2026-11-08/test_run_probes.py` → `rc=0`,
  10/10 PASS (test surface 0 → 10 executable pins for this instrument).

## 6. Deviations from the estimated "one file + one record" shape

- `test_run_probes.py` added: the campaign conservation law (every fix ships a
  permanent probe that turns red when the fix is removed) overrides the size
  estimate; precedent `xs3/test_xs3_validate.py` added a battery the same way.
- README one-sentence update: the behavior change made the old "returns the
  actual test exit code" sentence incomplete for the rc=3 case; outbound doc
  truthfulness kept in the same commit.

## 7. Documented boundaries (not closed by this instrument, honest scope)

- Deleting probe functions outright, or renaming one away from the
  `TestNGAReview_` prefix, shrinks `expected` in lockstep and stays silent —
  that drift belongs to the review surface (diff of `system_contract_probes.go.txt`
  vs the README evidence table); battery S8 pins file-vs-README family equality
  for whoever edits the file, but only when the battery is run.
- Output is buffered until `go test` exits (liveness of streaming not claimed).
- Skipped probes still count as executed (`=== RUN` present) — by design: zero
  *matching* is the disease, per-test skip policy is out of scope.
- Exit code 3 is new and means "instrument failure" (documented in README §note
  and the docstring).

## 8. Reproduce

```
python3 agent-festival/review/2026-11-08/run_probes.py /path/to/go   # live: rc=1 today
python3 -I agent-festival/review/2026-11-08/test_run_probes.py       # battery: rc=0
```

Raw logs: transferred bundle `~/nga-handoff/g12-evidence/` on aws-claw
(R0-R3, fixed retests, mutation logs, three regression variants).
