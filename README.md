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

Phase 1 scaffolding, public landing page (Linear design system), Clerk authentication, and minimal `/app` application shell. Product marketplace functionality will be implemented phase by phase according to `PRD.md`.

## Repository structure

```text
apps/
  web/
  api/

services/
  verifier/

infra/
  docker/

docs/

PRD.md
AGENTS.md
README.md
justfile
docker-compose.yml
```
