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

All milestones M0-M9 are complete. Deployment excluded until production infra exists.

| Milestone | Deliverable | Status |
|---|---|---|
| M0 | Foundation: migrations, sqlc, Clerk JWT middleware, CI | Done |
| M1 | Users, roles, onboarding | Done |
| M2 | Campaign marketplace (list, filters, detail) | Done |
| M3 | Campaign creation + owner dashboard | Done |
| M4 | Submissions (lifecycle, review, dedupe) | Done |
| M5 | Verification contract (snapshots, deltas, eligible views) | Done |
| M6 | Append-only financial ledger | Done |
| M7 | Payouts (UPI, stubbed Razorpay) | Done |
| M8 | Admin controls, fraud flags, audit logs | Done |
| M9 | Launch polish (SEO, meta tags, build verification) | Done |

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
