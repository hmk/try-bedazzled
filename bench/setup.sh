#!/usr/bin/env bash
# Creates fixture directories for benchmarking.
# Usage: ./bench/setup.sh [count]
set -euo pipefail

COUNT=${1:-100}
BASE="/tmp/try-bench-fixtures"

rm -rf "$BASE"
mkdir -p "$BASE"

# Generate realistic directory names with varied dates and mtimes
words=("redis" "postgres" "api" "go-tui" "react-app" "ml-model" "cli-tool"
       "auth-service" "data-pipeline" "webhook" "proxy" "scheduler" "cache"
       "monitor" "gateway" "worker" "indexer" "parser" "compiler" "linter"
       "formatter" "bundler" "router" "middleware" "handler" "logger"
       "profiler" "debugger" "tracer" "metrics" "alerts" "dashboard"
       "scraper" "crawler" "ingester" "exporter" "importer" "migrator"
       "validator" "sanitizer" "encoder" "decoder" "compressor" "archiver"
       "notifier" "mailer" "renderer" "templater" "generator" "builder")

for i in $(seq 1 "$COUNT"); do
  # Random date in last 90 days
  days_ago=$((RANDOM % 90))
  date_str=$(date -v-${days_ago}d +%Y-%m-%d 2>/dev/null || date -d "-${days_ago} days" +%Y-%m-%d)

  # Pick a word, add optional suffix for uniqueness
  word=${words[$((RANDOM % ${#words[@]}))]}
  if [ "$i" -gt "${#words[@]}" ]; then
    word="${word}-v${i}"
  fi

  dir_name="${date_str}-${word}"
  mkdir -p "${BASE}/${dir_name}"

  # Set varied mtime
  hours_ago=$((days_ago * 24 + RANDOM % 24))
  touch -t "$(date -v-${hours_ago}H +%Y%m%d%H%M 2>/dev/null || date -d "-${hours_ago} hours" +%Y%m%d%H%M)" "${BASE}/${dir_name}" 2>/dev/null || true
done

echo "Created $COUNT fixture directories in $BASE"
ls "$BASE" | wc -l | xargs echo "Verified:"
