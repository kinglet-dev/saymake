#!/usr/bin/env bash
# Prints the CHANGELOG.md section for a version, for the release page, and
# fails if there is none or it is still marked unreleased.
# Usage: scripts/release-notes.sh 1.2.3
set -euo pipefail
version="$1"
notes=$(awk -v v="$version" '
  /^## / { on = (index($0, "## " v " ") == 1) ; if (on) { head = $0; next } }
  on { print }
' CHANGELOG.md)
head=$(grep -m1 "^## $version " CHANGELOG.md || true)
if [[ -z "$head" ]]; then
  echo "CHANGELOG.md has no section for $version: add '## $version (YYYY-MM-DD)'." >&2
  exit 1
fi
if [[ "$head" == *unreleased* ]]; then
  echo "CHANGELOG.md still marks $version as unreleased: replace '(unreleased)' with the release date." >&2
  exit 1
fi
printf '%s\n' "$notes"
