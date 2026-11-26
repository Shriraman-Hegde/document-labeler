# Staging sign-off (R2)

| Field | Value |
| --- | --- |
| Date (UTC) | |
| Signer | |
| API SHA / tag | |
| UI SHA / tag | |
| Staging API URL | |
| Staging UI URL | |
| `DIGITIZER_ALLOW_STUB` | must be unset / 0 |

## Checklist

- [ ] CI green on both repos for these SHAs
- [ ] `/health` returns `ok`
- [ ] UI opens and talks to staging API (CORS OK)
- [ ] File upload scan succeeds (`rescan=false` on a known good manufacturing fixture when LLM healthy)
- [ ] Relate returns without 5xx for a dept with values
- [ ] Camera path checked **or** waived (reason: ________)
- [ ] No secrets in browser network panel beyond intended API

**Sign-off:** I approve promotion of these builds to production.

Signature / name: _______________  Date: _______________
