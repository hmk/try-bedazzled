#!/usr/bin/env bash
# Benchmark try-bedazzled (Go) vs try-cli (C) vs try (Ruby)
# Requires: hyperfine, fixture dirs from setup.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Paths — override with env vars
GO_BIN="${GO_BIN:-${PROJECT_DIR}/dist/try}"
C_BIN="${C_BIN:-/tmp/try-bench/try-cli/dist/try}"
RB_BIN="${RB_BIN:-/tmp/try-bench/try-ruby/bin/try}"
FIXTURES="${FIXTURES:-/tmp/try-bench-fixtures}"
RESULTS_DIR="${SCRIPT_DIR}/results"

mkdir -p "$RESULTS_DIR"

# Common hyperfine flags
HF="hyperfine --warmup 3 --min-runs 30 -i -N"

# Verify binaries exist
for bin in "$GO_BIN" "$C_BIN"; do
  if [ ! -x "$bin" ]; then
    echo "ERROR: Binary not found or not executable: $bin"
    exit 1
  fi
done

HAS_RUBY=false
if [ -f "$RB_BIN" ]; then
  HAS_RUBY=true
fi

# Verify fixtures exist
if [ ! -d "$FIXTURES" ]; then
  echo "ERROR: Fixtures not found. Run: ./bench/setup.sh"
  exit 1
fi

FIXTURE_COUNT=$(ls "$FIXTURES" | wc -l | tr -d ' ')
echo "=== try-bedazzled benchmark ==="
echo "Fixtures: $FIXTURE_COUNT directories in $FIXTURES"
echo "Go: $GO_BIN"
echo "C:  $C_BIN"
echo "Rb: $RB_BIN ($HAS_RUBY)"
echo ""

run_bench() {
  local name="$1" go_args="$2" c_args="$3" rb_args="${4:-$3}"

  echo ">>> $name"
  if $HAS_RUBY; then
    $HF \
      --export-json "$RESULTS_DIR/${name// /_}.json" \
      --export-markdown "$RESULTS_DIR/${name// /_}.md" \
      -n "Go" "$GO_BIN $go_args" \
      -n "C" "$C_BIN $c_args" \
      -n "Ruby" "ruby $RB_BIN $rb_args"
  else
    $HF \
      --export-json "$RESULTS_DIR/${name// /_}.json" \
      --export-markdown "$RESULTS_DIR/${name// /_}.md" \
      -n "Go" "$GO_BIN $go_args" \
      -n "C" "$C_BIN $c_args"
  fi
  echo ""
}

# Scenario 1: Version (pure startup overhead)
run_bench "version" "--version" "--version"

# Scenario 2: Init output
run_bench "init" "init /tmp/test-tries" "init /tmp/test-tries"

# Scenario 3: Startup + scan + cancel
run_bench "startup" \
  "exec --path $FIXTURES --and-keys ESCAPE" \
  "exec --path $FIXTURES --and-keys ESCAPE --and-exit"

# Scenario 4: Fuzzy match + select
run_bench "fuzzy" \
  "exec --path $FIXTURES --and-keys redis,ENTER" \
  "exec --path $FIXTURES --and-keys redis,ENTER --and-exit"

echo "=== Results saved to $RESULTS_DIR ==="
echo ""
echo "=== Binary sizes ==="
ls -lh "$GO_BIN" "$C_BIN" | awk '{print $5, $NF}'
$HAS_RUBY && echo "Ruby: interpreted (no binary)"
echo ""
