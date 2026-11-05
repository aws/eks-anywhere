#!/usr/bin/env bash
# Usage: endpoint_content_check.sh <URL> <expected text>
# Retries until the response body of URL contains the expected text.
set -uo pipefail

url=$1
expected=$2
deadline=$((SECONDS + ${CHECK_TIMEOUT_SECONDS:-600}))
while :; do
  out=$(curl -sS --max-time 10 -w '\nHTTP %{http_code}' "$url" 2>&1)
  if grep -qF -- "$expected" <<<"$out"; then
    echo "$url returned the expected content: $expected"
    exit 0
  fi
  if ((SECONDS >= deadline)); then
    echo "FAIL: $url did not return '$expected'. Last response:"
    echo "${out:0:4000}"
    exit 1
  fi
  sleep 10
done
