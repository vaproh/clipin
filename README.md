# ClipIN

India's performance clipping marketplace.

## Stack

### Frontend

- Bun
- Nuxt 3
- Vue 3
- TypeScript
- Tailwind CSS
- shadcn-vue
- Motion for Vue
- Lucide
- TanStack Query
- Zod
- VueUse
- Clerk (authentication)

### Backend

- Go 1.24
- chi
- Huma (OpenAPI)
- pgx
- sqlc
- PostgreSQL 16
- Redis 7

### Testing

- Go `testing` + `httptest`
- Vitest + @vue/test-utils + happy-dom
- Playwright (E2E + visual QA)

### External

- Clerk Auth
- Razorpay (stubbed)
- Cloudflare

## Local development

Prerequisites:

- Bun
- Go 1.24+
- Docker & Docker Compose
- just

Start databases:

```bash
just db-up
```

Run frontend dev server:

```bash
just dev-frontend
```

Run backend dev services:

```bash
just dev-backend
```

Run everything:

```bash
just dev
```

Individual services:

```bash
just web        # Nuxt frontend (http://localhost:3000)
just api        # Go API (http://localhost:8080)
just verifier   # Go verifier (http://localhost:8081)
```

Database migrations:

```bash
just migrate            # run all pending
just migrate status     # check current version
just migrate down       # rollback one
```

## Testing

Run all unit/integration tests:

```bash
just test
```

Individual test suites:

```bash
just test-api       # Go backend (191 tests)
just test-web       # Vitest frontend (23 tests)
just test-e2e       # Playwright E2E (57 tests)
```

Update visual QA baselines:

```bash
just test-e2e-update
```

## Current scope

Open marketplace for performance clipping campaigns. Content owners fund escrow pools, clippers publish short-form clips, verified views drive earnings and UPI payouts.

All milestones M0-M9 are complete. Mobile-optimized. Testing infrastructure in place. Deployment excluded until production infra exists.

| Milestone | Deliverable | Status |
|---|---|---|
| M0 | Foundation: migrations, sqlc, Clerk JWT, CORS, CI | Done |
| M1 | Users, roles, onboarding, shared UI primitives | Done |
| M2 | Campaign marketplace (list, filters, detail, cache) | Done |
| M3 | Campaign creation wizard + owner dashboard | Done |
| M4 | Submissions (lifecycle, review, dedupe, auto-approve) | Done |
| M5 | Verification contract (metric snapshots, deltas) | Done |
| M6 | Append-only financial ledger (idempotent entries) | Done |
| M7 | Payouts (UPI, stubbed Razorpay, webhook) | Done |
| M8 | Admin controls, fraud flags, audit logs, rate limiting | Done |
| M9 | Launch polish (SEO, meta tags, build verification) | Done |
| Mobile | Hamburger drawer, touch targets, overflow fixes | Done |
| Testing | Vitest unit tests, Playwright E2E + visual QA | Done |

## Repository structure

```text
apps/
  web/                  # Nuxt 3 frontend
    components/         # Vue components (app/, campaign/, landing/, ledger/, shared/, submission/, ui/)
    composables/        # TanStack Query composables
    e2e/                # Playwright E2E + visual QA tests
    layouts/            # Nuxt layouts (default, app)
    pages/              # Nuxt pages (landing, auth, app/*)
    lib/                # Utilities (formatPaise, cn)
    assets/css/         # Tailwind + design tokens
    vitest.config.ts    # Vitest configuration
    playwright.config.ts # Playwright configuration

  api/                  # Go API server
    cmd/api/            # API entrypoint
    cmd/migrate/        # Migration CLI (up/down/status/goto)
    db/migrations/      # SQL migrations (000001-000008)
    db/query.sql        # sqlc query definitions
    internal/
      auth/             # JWT + session middleware
      config/           # Environment configuration
      db/sqlc/          # Generated models + queries
      http/handlers/    # Huma handlers (health, user, campaign, submission, verification, ledger, payout, admin)
      middleware/        # Rate limiter
      payout/           # PayoutProvider interface + Razorpay stub
      redis/            # Redis client
      service/          # Business logic (campaign, submission, verification, ledger, payout, audit, fraud)
      worker/           # Background workers (auto-approve, payout processor)

services/
  verifier/             # View verification service (externally owned)

docs/
  architecture.md
  verification-contract.md

PRD.md                  # Product requirements
AGENTS.md               # Engineering principles + execution plan
justfile                # Developer commands
docker-compose.yml      # PostgreSQL + Redis
.env.example            # Environment template
```
