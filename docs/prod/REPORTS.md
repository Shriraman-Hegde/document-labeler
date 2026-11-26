# Required reports

Reports are the **evidence** that SDLC gates ran and production stays operable. Store filled copies outside git if they contain PII (recommended: shared drive / ticket). Keep **templates** in this repo.

## Report catalog

| ID | Report | Cadence | Required for prod? | Template |
| --- | --- | --- | --- | --- |
| R1 | Release report | Each production deploy | **Yes** | [templates/release-report.md](./templates/release-report.md) |
| R2 | Staging sign-off | Each production deploy | **Yes** | [templates/staging-signoff.md](./templates/staging-signoff.md) |
| R3 | Weekly catalog / conflicts | Weekly | **Yes** (ops) | [templates/weekly-catalog-report.md](./templates/weekly-catalog-report.md) |
| R4 | Weekly scan quality | Weekly | **Yes** (ops) | [templates/weekly-quality-report.md](./templates/weekly-quality-report.md) |
| R5 | Incident report | Per Sev-1/2 | When incidents occur | [templates/incident-report.md](./templates/incident-report.md) |
| R6 | CI attestation | Each release (attach links) | **Yes** | Use GitHub Actions run URLs in R1 |

## What each report must prove

### R1 — Release report

- API tag + UI tag (or commit SHAs)
- CI run URLs (both repos) — green
- Staging sign-off reference (R2)
- Deploy time, environment, operator
- Smoke results: `/health`, one `/v1/scan`, one `/v1/relate` or UI equivalent
- Rollback tags recorded

### R2 — Staging sign-off

- LLM mode = production path (not stub) unless explicitly waived
- Camera or file upload path exercised
- At least one manufacturing fixture labelled with `rescan=false` when OCR/LLM healthy
- CORS / UI→API origin checked on staging hosts

### R3 — Weekly catalog / conflicts

- Output of `digitizer conflicts --dept <name>` (per active dept)
- Promotes accepted (`digitizer accept …`) or deferred with owner
- Count of private draft types still open

### R4 — Weekly scan quality

- Count of `values/<dept>/unknown/` additions this week
- Sample of noisy OCR / rescanNeeded (no PII in git)
- Note if preprocess/OCR env knobs changed

### R5 — Incident report

- Timeline, impact (depts / scans blocked), root cause, fix, follow-ups
- Link to PR/tag that remediates

### R6 — CI attestation

- Not a separate file: paste workflow run links into R1  
  - API: `document-labeler` → Actions → `ci`  
  - UI: `document-labeler-modules-ui` → Actions → `ci`

## Retention

| Report | Minimum retention |
| --- | --- |
| R1, R2 | 12 months |
| R3, R4 | 6 months |
| R5 | 24 months |

## Metrics (where they live)

| Metric | Source | Used in |
| --- | --- | --- |
| CI pass/fail | GitHub Actions | R1 / R6 |
| Label success vs rescan | `values/` + scan JSON `rescan` | R4 |
| Conflict backlog | `digitizer conflicts` | R3 |
| API liveness | `GET /health` | R1, R2, ops |
