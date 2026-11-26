# SDLC — Document Labeler (both repos)

**Applies to:** `document-labeler` + `document-labeler-modules-ui`  
**Goal:** Repeatable path from change → verified release → production, with evidence listed in [REPORTS.md](./REPORTS.md).

## 1. Branching and change control

| Rule | Detail |
| --- | --- |
| Default branch | `main` (both repos) |
| Work | Short-lived `feat/*`, `fix/*`, `docs/*` branches |
| Integration | PR required; squash or merge commit OK |
| Review | ≥1 human review before merge to `main` for prod-bound changes |
| Secrets | Never commit `.env`, credentials, model weights, or live `values/` / PII scans |

Product behavior changes should reference the plan under `docs/plans/` or an issue.

## 2. Environments

| Env | API | UI | LLM |
| --- | --- | --- | --- |
| Local | `DIGITIZER_ALLOW_STUB=1 ./bin/digitizer serve` | `npm run dev` or `npm run preview` | Stub OK |
| CI | stub + fixture smoke | Playwright mocked API | Stub |
| Staging | Real or shared llama.cpp | Prod static build | Real preferred |
| Production | Real LLM, no stub | Prod static build (Rollup) | Required |

Promotion path: **Local → CI green → Staging sign-off → Production**.

## 3. Quality gates (must pass before `main` merge)

### 3.1 API repo (`document-labeler`)

| Gate | Command / check | Owner |
| --- | --- | --- |
| Unit / package tests | `go test ./...` with `DIGITIZER_ALLOW_STUB=1` | CI + author |
| Binary build | `go build -o bin/digitizer ./cmd/digitizer` | CI |
| Fixture smoke | `digitizer scan` on `invoice.png` + `packing_slip.png` | CI |
| OCR regression (when OCR touched) | Re-scan a known phone photo or fixture; confirm text not gibberish | Author |
| No stub in prod config | Review deploy env for `DIGITIZER_ALLOW_STUB` absent | Release |

Workflow: `.github/workflows/ci.yml`

### 3.2 UI repo (`document-labeler-modules-ui`)

| Gate | Command / check | Owner |
| --- | --- | --- |
| Typecheck + prod bundle | `npm run build` (tsc → Rollup → Babel) | CI |
| E2E | `npm test` (Playwright against prod `dist/`, mocked API) | CI |
| Cross-stack smoke (staging) | Upload fixture via UI → real `/v1/scan` | Release |

Workflow: `.github/workflows/ci.yml`  
UI-only notes: [modules-ui docs/SDLC.md](https://github.com/Shriraman-Hegde/document-labeler-modules-ui/blob/main/docs/SDLC.md)

### 3.3 Cross-repo contract

Any change to `/v1/scan`, `/v1/relate`, multipart fields, or CORS must:

1. Update API tests under `internal/server`.
2. Update UI `src/lib/api.ts` and Playwright specs if the contract shifts.
3. Note the change in the release report ([templates/release-report.md](./templates/release-report.md)).

## 4. Release process

1. **Freeze** — both `main`s green on CI.
2. **Tag** — API: `vX.Y.Z`; UI: `vX.Y.Z` (independent semver OK; record both in the release report).
3. **Build artifacts**
   - API: container (`docker compose build digitizer`) or binary for the target OS.
   - UI: `npm ci && npm run build` → publish `dist/`.
4. **Staging deploy** — API + UI + LLM; run [templates/staging-signoff.md](./templates/staging-signoff.md).
5. **Production deploy** — API first (health OK), then UI pointing at prod API URL.
6. **Post-deploy** — `/health` OK; one real scan; `/v1/relate` for a known dept; fill release report.
7. **Rollback** — redeploy previous tags; keep prior `dist/` and binary/image immutable.

## 5. Operations loops (ongoing SDLC)

| Cadence | Activity | Report |
| --- | --- | --- |
| Per PR | CI + review | PR checks |
| Per release | Staging sign-off + release report | [REPORTS.md](./REPORTS.md) §1–2 |
| Weekly | Draft/conflict review (`digitizer conflicts`) | Weekly catalog report |
| Weekly | Unknown / rescan queue triage | Weekly quality report |
| Monthly | Dependency / CVE glance (Go modules, npm) | Security notes in monthly ops |

## 6. Definition of done (feature)

- [ ] Behavior matches agreed plan or issue
- [ ] Tests updated; CI green on both affected repos
- [ ] README / prod docs updated if contract or ops changed
- [ ] No secrets or live PII committed
- [ ] Staging sign-off attached when the change is user-visible in prod
