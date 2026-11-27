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
```

(The `rc=0` marker was echoed by the harness terminal, not part of the archived
`g12-R1.log` body — review round 6 flagged the quote-block contamination; the
three lines above are exactly what the log holds. Exit code 0 independently
re-measured by rounds 4-6.) Zero tests ran; the old runner propagated rc=0 — a
"probe suite passed" signal with nothing probed.

### R2 — overlay-ineffective shape (CONFIRMED false green, family-granular)

Copy with the overlay `Replace` map emptied (simulates key drift / lost
injection). Measured: 13 `=== RUN` lines, all P0A/XS2, **0** `TestNGAReview_*`,
final `PASS`, `rc=0`. The known-red acceptance probes vanished and the suite
reported green.

Consequence for fix design: a total-count-only zero check does NOT close this
form (total was 13). Detection must be probe-family-granular.

### R3 — repo-root symlink invocation (E-report mechanism REFUTED as scoped)

`ln -sfn ~/nga-handoff/agent-hub-wt /tmp/nga-sym-repo`, then run the real
script through that symlink path. Measured: `rc=1`, 7 probes ran — RUN/FAIL
structure and rc identical to R0 (timing strings differ). The E-report's claim
(macOS `/private/tmp` vs `/tmp` aliasing between the overlay key and go's
package path) is refuted for this script: the tempdir never becomes the key,
and `cwd=go_root` and the overlay key derive from the same `__file__`-resolved
root — so that specific mechanism cannot occur (the 169 adversarial note
confirmed; re-verified independently here).

### R3b — symlinked `agent-festival/go`: a real key-divergence channel (found by review-loop round 3)

The original record sentence said the divergence "does not exist" — true for the
E-report mechanism, but the way it was written invited a broader reading that
exceeded R3's evidence radius (campaign rule: closing-word strength = evidence
radius; own violation, own correction). R3 symlinks the repo ROOT, which
`Path(__file__).resolve()` collapses before `go_root` is formed. If instead
`agent-festival/go` itself is a symlink, the appended `go` component is NOT
resolved: the overlay key stays on the lexical path while the child cwd is
physical, and go silently drops the injection. Measured on this machine
(logs `g12-evidence-symgo-*.log`):

| Generation | Runner | Result under a symlinked `agent-festival/go` |
| --- | --- | --- |
| A | main, pre-fix | `rc=0`, 13 RUN lines, **0** probes — a real false green; the original etiology note was incomplete |
| B | commit-2 (family guard, no normalization) | `rc=3` — red, but misattributed ("never ran" blames the probes; injection died) |
| C | commit-3 (`go_root` resolved) | `rc=1`, 20 RUN, 7 probes — injection works, honest red preserved |

Per 169 §五 G12.2 ("key normalization per adjudication outcome"), the
adjudication outcome now identifies this channel, so normalization was applied
(commit 3). The round-3 verifier REFUTED the finding's MAJOR grade (no live
false green on the tree; the record sentence holds under its literal E-scoped
reading) — accepted — but the substance (an unadjudicated real channel + an
overreaching absolute) is corrected here under the claim-fidelity duty.

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
- (commit 3) `go_root` resolved per R3b: overlay key and child cwd pinned to the
  same physical root under a symlinked `agent-festival/go` (battery S11).

Post-fix retests (real go): R1-shape `rc=3` with `zero tests matched`;
R2-shape `rc=3` listing all 7 probe names; live baseline `rc=1` unchanged,
7 probes still running and still honestly red (also under the go-symlink
topology, generation C).

## 4. Mutation table (M1-M8 re-run on committed state 28856ea; M9-M11 on
879d44c; detached worktrees; raw logs shipped `g12-evidence-mut-M*.log`)

| ID | Mutation | Battery cases turned red | Result |
| --- | --- | --- | --- |
| M1 | drop `expected - seen` guard | S1, S6, S11b, S12 | killed |
| M2 | drop `total == 0` guard | S2, S7 (message assertions) | killed |
| M3 | drop empty-`expected` guard | S10 | killed |
| M4 | invert top-level `/` filter | S1, S3, S4, S6, S7, S10, S11, S11b, S12 | killed |
| M5 | instrument exit 3 -> 0 | S1, S2, S6, S7, S10, S11b, S12 | killed |
| M6 | instrument branch masks rc!=0 to 3 | S5 | killed |
| M6b | final pass-through exit -> 3 | S3, S4, S11 | killed |
| M7 | drop `.resolve()` on go_root | S11 | killed |
| M8 | narrow PROBE_DEF to a one-probe prefix | S12 | killed |
| M9 | replay only matched RUN lines (swallow transcript) | S3, S13 | killed |
| M10 | child stderr -> DEVNULL (evidence swallow) | S13 | killed |
| M11 | delete one README evidence-table row | S8 | killed |

M9-M11 pin the round-6 closings (byte-exact replay, stderr merge, README as the
runtime anchor) and were re-run on committed state 879d44c; the M1-M8 rows hold
case-for-case there too (logs shipped).

M8 mechanically confirms the round-3 reviewer's survivor claim: against the
pre-S12 battery the PROBE_DEF-narrowing mutant passed every functional case;
S12 now kills it. (Round-1/2 mutant families — stream-swallow closed by the S3
replay assertion; R3's reviewer bonus mutant, symlinked module dir, closed by
S11/M7.)

Full revert (checkout of the pre-fix main version of the script) revives both
real disease shapes: R1-shape `rc=0`, R2-shape `rc=0` (logs `g12-R1.log`,
`g12-R2.log`), plus the R3b generation-A revival `rc=0` under a symlinked
`agent-festival/go` — the fix is load-bearing, not decoration. Every guard in
the fix has ≥1 killing battery case (`test_run_probes.py`, hermetic fake-`go`
stubs, precedent `xs3/test_xs3_validate.py` F7 battery).

## 5. Regression

- `go test -count=1 ./...` → `FULL_RC=0`, 21/21 packages ok (none dropped);
  re-run after commit 3: `FULL_RC=0` again, 21 ok, 0 FAIL lines.
- `go test -count=1 -race ./...` → `RACE_RC=0` (re-run 0).
- `go test -count=1 -shuffle=on ./...` → `SHUF_RC=0` (re-run 0; three distinct
  log md5s — not copies).
- `python3 -I agent-festival/review/2026-11-08/test_run_probes.py` → `rc=0`,
  14/14 PASS (test surface 0 → 14 executable pins for this instrument; earlier
  rounds measured 10 then 13 — monotone accumulation, nothing dropped).
- Post-polish self-attested live baseline (`g12-R0post3.log`: UTC stamp, HEAD,
  go version inline): `rc=1`, 20 RUN lines, honest red preserved. The earlier
  `g12-R0fix2.log` is content-identical to `g12-R0fix.log` by construction
  (normalization is a no-op on this physical path); the stamped re-run exists
  precisely because review round 6 correctly noted the pair was
  indistinguishable from a copy.

## 6. Deviations from the estimated "one file + one record" shape

- `test_run_probes.py` added: the campaign conservation law (every fix ships a
  permanent probe that turns red when the fix is removed) overrides the size
  estimate; precedent `xs3/test_xs3_validate.py` added a battery the same way.
- README one-sentence update: the behavior change made the old "returns the
  actual test exit code" sentence incomplete for the rc=3 case; outbound doc
  truthfulness kept in the same commit.
- Third code commit (28856ea, go_root normalization) originated inside the
  review-loop itself: round-3 filed it MAJOR, the falsification verifier
  REFUTED the grade (record held under its literal E-scoped reading; no live
  false green), and the author corrected the etiology + applied the spec-listed
  key normalization anyway because the R3b channel is physically real. Provenance
  kept in full in §2 R3b — findings must not launder their route to a change.

## 7. Documented boundaries (not closed by this instrument, honest scope)

- Deleting probe functions outright, or renaming one away from the
  `TestNGAReview_` prefix, shrinks `expected` in lockstep and stays silent —
  that drift belongs to the review surface (diff of `system_contract_probes.go.txt`
  vs the README evidence table); battery S8 pins file-vs-README family equality
  for whoever edits the file, but only when the battery is run.
- Output is buffered until `go test` exits (liveness of streaming not claimed).
- Skipped probes still count as executed (`=== RUN` present) — by design: zero
  *matching* is the disease, per-test skip policy is out of scope. Round-2
  confirmed no currently reachable skip in the run set (only `t1_llm_test.go`
  skips; it is outside the `-run` pattern); silencing a failing probe via
  `t.Skip` would require editing the boundary-forbidden probe bodies — accepted
  residual, flagged by review round 3 as minor.
- Exit code 3 is new and means "instrument failure" — all three triggers
  (zero tests matched / injected probes never ran / injected file defines no
  `TestNGAReview_` functions) are documented in README §note and the docstring.
- A real `nga_external_review_test.go` file committed at the overlay's target
  path would shadow the injection by content (same-path on-disk variant,
  round-6 note): verified nonexistent today (overlay wins over absent files by
  construction and the battery's expected set comes from the injected file),
  but a future commit there would need content-level review, not this
  instrument.

## 8. Reproduce

```
python3 agent-festival/review/2026-11-08/run_probes.py /path/to/go   # live: rc=1 today
python3 -I agent-festival/review/2026-11-08/test_run_probes.py       # battery: rc=0
```

Raw logs: bundle `~/nga-handoff/g12-evidence/` on aws-claw — R0-R3b
(three-generation symlink experiment included), self-attested post-polish
baseline `g12-R0post3.log`, mutation logs M1-M11 (complete set), battery logs
(10→13→14 growth), three regression variants x2 rounds (six distinct md5s).
