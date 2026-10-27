#!/usr/bin/env bash
# Builds a release's files into dist/, with their SHA-256 checksums:
#   saymake_VERSION.html                     the page as one file, to use offline
#   saymake_VERSION.wasm                     the tool, built for WebAssembly
#   saymake_VERSION_wasm_exec.js             Go's runner for it (and its
#   saymake_VERSION_wasm_exec_LICENSE.txt    BSD-3-Clause license)
#   saymake_VERSION_web.js, _web.css, _web.html   the kit another page embeds
#   saymake_VERSION_sbom.spdx.json           what ships, as an SPDX SBOM
#   saymake_VERSION_checksums.txt
# Everything is built reproducibly (no paths or build IDs, pinned tools), so
# the same commit gives the same files. Run by .github/workflows/release.yml.
# Usage: scripts/release.sh VERSION
set -euo pipefail
version="${1:-}"
if [[ ! "$version" =~ ^[0-9]+\.[0-9]+\.[0-9]+(-[0-9A-Za-z.-]+)?$ ]]; then
  echo "usage: scripts/release.sh VERSION (like 1.2.3)" >&2; exit 2
fi
root="$(cd "$(dirname "$0")/.." && pwd)"
(cd "$root/web" && npm ci --ignore-scripts --no-audit --no-fund >/dev/null && SAYMAKE_VERSION="$version" node build.mjs)
dist="$root/dist"
rm -rf "$dist" && mkdir -p "$dist"
p="saymake_${version}"
kit="$root/web/build/kit"
cp "$root/web/build/saymake.html" "$dist/$p.html"
cp "$kit/saymake.wasm" "$dist/$p.wasm"
cp "$kit/saymake_wasm_exec.js" "$dist/${p}_wasm_exec.js"
cp "$kit/saymake_wasm_exec_LICENSE.txt" "$dist/${p}_wasm_exec_LICENSE.txt"
for f in js css html; do cp "$kit/web.$f" "$dist/${p}_web.$f"; done
"$root/scripts/sbom.sh" "$version" > "$dist/${p}_sbom.spdx.json"
(cd "$dist" && sha256sum -- * > "${p}_checksums.txt")
echo "release: $(($(ls "$dist" | wc -l) - 1)) files and their checksums in dist/"
