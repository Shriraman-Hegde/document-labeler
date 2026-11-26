# Production documentation (system of record)

These docs cover the **Document Labeler** system: Go digitizer API (`document-labeler`) + Lit capture UI (`document-labeler-modules-ui`).

| Doc | Owner | Purpose |
| --- | --- | --- |
| [APPLICATION.md](./APPLICATION.md) | API repo | What the system is, boundaries, runtime topology |
| [SDLC.md](./SDLC.md) | API repo | Required lifecycle gates for both repos |
| [REPORTS.md](./REPORTS.md) | API repo | Required reports and where they live |
| [templates/](./templates/) | API repo | Report skeletons to fill per release / week |

**UI-only build/test addendum:** [document-labeler-modules-ui/docs/SDLC.md](https://github.com/Shriraman-Hegde/document-labeler-modules-ui/blob/main/docs/SDLC.md)

## Doc ownership (architect decision)

Two shapes considered:

1. **Mirror docs in both repos** — each ships APPLICATION + SDLC + REPORTS.  
   Cost: two sources of truth for one release; SDLC drift on every prod change.
2. **Hub here + thin UI addendum** — this folder owns system architecture, cross-repo SDLC, and report definitions; the UI repo only documents UI build/test/release mechanics and links back.  
   **Chosen:** one owner for system state (catalogs, API contract, reports); UI callers need only the addendum.

Do not duplicate APPLICATION or REPORTS in the UI repo.
