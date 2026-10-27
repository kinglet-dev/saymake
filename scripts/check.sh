#!/usr/bin/env bash
# Runs every check, in order: format, vet, Go specs with coverage, the
# WebAssembly end-to-end run, the page's unit and browser tests.
# Usage (from the repo root): scripts/check.sh
set -euo pipefail
test -z "$(gofmt -l .)" || { echo "gofmt would change:"; gofmt -l .; exit 1; }
go vet ./...
GOOS=js GOARCH=wasm go vet ./cmd/...
go tool ginkgo -r --cover --coverprofile=coverage.out --skip-package=e2e
scripts/check-coverage.sh coverage.out
go tool ginkgo ./e2e
(cd web && npm ci --no-audit --no-fund && node build.mjs && npm run -s test:unit && npx playwright test)
