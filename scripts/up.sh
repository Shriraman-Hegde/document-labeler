#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$ROOT"
./scripts/fetch-model.sh
docker compose up -d llama
echo "llama.cpp listening on http://127.0.0.1:8080"
echo "export DIGITIZER_LLAMA_URL=http://127.0.0.1:8080"
echo "Then: go run ./cmd/digitizer scan --dept stores --root . testdata/fixtures/invoice.png"
echo "Or: docker compose --profile tools run --rm digitizer scan --dept stores --root /data testdata/fixtures/invoice.png"
