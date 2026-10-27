#!/usr/bin/env bash
# Applies one mutation (exact text replacement, first match), runs the
# package's tests with a time limit, and restores the file however the run
# ends (exit trap). Prints CAUGHT when a test fails, MISSED when all pass,
# INVALID when the mutant doesn't compile (that proves nothing).
# Usage: scripts/mutate.sh <file> <old> <new> <package> [label]
set -uo pipefail
file="$1"; old="$2"; new="$3"; pkg="$4"; label="${5:-$old -> $new}"
if ! timeout 120 go test -count=1 "$pkg" >/dev/null 2>&1; then
  echo "BASELINE FAILS for $pkg: fix the suite first; a mutation run on a failing suite proves nothing"
  exit 5
fi
backup=$(mktemp)
cp "$file" "$backup"
trap 'cp "$backup" "$file"; rm -f "$backup"' EXIT
python3 - "$file" "$old" "$new" <<'PY' || { echo "NOT FOUND  $label"; exit 2; }
import sys
p, old, new = sys.argv[1:]
s = open(p).read()
if old not in s:
    sys.exit(1)
open(p, "w").write(s.replace(old, new, 1))
PY
if ! timeout 120 go test -count=1 -run '^$' "$pkg" >/dev/null 2>&1; then
  echo "INVALID    $label (does not compile; not a real catch)"; result=4
elif timeout 120 go test -count=1 "$pkg" >/dev/null 2>&1; then
  echo "MISSED     $label"; result=1
else
  echo "CAUGHT     $label"; result=0
fi
cp "$backup" "$file"
cmp -s "$backup" "$file" || { echo "RESTORE FAILED for $file"; exit 3; }
exit $result
