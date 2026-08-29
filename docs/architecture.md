# Architecture

## System overview

```
                    ┌──────────────────┐
                    │   Nuxt 3 (web)   │
                    │  Vue 3 + TS      │
                    │  Clerk (auth)    │
                    │  TanStack Query  │
                    └────────┬─────────┘
                             │ Clerk JWT
                    ┌────────▼─────────┐
                    │    Go API (chi)   │
                    │  Huma (OpenAPI)   │
                    │  pgx + sqlc       │
                    └──┬────────────┬──┘
                       │            │
              ┌────────▼──┐  ┌──────▼──────────┐
              │ PostgreSQL │  │     Redis        │
              │ (business  │  │ (cache, rate     │
              │  truth)    │  │  limits, locks)  │
              └────────────┘  └──────────────────┘

                    ┌──────────────────┐
                    │  services/        │
                    │  verifier         │
                    │  (writes to       │
                    │  metric_snapshots)│
                    └────────┬─────────┘
                             │
                    ┌────────▼─────────┐
                    │   PostgreSQL      │
                    │   (metric_snapshots)
                    └──────────────────┘
```

## Services

| Service | Location | Owner | Purpose |
|---|---|---|---|
| Web frontend | `apps/web/` | ClipIN | Nuxt 3 SSR app, Clerk auth, marketplace UI |
| API server | `apps/api/` | ClipIN | Go backend, business logic, financial ledger |
| Verifier | `services/verifier/` | External | Fetches social metrics, writes snapshots to shared DB (stub only, see TODO.md) |

## Data flow

1. **Campaign creation**: owner creates campaign (validated, platform fee calculated) → campaign goes live with funded status.
2. **Clipper submission**: clipper submits post URL → dedup check, clipper limit check → pending → owner reviews (approve/reject) or auto-approve after N hours.
3. **Verification**: verifier polls submission URLs → writes snapshots to `metric_snapshots` via internal API (X-Verifier-Key auth).
4. **Earnings**: main backend reads snapshots → computes eligible views (delta × CPM) → creates idempotent ledger entries → updates campaign remaining budget.
5. **Payout**: clipper requests withdrawal (₹500 min) → Razorpay stub processes → ledger updated with payout entry.

## Key boundaries

- **Verifier writes snapshots, nothing else.** It does not own balances, budgets, payouts, or eligibility. The main backend decides what verified metrics mean financially.
- **PostgreSQL is the source of truth** for business state. Redis caches public data and handles rate limits / locks.
- **Clerk owns authentication.** The API verifies JWTs via JWKS; the frontend uses Clerk's hosted UI.
- **Razorpay is stubbed.** A `PayoutProvider` interface allows swapping in real implementation later.

## Database ownership

- `public.*` (users, profiles, roles, social_accounts, campaigns, submissions, ledger_entries, withdrawals, audit_logs) — owned by the main API.
- `public.metric_snapshots` — written by the verifier, read by the main API for earnings computation.

## Testing

- **Backend**: Go `testing` + `httptest` (337 tests across 13 packages). Postgres integration tests (82 tests, real database). Service tests mock DB interfaces.
- **Frontend**: Vitest + @vue/test-utils (49 tests). Utility and component smoke tests.
- **E2E**: Playwright with desktop + mobile projects (57 tests). Clerk stub module for testing without live auth instance. Visual QA with screenshot baselines.

Test commands:
```bash
just test-api        # Go backend tests
just test-web        # Vitest frontend tests
just test-e2e        # Playwright E2E + visual QA
just test            # all unit/integration tests
```
