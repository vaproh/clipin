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
| Verifier | `services/verifier/` | External | Fetches social metrics, writes snapshots to shared DB |

## Data flow

1. **Campaign creation**: owner creates campaign → admin approves → owner deposits funds → campaign goes live.
2. **Clipper submission**: clipper joins campaign → clips source content → publishes to own account → submits post URL → owner reviews (approve/reject/auto-approve after 48h).
3. **Verification**: verifier service fetches metrics from social platforms → writes snapshots to `metric_snapshots` table.
4. **Earnings**: main backend reads snapshots → computes growth deltas and eligible views per snapshot window → creates ledger entries (earnings minus fee).
5. **Payout**: clipper requests withdrawal → payout provider executes → ledger updated with payout completion/failure.

## Key boundaries

- **Verifier writes snapshots, nothing else.** It does not own balances, budgets, payouts, or eligibility. The main backend decides what verified metrics mean financially.
- **PostgreSQL is the source of truth** for business state. Redis caches public data and handles rate limits / locks.
- **Clerk owns authentication.** The API verifies JWTs via JWKS; the frontend uses Clerk's hosted UI.
- **Razorpay is stubbed.** A `PayoutProvider` interface allows swapping in real implementation later.

## Database ownership

- `public.*` (users, profiles, roles, social_accounts, campaigns, submissions, ledger_entries, withdrawals, audit_logs) — owned by the main API.
- `public.metric_snapshots` — written by the verifier, read by the main API for earnings computation.
