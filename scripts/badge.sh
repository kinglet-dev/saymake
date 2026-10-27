#!/usr/bin/env bash
# Writes a flat SVG badge to standard output, made here so no badge service
# sees the repository. Usage: scripts/badge.sh LABEL VALUE [COLOR]
# Text is white; the default color is green (#1A7F37, 5.1:1 with white),
# and the label side is gray (#3D444D, 9.9:1), both above WCAG AA's 4.5:1.
set -euo pipefail
label="$1" value="$2" color="${3:-#1A7F37}"
esc() { sed -e 's/&/\&amp;/g' -e 's/</\&lt;/g' -e 's/>/\&gt;/g' -e 's/"/\&quot;/g' <<<"$1"; }
l=$(esc "$label") v=$(esc "$value")
lw=$(( ${#label} * 7 + 12 )) vw=$(( ${#value} * 7 + 12 ))
w=$(( lw + vw ))
cat <<SVG
<svg xmlns="http://www.w3.org/2000/svg" width="$w" height="20" role="img" aria-label="$l: $v"><title>$l: $v</title><rect width="$lw" height="20" rx="3" fill="#3D444D"/><rect x="$lw" width="$vw" height="20" rx="3" fill="$color"/><rect x="$lw" width="4" height="20" fill="$color"/><g fill="#FFFFFF" font-family="Verdana,DejaVu Sans,sans-serif" font-size="11" text-anchor="middle"><text x="$(( lw / 2 ))" y="14">$l</text><text x="$(( lw + vw / 2 ))" y="14">$v</text></g></svg>
SVG
