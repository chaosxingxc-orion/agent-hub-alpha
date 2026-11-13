#!/usr/bin/env bash
# Java conformance consumer runner: builds and runs the fixture corpus.
# Usage: ./run.sh [path-to-schema]  (default: the repo's schema/ two levels up)
set -euo pipefail
cd "$(dirname "$0")"

# JAVA_HOME: respect the user's environment (issue #9); when JAVA_HOME is set use
# $JAVA_HOME/bin/java, otherwise fall through to the PATH-resolved java.
MVN=${MVN:-mvn}

echo "== build =="
$MVN -q -DskipTests package
$MVN -q dependency:build-classpath -Dmdep.outputFile=target/cp.txt

echo "== run =="
SCHEMA_DIR="${1:-$(cd ../../schema && pwd)}"
JAVA="${JAVA_HOME:+$JAVA_HOME/bin/}java"
exec "$JAVA" -cp "target/classes:$(cat target/cp.txt)" spike.Main "$SCHEMA_DIR"
