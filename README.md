# Document labeler (digitizer)

Go manufacturing document digitizer: PDF/PNG/JPEG → reusable schemas + value files.

## Requirements

- Go 1.22+
- [Tesseract](https://github.com/tesseract-ocr/tesseract) (`eng` data)
- [Poppler](https://poppler.freedesktop.org/) (`pdftoppm`) for PDFs
- A [llama.cpp](https://github.com/ggerganov/llama.cpp) OpenAI-compatible server with a small Qwen-class GGUF  
  **or** `DIGITIZER_ALLOW_STUB=1` for heuristic-only local/dev

## Build

```bash
go build -o bin/digitizer ./cmd/digitizer
```

## Run with llama.cpp (recommended)

```bash
./scripts/up.sh          # fetches Qwen2.5-0.5B GGUF + starts llama.cpp on :8080
export DIGITIZER_LLAMA_URL=http://127.0.0.1:8080
export DIGITIZER_MODEL=qwen
./bin/digitizer scan --dept stores --root . testdata/fixtures/invoice.png
./scripts/down.sh
```

Copy `.env.example` to `.env` if you prefer dotenv-style exports.

The digitizer **fails closed** if the LLM server is unreachable (unless stub mode is on). It probes `/health` then `/v1/models`.

## Stub mode (tests / no GPU)

```bash
export DIGITIZER_ALLOW_STUB=1
./bin/digitizer scan --dept stores --root . testdata/fixtures/invoice.png
```

## Docker digitizer binary

```bash
docker compose build digitizer
docker compose --profile tools run --rm digitizer scan --dept stores --root /data testdata/fixtures/packing_slip.png
```

## Outputs

- `schemas/<dept>/catalog.json` — private schemas (seeded manufacturing types)
- `values/<dept>/*.json` — field values + soft key hints
- `labelled/<dept>/` — copy of successfully labelled files

Low confidence stays in inbox for watch mode and writes `values/<dept>/unknown/` with `rescanNeeded`.

## Promote / weekly conflicts

```bash
./bin/digitizer accept --dept stores --type invoice --root .
./bin/digitizer conflicts --dept stores --root .
```

## Watch folder

```bash
mkdir -p inbox/stores
./bin/digitizer watch --dept stores --root . --once
```

## Fixtures

| File | Purpose |
| --- | --- |
| `testdata/fixtures/invoice.png` | clear invoice |
| `testdata/fixtures/packing_slip.png` | packing slip |
| `testdata/fixtures/goods_receipt.png` | goods receipt |
| `testdata/fixtures/weak_scan.png` | blurry / weak OCR path |

## Schema modes

`fully_learned` | `seeded_extensible` | `known_dynamic` — auto-selected; override with `--mode` or `DIGITIZER_MODE`.

## Not in scope

Product UI, typed document relations / quantity checks, n8n as the OCR engine.
