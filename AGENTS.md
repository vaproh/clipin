# AGENTS.md

## Project

ClipIN is an India-focused performance clipping marketplace.

Core transaction:

> Content owner → clipping campaign → clipper → published short-form post → verified views → earnings → payout

## Engineering Principles

### 1. Keep the architecture lean

Do not introduce infrastructure because it is fashionable.

Default to:

- Go API
- PostgreSQL
- Redis
- background workers
- Nuxt/Vue frontend
- external auth provider
- Razorpay
- Cloudflare

Avoid Kubernetes, microservices, managed infrastructure, or additional queues unless there is a concrete reason.

### 2. Build in phases

Implement only the current product phase.

Do not prematurely implement:

- advanced fraud systems
- Android
- AI features
- recommendations
- influencer marketplace
- video hosting
- enterprise features

### 3. Backend source of truth

PostgreSQL is authoritative for business state.

Redis is for:

- caching
- queues
- rate limiting
- locks
- temporary state

Never treat Redis as the source of truth for financial records.

### 4. Financial correctness

All money movement must be represented in an auditable ledger.

Prefer:

- transactions
- idempotency keys
- explicit state transitions
- immutable ledger entries
- audit logs

Never implement financial logic as a loose collection of balance mutations.

### 5. Verification separation

The verification service only retrieves/normalizes social metrics.

It must not own:

- balances
- campaign budgets
- payout decisions
- campaign eligibility

The main backend decides what verified metrics mean financially.

### 6. External authentication

Do not implement password storage or custom authentication unless explicitly required.

The application should verify the external auth provider's identity and map it to an internal user.

### 7. API design

- Keep handlers thin.
- Put business logic in service/domain layers.
- Validate inputs at boundaries.
- Return predictable errors.
- Use typed request/response models.
- Keep API contracts documented through Huma/OpenAPI.
- Prefer idempotent operations where retries are possible.

### 8. Database

Use PostgreSQL with `pgx` and `sqlc`.

Prefer SQL-first design over an ORM.

Use:

- migrations
- foreign keys
- constraints
- indexes based on real query patterns
- transactions for multi-row state changes

Do not optimize prematurely.

### 9. Redis and caching

Cache only data that is safe to cache.

Good candidates:

- public campaign lists
- public campaign details
- expensive read-heavy queries

Avoid stale caching for:

- balances
- payout state
- financial ledger
- security-sensitive state

Invalidate or refresh deliberately after writes.

### 10. Frontend

Use:

- Nuxt
- Vue 3
- TypeScript
- Tailwind
- shadcn-vue
- Motion for Vue
- Lucide
- TanStack Query
- Zod
- VueUse

Keep server state in TanStack Query.

Do not introduce Pinia unless the product develops a genuine global client-state requirement.

### 11. Design

ClipIN should feel:

- precise
- modern
- monochrome
- trustworthy
- restrained
- information-dense where appropriate

Avoid:

- fake counters
- fake testimonials
- fake campaign activity
- fake earnings
- excessive gradients
- unnecessary animation
- generic startup templates

### 12. Security

- Never commit secrets.
- Use `.env.example`.
- Validate all external input.
- Rate-limit sensitive endpoints.
- Use Cloudflare Turnstile where appropriate.
- Never trust client-provided earnings/views/balances.
- Treat all social metrics as untrusted until verified.
- Protect webhook endpoints.
- Make payout operations idempotent.
- Log security-sensitive changes.

### 13. Testing

Always follow the TDD development workflow:

1. Always write a test first.
2. Then write the feature code to satisfy the test.
3. Then test, fix, and complete.

#### Backend

- Unit tests with Go `testing` + `httptest` (358 API tests plus 63 verifier tests)
- Postgres integration tests (82 tests, real database)
- Service tests mock DB interfaces
- Financial-critical paths thoroughly tested (ledger arithmetic, budget caps, idempotency)

#### Frontend

- Unit tests with Vitest + @vue/test-utils + happy-dom (49 tests)
- Test composables, utilities, and shared components
- Run: `cd apps/web && npx vitest run`

#### E2E

- Playwright for end-to-end testing + visual QA
- Desktop (1280x720) and mobile (iPhone 13) projects
- Clerk stub module enables testing without a live auth instance
- Run: `cd apps/web && npx playwright test`
- Update baselines: `cd apps/web && npx playwright test --update-snapshots`

### 14. Observability

Prefer simple structured logs first.

Important events:

- authentication failures
- campaign state changes
- submission state changes
- verification failures
- ledger changes
- payout changes
- admin overrides

Do not log passwords, tokens, payment secrets, or sensitive credentials.

The verifier exposes `/health` for liveness and `/metrics` for Prometheus
counters covering poll cycles, provider failures, and recorded snapshots.

### 15. Documentation

Update documentation when architecture or workflows change.

Keep:

- README
- PRD
- AGENTS
- API/OpenAPI documentation
- migration notes where necessary

### 16. Git

Use small, focused commits.

Prefer conventional commit style if practical:

- `feat:`
- `fix:`
- `chore:`
- `refactor:`
- `docs:`
- `test:`

Do not commit generated secrets, local databases, build output, or dependency caches.

## Definition of Done

A feature is not complete merely because the happy path works.

Before marking work complete:

1. Frontend builds.
2. Backend builds.
3. Relevant tests pass.
4. Migrations run cleanly.
5. API contracts are consistent.
6. Errors are handled.
7. Authorization is checked.
8. Financial operations are auditable where relevant.
9. No secrets are committed.
10. Documentation is updated when needed.
11. Playwright visual QA passes on new screens (desktop + mobile widths, all three component states: loading / empty / error).

## Key Decisions

### Marketplace archetype

Open self-serve marketplace. No clipper application or vetting gate. Anyone can sign up, browse campaigns, and submit clips. Verification and fraud controls are the quality moat — not curation.

Supported campaign platforms are YouTube and Instagram. TikTok is not part of the India-focused product scope.

### Campaign mechanics (research-backed)

- Escrow-funded pools: campaigns only go live after deposit. No unfunded campaigns.
- Rate per 1,000 views (CPM), set by campaign owner.
- Remaining budget publicly visible on listings and detail pages.
- Auto-approve-after-N-hours for submissions (default 48h, owner-configurable).
- Per-clip and per-clipper payout caps enforced in rules.
- Minimum view floor per clip enforced in rules.
- Unspent budget refunded to campaign owner on campaign end.

### Platform fee

10% flat on campaign deposits. Charged at deposit time. One constant, swappable later.

### Verification boundary

The `services/verifier` polls the main API for approved submissions, fetches YouTube and Instagram metrics, and writes snapshots through the internal snapshot endpoint. It does not own balances, campaign budgets, payout decisions, or earnings calculations. The main backend computes all financial logic (growth deltas, engagement ratios, eligible views, earnings). See `docs/verification-contract.md` for the write contract and `TODO.md` for remaining PWA capabilities and production operations.

### Payouts

RazorpayX SDK integrated (test mode). `PayoutProvider` interface with real implementation behind it. Webhook endpoint built, signature-checked, disabled until production keys exist.

### UI approach

shadcn-vue installed just-in-time per milestone. Shared primitives (PageHeader, StatCard, StatusBadge, EmptyState) built as each page needs them. Design system tokens in `tailwind.config.js` + `main.css`. The Nuxt PWA module provides installability, a service worker, offline status, and mobile bottom navigation. Vitest unit tests for components and composables. Playwright E2E + visual QA on every new screen (desktop + mobile).

### Deployment

Production deployment remains excluded until the VPS / Cloudflare infrastructure exists. The verifier already has a local Docker image, Compose `app` profile, CI workflow, healthcheck, and Prometheus metrics endpoint.

## Execution Order

Sequential by milestone. Within each: tests first (TDD), small conventional commits, migrations clean, builds green, Playwright visual QA before moving on.

| Milestone | Scope | Status |
|---|---|---|
| M0 | Foundation: migrations, sqlc, Clerk JWT, CI, CORS, docs | Done |
| M1 | Users, roles, onboarding, first shared UI primitives | Done |
| M2 | Campaign marketplace: schema, listing, filters, detail | Done |
| M3 | Campaign creation wizard + owner dashboard | Done |
| M4 | Submissions: lifecycle, review, dedupe | Done |
| M5 | Verification contract: snapshot table, deltas, status | Done |
| M6 | Financial ledger: append-only, eligible views, fee, caps | Done |
| M7 | Payouts: UPI, stubbed provider, webhook endpoint | Done |
| M8 | Admin controls, fraud flags, audit logs | Done |
| M9 | Launch polish: SEO, notifications, perf | Done |

All milestones M0-M9 are complete. The verifier and PWA foundation are implemented; see [TODO.md](TODO.md) for remaining push/offline capabilities and production operations.
