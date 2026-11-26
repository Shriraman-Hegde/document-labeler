# Application documentation

**Product:** Document Labeler (digitizer)  
**Date:** 2026-11-26  
**Repos:** [document-labeler](https://github.com/Shriraman-Hegde/document-labeler) (engine + HTTP API) · [document-labeler-modules-ui](https://github.com/Shriraman-Hegde/document-labeler-modules-ui) (Lit modules UI)

## 1. Purpose

Manufacturing (and adjacent) departments digitize PDF/PNG/JPEG scans into **reusable schemas** and **value files** without re-keying. Catalogs stay **department-private** until an explicit promote. Operators capture via CLI, watch folder, or the browser UI.

## 2. System context

```text
┌─────────────────────────────┐     CORS      ┌──────────────────────────────┐
│  document-labeler-modules-ui │ ─────────────▶ │  document-labeler            │
│  Lit + Vite (dev)            │   HTTPS/HTTP  │  digitizer serve :8091       │
│  Rollup+Babel (prod bundle)  │  /v1/scan     │  engine → OCR → LLM → label  │
└─────────────────────────────┘  /v1/relate    │  schemas/ values/ labelled/  │
                                               └──────────────┬───────────────┘
                                                              │
                               ┌──────────────────────────────┼───────────────┐
                               ▼                              ▼               ▼
                        Tesseract OCR                  llama.cpp         filesystem
                        (+ preprocess)               (or STUB=1)      root catalogs
```

**Hard boundary:** the Go process is API/CLI only. The UI is never embedded.

## 3. Components and ownership

| Piece | Repo / package | Owns |
| --- | --- | --- |
| CLI (`scan`, `watch`, `accept`, `conflicts`, `serve`) | `cmd/digitizer`, `internal/cli` | Operator entrypoints |
| HTTP API + CORS | `internal/server` | Browser contract |
| Ingest (PDF→PNG, image load) | `internal/ingest` | Page images for OCR |
| OCR preprocess + Tesseract | `internal/ocr` | Text + confidence tokens |
| Local LLM assist | `internal/llm` | Weak-OCR / type assist (fail-closed unless stub) |
| Schema catalog + modes | `internal/schema` | Private/shared types, drafts, conflicts |
| Label / differentiate | `internal/label` | Type choice, value files |
| Relations + quantity checks | `internal/relate` | Cross-doc soft keys / checks |
| Capture UI modules | UI `src/modules/*` | Camera/upload, scan/relate UX |
| API client | UI `src/lib/api.ts` | Browser → API calls |

**State owners (one each):**

- **Catalog state** → filesystem under `schemas/<dept>/` (engine writes; UI never writes).
- **Value / unknown inbox artifacts** → `values/<dept>/` (+ `unknown/`).
- **UI session state** → Lit `@state()` on `digitizer-app` (no Zustand).
- **API base URL preference** → browser `localStorage` key `DIGITIZER_API`.

## 4. Runtime modes

| Mode | When | Notes |
| --- | --- | --- |
| Production LLM | `DIGITIZER_LLAMA_URL` set | Fail closed if `/health` or `/v1/models` unreachable |
| Stub | `DIGITIZER_ALLOW_STUB=1` | Dev/CI only — not for prod accuracy claims |
| Schema modes | auto or `--mode` / form override | `fully_learned` \| `seeded_extensible` \| `known_dynamic` |

## 5. HTTP contract (UI ↔ API)

| Method | Path | Body / query | Success |
| --- | --- | --- | --- |
| `GET` | `/health` | — | `{ "status": "ok" }` |
| `GET` | `/` | — | service discovery JSON |
| `POST` | `/v1/scan` | multipart: `dept`, optional `mode`, `file` | documents + relations + checks |
| `GET`/`POST` | `/v1/relate` | `dept` | relations + checks |

CORS is enabled for separately hosted UI origins. Max upload parse size: 64 MiB.

## 6. Data on disk (API root)

| Path | Meaning |
| --- | --- |
| `schemas/<dept>/catalog.json` | Private department schemas |
| `values/<dept>/*.json` | Labelled field values |
| `values/<dept>/unknown/` | Weak / unmatched scans (`rescanNeeded`) |
| `labelled/<dept>/` | Copies of successfully labelled sources |
| `inbox/<dept>/` | Watch-folder intake |

## 7. Security & privacy (prod assumptions)

- Run API on a private network or behind auth at the edge (app has no built-in auth yet).
- Department catalogs and scans may contain PII / commercial data — treat root volume as sensitive.
- Do not enable `DIGITIZER_ALLOW_STUB` in production.
- UI talks only to the configured API origin; no third-party analytics in the modules UI.

## 8. Deploy topology (recommended)

1. **API host:** `digitizer serve --addr :8091 --root /data/digitizer` with LLM sidecar (`llama.cpp`) or org-approved model gateway.
2. **UI host:** static files from `npm run build` (`dist/`) on any HTTPS static host / object storage + CDN.
3. **Config:** UI `?api=` or saved API base pointing at the public/internal API origin over HTTPS.

See [SDLC.md](./SDLC.md) for release gates and [REPORTS.md](./REPORTS.md) for required evidence.
