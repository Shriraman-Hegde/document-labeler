---
title: Manufacturing Document Digitizer - Plan
type: feat
date: 2026-11-26
topic: manufacturing-document-digitizer
artifact_contract: ce-unified-plan/v1
product_contract_source: ce-brainstorm
execution: code
---

# Manufacturing Document Digitizer - Plan

## Goal Capsule

- **Objective:** Manufacturing departments digitize paper scans into reusable schemas and field values without re-keying, while keeping private catalogs until they choose to promote a type company-wide.
- **Means:** A Go digitizer engine (CLI + department watch folders) with Tesseract-first OCR, required local LLM assist, and three schema modes — redesigned for this product, not a layout port of the reference labeler.
- **Product authority:** This plan owns the digitizer engine and catalog behavior for v1. Frontend/camera capture and n8n-as-engine are not active scope.
- **Open blockers:** None for planning.
- **Execution profile:** Greenfield Go module; CLI-first vertical slice then watch-folder parity.
- **Who finishes:** `ce-work` / implementing agent ships units in order; human merges the PR.

## Product Contract
<!-- Product Contract preservation: Product Contract unchanged -->

### Summary

Build a Go manufacturing document digitizer that turns PDF/PNG/JPEG scans into reusable schemas and value files for cross-department shop-floor use. Tesseract reads pages first; a required local model assists on hard pages and on typing/schema decisions. Private-first catalogs, a small shipped manufacturing seed set, soft shared-key hints after promote, and a weekly draft/conflict review define v1 success.

### Problem Frame

Today manufacturing paperwork stays on paper across stores, audit, and other departments. Nothing is digital, so every handoff is manual re-entry. Forms and headings vary by department, so a hard-coded form list will not travel. Departments will also resist forced company-wide schema adoption, and labels will be inconsistent across scans of the “same” form.

### Key Decisions

- **Private-first conflict network with scan-first day one** over pure catalog-first or scan-only — clear paper piles while supporting multi-dept refusal to promote. (session-settled: user-directed — chosen over catalog-first / scan-only: multi-dept reality plus first-week digitization) Governs R1, R8, R14.
- **All three schema modes required in v1** with automatic selection and optional override — not a single learning path. (session-settled: user-directed — chosen over shipping one mode first: not useful until all three exist) Governs R5, R6.
- **Required local LLM** (e.g. Qwen-class via llama.cpp) — not OCR-only or optional-only. (session-settled: user-directed — chosen over OCR-only / optional assist: dynamic labeling depends on the model) Governs R3, R4.
- **Hybrid read path:** Tesseract first; local AI when OCR is weak/hard — not AI-only vision. (session-settled: user-directed — chosen over Tesseract-only or AI-vision-only) Governs R2, R3.
- **Differentiation:** page headings + AI type guess, filename as hint. (session-settled: user-directed — combined headings+filename with headings+AI) Governs R4.
- **Soft shared-key links only** in v1 — no typed relations or quantity checks. (session-settled: user-directed — chosen over full relation engine) Governs R10.
- **Soft keys cross-department only after promote;** private types stay department-local. (session-settled: user-directed) Governs R10, R9.
- **Ship a small manufacturing seed set** for seeded mode. (session-settled: user-directed — chosen over no built-in seed) Governs R7.
- **Reject n8n as the digitizer;** orchestrator-only later if ever. (session-settled: user-directed after project-grounded Reject verdict) Governs Scope Boundaries.
- **Low confidence:** stay in inbox, write minimal value file, flag rescan needed, no schema learning. (session-settled: user-directed) Governs R11.
- **Inconsistent labels:** keep both fields, flag conflict, still label. (session-settled: user-directed) Governs R12.
- **Weekly draft/conflict review** is part of v1 success. (session-settled: user-directed — chosen over ignore-until-later / hard cap) Governs R14.

<!-- ce-section: work-relationships -->
### How This Work Fits Together

This plan covers the **digitizer engine and catalog behavior**. Broader understanding (not a committed roadmap):

- **Frontend / camera-scanner capture UI** — Depends on this engine’s intake and outputs; Can proceed independently of typed relations. Still to decide: capture UX.
- **Typed document relations and quantity checks** — Depends on stable schemas/values and soft keys from this plan; deferred deliberately.
- **Optional n8n orchestration** — Can proceed independently only as a caller of this engine; Outside this product’s identity as a substitute engine.

### Actors

- A1. Department operator — drops scans or runs CLI for their department.
- A2. Department catalog owner — reviews drafts/conflicts weekly; accepts promotes to company-shared.
- A3. Digitizer engine — OCR, model assist, schema modes, catalog writes, soft-key hints.
- A4. Local LLM runtime — required assist for hard pages and typing/schema judgment.

### Requirements

**Ingest and outputs**

- R1. Operators digitize PDF, PNG, and JPEG via CLI and department watch folders without a product UI.
- R2. The engine reads page text with Tesseract first.
- R3. A required local LLM assists when OCR is weak or the page is hard, and for typing/schema judgment.
- R4. Document type differentiation uses page headings/labels and AI type guess together; filename may hint but must not be sole authority.
- R13. Each successful label produces a reusable schema artifact and a value artifact for the document.

**Schema modes and catalogs**

- R5. v1 exposes three schema modes as options: fully learned; seeded + extensible; known types with dynamic fields.
- R6. Mode selection is automatic on the happy path; CLI or folder config may override.
- R7. Seeded mode ships with a small manufacturing seed set departments can extend.
- R8. New types start private to a department and become company-shared only after an explicit accept/promote.
- R9. Company-shared types may carry department field overlays without forking the type.
- R15. Within a department, private schemas are reusable immediately; company promotion requires accept.

**Links, conflicts, failure**

- R10. Soft shared-key hints record matching key text across documents; cross-department soft keys apply only for promoted/company-shared types.
- R11. When confidence is too low to classify, the file stays in the inbox, a minimal value file (raw OCR, type unknown) is written, “rescan needed” is surfaced, and no schema learning occurs.
- R12. When the same form uses divergent heading wording, keep both fields, flag a conflict, and still label the document.

**Operating success**

- R14. v1 success expects a simple weekly review of drafts and conflicts by catalog owners.

### Key Flows

- F1. Happy-path digitize
  - **Trigger:** Operator submits PDF/PNG/JPEG via CLI or department watch folder.
  - **Actors:** A1, A3, A4
  - **Steps:** Read with Tesseract; invoke local LLM if OCR weak/hard; differentiate type (headings + AI, filename hint); auto-select schema mode (or honor override); write/update private schema as needed; write value artifact; record soft keys when allowed.
  - **Outcome:** Labelled document with schema + values; no product UI required.
  - **Covered by:** R1–R7, R13, R15

- F2. Low-confidence / rescan
  - **Trigger:** Engine cannot confidently classify or learn from the scan.
  - **Actors:** A1, A3
  - **Steps:** Leave file in inbox; write minimal value file (raw OCR, unknown type); surface rescan needed; skip schema learning.
  - **Outcome:** Evidence preserved without polluting catalogs.
  - **Covered by:** R11

- F3. Conflict and weekly review
  - **Trigger:** Divergent labels on an otherwise labelable document, or accumulated drafts/conflicts.
  - **Actors:** A2, A3
  - **Steps:** Keep both fields and flag conflict while labeling; catalog owner reviews drafts/conflicts on a weekly cadence; promote to company-shared only via accept.
  - **Outcome:** Department keeps moving; company catalog grows only by deliberate promote.
  - **Covered by:** R8, R12, R14, R15

### Acceptance Examples

- AE1. Clear invoice scan
  - **Covers:** R1, R2, R4, R13
  - **Given:** A readable invoice PNG in a stores watch folder with headings present and filename containing `invoice`.
  - **When:** The engine processes the folder.
  - **Then:** It assigns an invoice-like type using headings + AI (filename only as hint), writes schema + value artifacts, and does not require a UI.

- AE2. Weak OCR page
  - **Covers:** R2, R3
  - **Given:** A hard/noisy scan where Tesseract text is weak.
  - **When:** The engine processes the file.
  - **Then:** The required local LLM assists the read and/or typing path before giving up.

- AE3. Rescan needed
  - **Covers:** R11
  - **Given:** A file with no usable labels/confidence.
  - **When:** The engine processes it.
  - **Then:** The file remains in inbox, a minimal unknown-type value file exists, rescan needed is surfaced, and no schema is learned.

- AE4. Divergent headings
  - **Covers:** R12
  - **Given:** Two scans of the same private type use “PO No.” and “Purchase Order #”.
  - **When:** Both are labeled.
  - **Then:** Both fields are kept, a conflict is flagged, and both documents are still labeled.

- AE5. Soft keys and promote boundary
  - **Covers:** R9, R10, R15
  - **Given:** Stores and audit each have private types mentioning the same PO text; neither type is promoted.
  - **When:** Both are processed.
  - **Then:** Soft keys do not cross departments. After accept/promote of a shared type, later documents may soft-link across departments on that key.

### Success Criteria

- Departments can clear paper piles into digital schema + value artifacts without re-keying extracted fields.
- All three schema modes are available in v1 with auto-select and override.
- Catalog owners perform a simple weekly draft/conflict review as part of normal operation.
- Private types remain usable department-locally without forcing promote.

### Scope Boundaries

**Deferred for later**

- Product frontend and camera/scanner capture UI
- Typed document relations and quantity match/mismatch checks
- Optional n8n (or similar) orchestration that only calls this engine

**Outside this product's identity**

- Replacing the digitizer with n8n workflows as the understanding engine
- A hard-coded-only form catalog with no learning/extension path
- Porting the reference repo’s layout as the design of this codebase

### Dependencies / Assumptions

- Operators can install/run Tesseract and a local LLM runtime (llama.cpp-class) with a small Qwen-class model for v1.
- Department identity is supplied at intake (CLI argument and/or watch-folder mapping).
- This repository starts empty; the product is greenfield (`document-labeler`), informed by but not copying https://github.com/Shriraman-Hegde/manufacturing-doc-labeler.
- “Small manufacturing seed set” contents are chosen during planning from common shop-floor forms, not frozen in this contract beyond “small and extensible.”

### Outstanding Questions

None blocking. Defaults for former planning deferrals are recorded as KTDs below.

### Sources / Research

- Reference behavior (inspiration, not a port): https://github.com/Shriraman-Hegde/manufacturing-doc-labeler
- Project grounding: empty greenfield repo `document-labeler` (no app sources/commits at brainstorm time)
- External adoption verdict on n8n-as-engine: Reject (orchestrator ≠ document understanding; native PDF extract is not OCR for scans)

## Planning Contract

### Key Technical Decisions

- KTD1. **Module layout:** `cmd/digitizer` + `internal/{ocr,llm,schema,label,ingest,engine}` — domain packages over a flat main. (session-settled product constraints govern; layout is planning) Governs U1–U9.
- KTD2. **OCR:** shell out / bind to system Tesseract (eng) for PNG/JPEG; PDF pages via `pdftoppm` then Tesseract. Governs R2, U2.
- KTD3. **LLM:** llama.cpp HTTP or CLI against a configured Qwen-class GGUF; fail closed with clear error if required model unavailable at process start. Governs R3, U5.
- KTD4. **Weak OCR gate:** mean word confidence below 0.55 or fewer than 12 tokens → treat as weak/hard and invoke LLM assist. Configurable via env. Governs R3, R11, U2, U5.
- KTD5. **Rescan gate:** after Tesseract+LLM still fewer than 4 heading-like tokens or type confidence below 0.4 → R11 path. Governs R11, U8.
- KTD6. **Artifacts:** JSON on disk — `schemas/<dept>/catalog.json` versioned entries; `values/<dept>/<stamp>-<basename>.json`; minimal unknown values under `values/<dept>/unknown/`. Governs R13, U3, U4.
- KTD7. **Seed set v1:** `invoice`, `purchase_order`, `packing_slip`, `goods_receipt`, `inspection_report` with starter heading synonyms. Governs R7, U4.
- KTD8. **Mode auto-select:** prefer known-type+dynamic when seed/catalog type matches strongly; seeded+extensible when partial seed match; fully learned when no match — override via `--mode` / `DIGITIZER_MODE` / folder `.digitizer.yaml`. Governs R5, R6, U6.
- KTD9. **Soft keys:** normalize `poNumber`, `lot`, `partNumber`, `deliveryNumber`, `drawingNumber`; emit `softLinks` on value artifacts; cross-dept index only for promoted types. Governs R10, U7.
- KTD10. **Watch folders:** `inbox/<dept>/` processed into `labelled/<dept>/` on success; failures stay in inbox per R11. Governs R1, U9.

### High-Level Technical Design

```text
[CLI | watch inbox/<dept>]
        |
        v
   ingest (PDF→images, PNG/JPEG)
        |
        v
   ocr.Tesseract  --weak?-->  llm.Assist(read/type)
        |
        v
   label.Differentiate(headings + AI, filename hint)
        |
        v
   schema.ModeAuto | override
        |
        +--> schema.Catalog (private → promote)
        +--> values.Writer (+ softLinks)
        +--> conflict flags / rescan minimal value
```

Domain model (invalid states unrepresentable):

- `DocType` branded string; `SchemaVersion` with status `private|promoted` and optional `conflicts[]`
- `SchemaMode` enum: `fully_learned | seeded_extensible | known_dynamic`
- `FieldValue` = text + confidence + optional box; never a bare string at boundaries
- External OCR/LLM JSON parsed at boundary into typed models

### Assumptions

- Host has Tesseract + Poppler (`pdftoppm`) installed for local/dev; Docker image may bundle later (out of U1–U9 critical path if binary runs with system deps).
- Model path via `DIGITIZER_MODEL` / `DIGITIZER_LLAMA_URL`; CI may skip LLM tests without the binary (`testing` build tag or short mode).
- No HTTP server in v1 units (CLI + watch only).

### Sequencing

U1 → U2 → U3 → U4 → U5 → U6 → U8 → U7 → U9 (soft keys after label path; watch last).

## Implementation Units

### U1. Go module and CLI scaffold
- **Goal:** Runnable `digitizer` binary with `--help`, department flag, version.
- **Covers:** R1 (intake surface)
- **Files:** `go.mod`, `cmd/digitizer/main.go`, `internal/cli/`, `README.md` (run deps)
- **Verify:** `go build ./...`; `go test ./internal/cli/...`
- **Done when:** Binary prints usage including department and mode override flags.

### U2. Ingest + Tesseract OCR
- **Goal:** PDF/PNG/JPEG → page images → OCR tokens with confidence.
- **Covers:** R1, R2
- **Depends on:** U1
- **Files:** `internal/ingest/`, `internal/ocr/`
- **Verify:** unit tests with fixture images; integration test skipped if no tesseract
- **Done when:** PNG and multi-page PDF return non-empty token streams on fixtures.

### U3. Domain types and value/schema writers
- **Goal:** Typed models + JSON write/read for schema catalog entries and value files.
- **Covers:** R13
- **Depends on:** U1
- **Files:** `internal/schema/types.go`, `internal/schema/store.go`, `internal/label/value.go`
- **Verify:** round-trip tests for catalog and value JSON
- **Done when:** Invalid status/mode combinations fail compile or parse.

### U4. Private catalog, seed set, promote
- **Goal:** Department catalogs with seed types; immediate private reuse; `digitizer accept --promote`.
- **Covers:** R7, R8, R9, R15
- **Depends on:** U3
- **Files:** `internal/schema/seed.go`, `internal/schema/catalog.go`, CLI accept subcommand
- **Verify:** tests for seed load, private write, promote gate
- **Done when:** Promote without accept is impossible; overlays attach without forking type id.

### U5. Required local LLM assist
- **Goal:** Process fails fast if model runtime missing; assist OCR/type on weak pages.
- **Covers:** R3
- **Depends on:** U2, U3
- **Files:** `internal/llm/`
- **Verify:** fake server tests; skip live model in `-short`
- **Done when:** Weak-OCR path calls LLM; missing model returns actionable error at start.

### U6. Differentiate + mode auto-select + override
- **Goal:** Headings + AI (+ filename hint) choose type; auto mode with override.
- **Covers:** R4, R5, R6
- **Depends on:** U4, U5
- **Files:** `internal/label/differentiate.go`, `internal/schema/mode.go`
- **Verify:** table-driven tests for mode selection and filename-not-sole-authority
- **Done when:** AE1-shaped fixture labels without treating filename as sole type.

### U7. Soft shared-key hints
- **Goal:** Emit normalized soft links; cross-dept only for promoted types.
- **Covers:** R10
- **Depends on:** U6
- **Files:** `internal/label/softkeys.go`
- **Verify:** same-dept vs cross-dept cases
- **Done when:** AE5 behavior holds in tests.

### U8. Conflicts + low-confidence rescan path
- **Goal:** Divergent labels keep both + conflict flag; rescan path per R11.
- **Covers:** R11, R12, R14 (surface conflicts for weekly review via `digitizer conflicts` list)
- **Depends on:** U6
- **Files:** `internal/label/conflict.go`, CLI `conflicts` subcommand
- **Verify:** AE3, AE4 tests
- **Done when:** Inbox file unchanged on rescan path; conflict list enumerates open flags.

### U9. Watch-folder runner
- **Goal:** Process `inbox/<dept>/` continuously or one-shot `--watch-once`.
- **Covers:** R1
- **Depends on:** U6–U8
- **Files:** `internal/ingest/watch.go`, CLI `watch`
- **Verify:** temp-dir integration test
- **Done when:** Dropping a fixture PNG yields labelled output without CLI scan of that file path.

## Verification Contract

- `go test ./...` (LLM/OCR integration skipped under `-short` or missing binaries)
- `go vet ./...`
- Fixture-driven AE1–AE5 coverage under `testdata/`
- Manual smoke (dev): build binary, point at sample scan with model+tesseract installed

## Definition of Done

- All units U1–U9 merged criteria met
- Product Contract R1–R15 addressed by unit coverage above
- README documents Tesseract, Poppler, model env vars, seed types, weekly `conflicts` review
- No n8n dependency; no HTTP product UI

