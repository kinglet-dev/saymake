#!/usr/bin/env bash
# Writes the release's SBOM (SPDX 2.3, JSON) to standard output. Tools that
# scan binaries can't read a WebAssembly file's Go build information (nor can
# `go version -m`), so the list is written here, from what actually ships:
# Saymake itself, the Go standard library and runtime compiled into the
# .wasm, Go's wasm_exec.js, and the Outfit font inlined in the single file.
# Saymake uses no other Go modules at run time (go.mod's are for tests), and
# the page has no third-party JavaScript.
# Usage: scripts/sbom.sh VERSION
set -euo pipefail
version="$1"
go_version="$(go env GOVERSION)"
# Dated by the release's date in CHANGELOG.md, not the commit's, so the same
# source gives the same file on any commit.
day="$(grep -m1 -oE "^## ${version//./\\.} \\([0-9]{4}-[0-9]{2}-[0-9]{2}\\)" "$(dirname "$0")/../CHANGELOG.md" | grep -oE '[0-9]{4}-[0-9]{2}-[0-9]{2}' || true)"
created="${day:-1970-01-01}T00:00:00Z"
cat <<JSON
{
  "spdxVersion": "SPDX-2.3",
  "dataLicense": "CC0-1.0",
  "SPDXID": "SPDXRef-DOCUMENT",
  "name": "saymake ${version}",
  "documentNamespace": "https://github.com/kinglet-dev/saymake/releases/tag/v${version}",
  "creationInfo": { "created": "${created}", "creators": ["Organization: Kinglet (https://kinglet.dev)", "Tool: scripts/sbom.sh"] },
  "packages": [
    { "SPDXID": "SPDXRef-saymake", "name": "saymake", "versionInfo": "${version}",
      "downloadLocation": "https://github.com/kinglet-dev/saymake", "licenseConcluded": "MIT", "licenseDeclared": "MIT",
      "copyrightText": "Copyright (c) 2026 kinglet-dev", "filesAnalyzed": false },
    { "SPDXID": "SPDXRef-go-stdlib", "name": "go standard library and runtime", "versionInfo": "${go_version}",
      "downloadLocation": "https://go.dev/dl/", "licenseConcluded": "BSD-3-Clause", "licenseDeclared": "BSD-3-Clause",
      "copyrightText": "Copyright 2009 The Go Authors", "filesAnalyzed": false,
      "comment": "Compiled into saymake_${version}.wasm and the single-file page; wasm_exec.js ships as saymake_${version}_wasm_exec.js." },
    { "SPDXID": "SPDXRef-outfit", "name": "Outfit font", "versionInfo": "variable",
      "downloadLocation": "https://github.com/Outfitio/Outfit-Fonts", "licenseConcluded": "OFL-1.1", "licenseDeclared": "OFL-1.1",
      "copyrightText": "Copyright 2021 The Outfit Project Authors", "filesAnalyzed": false,
      "comment": "Inlined in saymake_${version}.html only." }
  ],
  "relationships": [
    { "spdxElementId": "SPDXRef-DOCUMENT", "relationshipType": "DESCRIBES", "relatedSpdxElement": "SPDXRef-saymake" },
    { "spdxElementId": "SPDXRef-saymake", "relationshipType": "DEPENDS_ON", "relatedSpdxElement": "SPDXRef-go-stdlib" },
    { "spdxElementId": "SPDXRef-saymake", "relationshipType": "CONTAINS", "relatedSpdxElement": "SPDXRef-outfit" }
  ]
}
JSON
