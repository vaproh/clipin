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

- Auth provider
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

Start the web app:

```bash
just web
```

Start the API:

```bash
just api
```

Start the verifier:

```bash
just verifier
```

Run checks:

```bash
just test
just lint
just format
```

## Current scope

This repository is initially scaffolding only. Product functionality will be implemented phase by phase according to `PRD.md`.

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
