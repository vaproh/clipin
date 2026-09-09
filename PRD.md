# ClipIN Product Requirements Document

## 1. Product

**Name:** ClipIN  
**Domain:** `clipin.pro`

**Positioning:** India's performance clipping marketplace.

ClipIN connects content owners with clippers who turn existing long-form content into short-form videos, publish them on social platforms, and earn based on verified performance.

The initial marketplace is focused on:

> **Content owner → clipper → short-form distribution → verified views → payout**

This is not an influencer marketplace in V1.

## 2. Problem

People in India are already interested in clipping as a side hustle, but discovery and access are fragmented across foreign platforms, Discords, DMs, and informal communities.

At the same time, Indian creators and other content owners have large amounts of existing long-form content that can be redistributed as Shorts/Reels.

The gap ClipIN targets is a centralized, transparent Indian marketplace where:

- content owners can launch performance-based clipping campaigns;
- clippers can discover active campaigns;
- clips can be submitted for verification;
- verified views can translate into earnings;
- payouts can be made through Indian payment infrastructure.

## 3. Target Users

### Clippers

People who can create and publish short-form clips, including:

- video editors
- students
- side-hustlers
- meme/Shorts/Reels page owners
- micro-creators
- freelance editors

Their primary goal is:

> Find campaigns → clip → publish → earn from verified views.

### Campaign Owners

Initially:

- podcasters
- YouTubers
- creators
- publishers with clip-worthy long-form content

Later:

- agencies
- startups
- brands
- apps

Their primary goal is:

> Provide content → define CPM/budget/rules → receive distributed short-form reach.

## 4. Core User Flows

### Clipper

1. Sign up
2. Complete profile
3. Add/link social account(s)
4. Browse campaigns
5. Open campaign
6. Access source content
7. Create and publish clip externally
8. Submit post URL
9. Platform verifies metrics
10. Eligible views become earnings
11. Request payout
12. Receive payout

### Campaign Owner

1. Sign up
2. Complete profile
3. Create campaign
4. Provide source URL or external asset link
5. Define CPM, budget, dates, platforms, audience, and rules
6. Submit campaign for approval
7. Fund campaign
8. Monitor submissions and verified views
9. Campaign continues until budget/end conditions are reached
10. Review campaign performance

### Admin

1. Review campaign submissions
2. Moderate users/submissions
3. Inspect verification results
4. Flag or restrict suspicious activity
5. Approve/manage payouts
6. Audit financial and marketplace activity

## 5. MVP Scope

### Frontend

- Public landing page
- Authentication integration
- Role onboarding
- Clipper dashboard
- Campaign marketplace
- Campaign detail page
- Campaign creation flow
- Submission flow
- Earnings/wallet view
- Withdrawal request flow
- Campaign owner dashboard
- Basic admin interface
- Responsive web experience

### Backend

- User/profile management
- Roles and authorization
- Campaign lifecycle
- Campaign rules
- Submission lifecycle
- Social-account records
- View snapshots
- Earnings ledger
- Campaign budget accounting
- Payout records
- Notifications
- Audit logs

### Infrastructure

- Cloudflare for DNS/CDN/WAF/Turnstile
- Mumbai VPS for initial application infrastructure
- Go API
- Go background workers
- PostgreSQL
- Redis
- Separate verification service when justified
- External auth provider
- Razorpay for payments/payouts

## 6. Verification Service

The verification service is a separate service.

Its responsibility is narrow:

> Given a submitted social post URL, retrieve and normalize available public metrics.

It should not own:

- campaign budgets
- user balances
- payout logic
- eligibility rules
- marketplace business logic

The main ClipIN backend decides how retrieved metrics affect campaign accounting and earnings.

Initial implementation may combine automated verification with manual review.

## 7. Payments and Ledger

The platform must use an append-only financial ledger as the source of truth.

Do not rely on a single mutable user balance as the authoritative record.

Ledger events may include:

- eligible earnings
- reversals
- adjustments
- withdrawal requests
- payout completion
- payout failure

Campaign budgets and clipper earnings must be auditable.

Razorpay handles payment/payout execution; ClipIN remains responsible for internal accounting.

## 8. Abuse and Fraud

V1 should prioritize practical controls over complex ML.

Initial controls can include:

- authentication and role enforcement
- Cloudflare Turnstile
- API rate limits
- Redis locks/deduplication
- duplicate submission detection
- suspicious view-growth flags
- public/private/deleted post checks
- campaign expiry checks
- manual admin review
- immutable audit logs

The system should be designed so more sophisticated fraud detection can be added later.

## 9. UX / Visual Direction

The product should feel like a serious piece of financial/marketplace infrastructure rather than a generic creator-economy landing page.

Design direction:

- Linear-inspired
- monochrome-first
- restrained use of color
- strong typography
- generous spacing
- precise data presentation
- subtle borders
- minimal shadows
- purposeful animation
- no fake social proof
- no fake earnings
- no fake traction numbers

The interface should emphasize:

- CPM
- campaign budget
- verified views
- earnings
- payout status
- campaign rules

## 10. Frontend Stack

- Bun
- Nuxt
- Vue 3
- TypeScript
- Tailwind CSS
- shadcn-vue
- Motion for Vue
- Lucide
- TanStack Query for Vue
- Zod
- VueUse

No Android app in the initial product.

## 11. Backend Stack

- Go
- `net/http`
- chi
- Huma
- pgx
- sqlc
- PostgreSQL
- Redis
- Go background workers

External:

- Auth provider
- Razorpay
- social-platform verification mechanisms

Supported launch platforms:

- YouTube Shorts
- Instagram Reels

TikTok is outside the India-focused product scope.

## 12. Initial Deployment

Initial infrastructure is intentionally minimal.

Cloudflare:

- DNS
- TLS
- CDN/edge caching
- WAF
- Turnstile

Mumbai VPS:

- Go API
- Go workers
- PostgreSQL
- Redis

The verification service may initially share infrastructure and later move to a separate VPS.

No Kubernetes, managed database, managed Redis, media hosting, or complex cloud architecture is required for the initial stage.

## 13. Campaign Content

ClipIN does not host or process large video files by default.

Campaign owners should generally provide:

- existing public video URLs
- or external file-sharing URLs such as Google Drive when necessary

ClipIN primarily stores campaign metadata and links.

## 14. Non-Goals for V1

Do not build:

- Android app
- built-in video editor
- AI clip generation
- video hosting platform
- brand/influencer marketplace
- advanced recommendation engine
- ML-based fraud system
- complex creator analytics suite
- public API
- enterprise team features
- international marketplace

## 15. Phased Development

### M0 — Foundation (completes Phase 1)

- Migration runner (`golang-migrate`) + baseline schema
- sqlc code generation pipeline
- Clerk JWT verification middleware (JWKS) + frontend token flow
- CORS restricted via env allowlist
- GitHub Actions CI (Go test/lint/build, Nuxt typecheck/build)
- Delete stray root `layouts/app.vue`
- `docs/architecture.md`

### M1 — Users & Roles

- Schema: `users` (clerk_id), `profiles`, roles, `social_accounts`
- Endpoints: `GET/PUT /me`, role selection, social-account CRUD
- Authorization layer (clipper / owner / admin)
- Onboarding UI (role selection after sign-up)
- First TanStack Query composables + Zod schemas
- Install shadcn-vue; build shared patterns as needed (PageHeader, StatCard, StatusBadge, Money formatter)

### M2 — Campaign Marketplace

- Campaign state machine: `draft → pending_approval → funded → live → ended/cancelled`
- Schema carries: `reward_pool` (escrowed), `rate_per_1k`, `remaining_budget` (publicly visible)
- Rules struct: allowed platforms, required tags/watermark/disclosure text, min/max duration, min view floor per clip, max payout per clip, max payout per clipper
- Brief + rejection criteria, source content links
- Public list/detail endpoints (live campaigns only), Redis cache
- Marketplace listing + filters + detail page

### M3 — Campaign Creation + Owner Dashboard

- Creation wizard (all rules fields + deposit step — stub payment, ledger records escrow)
- Admin approval workflow (approve / reject campaign)
- Owner dashboard: submissions table, spend vs pool, remaining budget

### M4 — Submissions

- Submission lifecycle: `submitted → pending_review → approved/rejected → earning → settled`
- Unique `(campaign_id, post_url)` constraint + cross-campaign same-URL dedupe
- Rate limits on submission creation
- Owner review with configurable auto-approve-after-N-hours (default 48h)
- Clipper submit UI + submission history

### M5 — Verification Contract

- `metric_snapshots` table (submission_id FK, views/likes/comments/shares, captured_at, append-only)
- Submission state machine extended for verification transitions
- Growth deltas + engagement-ratio + eligible-views math computed by main backend
- `docs/verification-contract.md` defining the snapshot write contract
- Verification status surfaced in clipper and owner UIs
- The `services/verifier` is a separate Go worker operated with the ClipIN stack. It polls the main API, fetches YouTube and Instagram metrics, and submits snapshots through the internal API. It does not write PostgreSQL directly and does not own financial logic.

### M6 — Earnings / Ledger

- Append-only `ledger_entries`; balances always derived, never mutated
- Eligible-view calculation: new verified views × campaign rate, floor/caps enforcement, hard-capped by remaining pool
- Platform fee (10% on deposits) recorded as separate ledger entries
- Unspent-budget refund entries on campaign end
- Idempotency keys on every entry
- TDD mandatory: arithmetic, caps, concurrency, budget exhaustion edge cases

### M7 — Payouts

- UPI details storage, withdrawal requests, ₹500 minimum threshold
- Weekly batch payout cycle (default cadence)
- `PayoutProvider` interface + stub implementation (records intent/completion in ledger)
- Razorpay client drops in behind the same interface when keys are available
- Webhook endpoint built (signature-checked, disabled until Razorpay keys exist)
- Reconciliation job skeleton

### M8 — Admin & Fraud

- Admin UI: campaign approval queue, submission moderation, payout review, user management, audit-log viewer
- Audit log written at every financial/state transition
- Fraud controls: engagement-ratio anomaly flags, duplicate-URL detection, suspicious-growth flags
- Turnstile on sensitive forms, Redis rate limits

### M9 — Launch Polish

- SEO / meta tags, notifications (in-app + email)
- Onboarding refinement, performance improvements
- Dual-pay marketing line for clippers (keep own platform monetization + campaign payout)
- Deployment planned separately when production infra exists

Execution order is strictly sequential. Within each milestone: tests first, small conventional commits, migrations clean, builds green. Playwright visual QA on every new screen before the milestone closes.

## 16. Platform Fee Model

- Flat percentage on campaign deposits: **10%**
- Charged at deposit time (escrow funds include the fee)
- One number, both sides see it transparently
- Swappable as a single constant if the rate needs to change later

## 17. Success Criteria for Early Launch

The product should aim to reach:

- real clippers using the platform
- real campaign owners submitting campaigns
- first funded campaign
- first verified views
- first real clipper payouts
- repeat campaign owners

All traction metrics must come from real platform activity.

No fabricated campaign numbers, users, earnings, testimonials, or social proof.

## 18. Core Product Principle

ClipIN should compete on **actual marketplace liquidity and trust**, not on inflated landing-page numbers.

Every major metric displayed publicly should be backed by a real event in the system.

## Current Status (as of latest push)

All milestones M0-M9 are complete. Additional features built:

- Clipper analytics dashboard
- Notification system (7 triggers)
- Clipper reputation/leveling (4 tiers)
- RazorpayX SDK integration (test mode)
- Playwright E2E + visual QA testing
- Redis caching (6 caches)
- Verifier worker with YouTube and Instagram providers
- PWA foundation with install flow, service worker, offline indicator, and mobile navigation
- Verifier Docker image, CI workflow, healthcheck, and Prometheus metrics
- 82 Postgres integration tests
- 358 API tests, 63 verifier tests, 49 frontend tests, and 57 Playwright tests

The verifier is implemented with YouTube and Instagram providers. The PWA foundation, install flow, service worker, offline indicator, mobile navigation, Docker image, CI, and Prometheus monitoring are also implemented. See [TODO.md](TODO.md) for remaining Web Push, offline queue, production deployment, and growth work.
