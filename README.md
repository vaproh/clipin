# ClipIN

India's performance clipping marketplace.

## Stack

### Frontend

- Bun
- Nuxt
- Vue 3
- TypeScript
- Tailwind CSS
- shadcn-vue
- Motion for Vue
- Lucide
- TanStack Query
- Zod
- VueUse
- Clerk

### Backend

- Go
- net/http
- chi
- Huma
- pgx
- sqlc
- PostgreSQL
- Redis

### External

- Clerk Auth
- Razorpay
- Cloudflare
- Social verification integrations

## Local development

Prerequisites:

- Bun
- Go
- Docker
- Docker Compose
- just

Start databases:

```bash
just db-up
```

Run frontend dev server:

```bash
just dev-frontend
```

Run backend dev services (PostgreSQL, Redis, Go API, Go Verifier):

```bash
just dev-backend
```

Run all dev servers:

```bash
just dev
```

Run individual services:

```bash
just web        # Nuxt frontend
just api        # Go API server
just verifier   # Go verifier service
```

Run checks:

```bash
just test
just lint
just format
just build
```

## Current scope

Open marketplace for performance clipping campaigns. Content owners fund escrow pools, clippers publish short-form clips, verified views drive earnings and UPI payouts.

Phase 1 foundation (scaffolding, landing, Clerk auth, app shell) is built. Remaining execution follows the milestone plan in `AGENTS.md`:

| Milestone | Deliverable |
|---|---|
| M0 | Foundation: migrations, sqlc, Clerk JWT middleware, CI |
| M1 | Users, roles, onboarding |
| M2 | Campaign marketplace (list, filters, detail) |
| M3 | Campaign creation + owner dashboard |
| M4 | Submissions (lifecycle, review, dedupe) |
| M5 | Verification contract (snapshots, deltas, eligible views) |
| M6 | Append-only financial ledger |
| M7 | Payouts (UPI, stubbed Razorpay) |
| M8 | Admin controls, fraud flags, audit logs |
| M9 | Launch polish (SEO, notifications, perf) |

Deployment excluded until production infra exists.

## Repository structure

```text
apps/
  web/          # Nuxt 3 frontend
  api/          # Go API server

services/
  verifier/     # View verification service (owned externally)

infra/
  docker/       # Dockerfiles (empty until deployment)

docs/
  architecture.md
  verification-contract.md

PRD.md          # Product requirements
AGENTS.md       # Engineering principles + execution plan
DESIGN.md       # Design system tokens
README.md
justfile
docker-compose.yml
```
