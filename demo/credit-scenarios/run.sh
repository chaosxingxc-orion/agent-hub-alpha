#!/usr/bin/env bash
set -euo pipefail

demo_root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/../.." && pwd)"
demo_go="${GO_BIN:-go}"
cd "$demo_root/agent-festival/go"
demo_build="$(mktemp -d "${TMPDIR:-/tmp}/agent-hub-credit-demo-bin.XXXXXX")"
trap 'rm -f -- "$demo_build/credit-demo"; rmdir -- "$demo_build"' EXIT
# The report records Git HEAD itself. Avoid automatic build stamping, which
# can misidentify Git worktrees with the supported Go 1.24 toolchain.
"$demo_go" build -buildvcs=false -o "$demo_build/credit-demo" ./cmd/credit-demo
"$demo_build/credit-demo" "$@"
