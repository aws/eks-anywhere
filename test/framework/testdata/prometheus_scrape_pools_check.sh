#!/usr/bin/env bash
# Usage: prometheus_scrape_pools_check.sh <prometheus URL> <job>...
# Waits for the Prometheus server to be ready and for every job to have an active scrape pool.
set -uo pipefail

url=$1
shift
deadline=$((SECONDS + ${CHECK_TIMEOUT_SECONDS:-600}))
targets=""
while :; do
  missing=("$@")
  if curl -sSf --max-time 10 "$url/-/ready" >/dev/null; then
    targets=$(curl -sSf --max-time 30 "$url/api/v1/targets?state=active") || targets=""
    missing=()
    for job in "$@"; do
      grep -q "\"scrapePool\":\"${job}\"" <<<"$targets" || missing+=("$job")
    done
    if [[ ${#missing[@]} -eq 0 ]]; then
      echo "Prometheus is ready and has active scrape pools: $*"
      exit 0
    fi
  fi
  if ((SECONDS >= deadline)); then
    echo "FAIL: missing active scrape pools: ${missing[*]}"
    echo "Active targets: ${targets:0:4000}"
    curl -sS --max-time 10 "$url/api/v1/status/config" | head -c 4000
    exit 1
  fi
  sleep 10
done
