# Java consumer spike — the "run it once" entry point

A minimal Maven project (Java 21 + Jackson 2.17.2 + networknt 2.0.0) that consumes the
growth-family fixture corpus (ledger / adoption / judgment / release) through
the conformance interface defined in [docs/conformance-fixture-format.md](../../docs/conformance-fixture-format.md):

- `load(path)` — shared loader semantics (UseNumber-equivalent via `BigDecimal` parsing, `_expect*` pin popping, loud wrong-typed-pin)
- `check(doc)` — json-schema validation through networknt
- `suite(dirs)` — two-phase discipline: must-accept / must-reject with wrong-reason pins

## Run it

```sh
cd examples/java-consumer
./run.sh          # incremental rebuild + full corpus, rc=0 on all-green
```

## Status (2026-11-13)

The control event family and attribution key-order handling now match the current
Go/Python fixture contracts. Exact suite counts are printed by `spike.Main`;
the repository-wide corpus manifest is in the conformance format document above.
The old 81/81 statement described the 2026-11-01 corpus, not the current one.

Independent review compiled the source with Temurin JDK 21.0.11 on Windows and
ran both numeric arms, wrong-reason checks, loader checks and the mixed key-order
probe: `SPIKE-GREEN`. This validates the consumer, not a deployed Java Runtime.
Use a full JDK 21 and Maven for `run.sh`; it respects an explicitly configured
`JAVA_HOME`, otherwise uses Java from `PATH`. The reviewer could not complete the
Linux Maven build with the available JRE-only environment; this does not certify
the author's reported Linux build. Zero Go dependencies are required.

The existing networknt trailing-newline pattern discrepancy remains a reported
experiment; `SPIKE-GREEN` does not mean universal cross-language equivalence.

**Honest boundary** (per the collision report): this spike was written by the
fixture authors against their own format document — it proves the corpus is
 consumable by the Java ecosystem in a day, not anything about the quality of
the NGA-side Java implementation. The known contract gaps it surfaced (networknt
2.x/3.x API break, pattern-match mode divergence, float-syntax integer domain,
the former stale counts in the format doc) are registered and driving 0.2.0.
