# Release report (R1)

| Field | Value |
| --- | --- |
| Date (UTC) | |
| Operator | |
| Environment | staging / production |
| API repo SHA / tag | |
| UI repo SHA / tag | |
| API CI run URL | |
| UI CI run URL | |
| Staging sign-off (R2) link | |
| LLM mode | llama.cpp URL / other (stub **forbidden** in prod) |

## Smoke

| Check | Result | Notes |
| --- | --- | --- |
| `GET /health` | pass / fail | |
| `POST /v1/scan` (fixture or approved sample) | pass / fail | type / rescan |
| `GET /v1/relate?dept=` | pass / fail | |
| UI load + scan or relate | pass / fail | |

## Rollback

| Component | Previous tag / SHA |
| --- | --- |
| API | |
| UI | |

## Notes / risks

-
