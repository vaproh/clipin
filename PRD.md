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

### Phase 1 — Foundation + Landing + Auth

- repository setup
- frontend scaffold
- backend scaffold
- PostgreSQL/Redis
- landing page
- auth-provider integration

### Phase 2 — User Onboarding

- role selection
- profiles
- social account records
- authorization

### Phase 3 — Campaign Marketplace

- campaign listing
- filters
- campaign detail page
- campaign lifecycle
- caching

### Phase 4 — Campaign Creation

- campaign creation
- CPM/budget/rules
- campaign owner dashboard
- approval workflow

### Phase 5 — Clip Submissions

- submission flow
- submission history
- moderation
- duplicate detection

### Phase 6 — Verification

- verification service
- background jobs
- metric snapshots
- verification status

### Phase 7 — Earnings / Ledger

- eligible views
- earnings calculations
- financial ledger
- campaign budget accounting

### Phase 8 — Payouts

- UPI details
- withdrawal flow
- Razorpay integration
- webhooks
- reconciliation

### Phase 9 — Abuse / Operations

- fraud rules
- admin controls
- audit logs
- operational tooling

### Phase 10 — Launch Polish

- SEO
- notifications
- performance
- onboarding improvements
- campaign discovery improvements

## 16. Success Criteria for Early Launch

The product should aim to reach:

- real clippers using the platform
- real campaign owners submitting campaigns
- first funded campaign
- first verified views
- first real clipper payouts
- repeat campaign owners

All traction metrics must come from real platform activity.

No fabricated campaign numbers, users, earnings, testimonials, or social proof.

## 17. Core Product Principle

ClipIN should compete on **actual marketplace liquidity and trust**, not on inflated landing-page numbers.

Every major metric displayed publicly should be backed by a real event in the system.
