# Document labeler (digitizer)

Go manufacturing document digitizer: PDF/PNG/JPEG → reusable schemas + value files.

## Requirements

- Go 1.22+
- [Tesseract](https://github.com/tesseract-ocr/tesseract) (`eng` data)
- [Poppler](https://poppler.freedesktop.org/) (`pdftoppm`) for PDFs
- A llama.cpp (or compatible) HTTP server with a small Qwen-class GGUF, **or** `DIGITIZER_ALLOW_STUB=1` for heuristic-only local/dev

## Build

```bash
go build -o bin/digitizer ./cmd/digitizer
```

## Scan

```bash
export DIGITIZER_LLAMA_URL=http://127.0.0.1:8080   # or DIGITIZER_ALLOW_STUB=1
./bin/digitizer scan --dept stores --root . path/to/invoice.png
```

Outputs:

- `schemas/<dept>/catalog.json` — private schemas (seeded manufacturing types)
- `values/<dept>/*.json` — field values + soft key hints
- `labelled/<dept>/` — copy of successfully labelled files

Low confidence stays in place for watch-folder inputs and writes `values/<dept>/unknown/` with `rescanNeeded`.

## Promote / weekly conflicts

```bash
./bin/digitizer accept --dept stores --type invoice --root .
./bin/digitizer conflicts --dept stores --root .
```

Catalog owners should review drafts/conflicts weekly.

## Watch folder

```bash
mkdir -p inbox/stores
./bin/digitizer watch --dept stores --root . --once
```

## Schema modes

`fully_learned` | `seeded_extensible` | `known_dynamic` — auto-selected; override with `--mode` or `DIGITIZER_MODE`.

## Not in scope

Product UI, typed document relations / quantity checks, n8n as the OCR engine.
