#!/usr/bin/env bash
# Fails unless every statement is covered, and unless every package with
# code (functions) has tests: a package without tests is left out of the
# coverage profile, so its code would look covered. The composition root
# (cmd/) is covered by the end-to-end tests instead (e2e/wasm_test.go).
# (Adapted from laserlint's script.)
# Usage: scripts/check-coverage.sh coverage.out
set -euo pipefail
profile="${1:-coverage.out}"
untested=""
while read -r pkg dir files; do
  [[ -z "$pkg" || "$pkg" == */cmd/* ]] && continue
  if (cd "$dir" && grep -qE '^func ' $files); then untested+="$pkg"$'\n'; fi
done < <(go list -f '{{if and .GoFiles (not .TestGoFiles) (not .XTestGoFiles)}}{{.ImportPath}} {{.Dir}} {{join .GoFiles " "}}{{end}}' ./...)
if [[ -n "$untested" ]]; then
  echo "These packages have code but no tests, so coverage can't see them:"
  echo "$untested"
  exit 1
fi
total=$(go tool cover -func="$profile" | awk '/^total:/ {print $NF}')
echo "coverage: $total of statements"
if [[ "$total" != "100.0%" ]]; then
  echo "Coverage must be 100%. Uncovered functions:"
  go tool cover -func="$profile" | awk '$NF != "100.0%" && !/^total:/'
  exit 1
fi
