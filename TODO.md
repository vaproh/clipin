# TODO - Next Phase

## Priority 1: Verifier Service

The verifier is the critical missing piece. Without it, no views get verified and nobody gets paid.

### Architecture

- Verifier runs as a separate Go service (already scaffolded at `services/verifier/`)
- It polls for approved submissions via `POST /internal/snapshots/list` (internal API key auth)
- For each submission, fetches the post URL from the social platform
- Extracts view/like/comment/share counts
- Normalizes metrics into a standard schema
- Writes snapshots via `POST /internal/snapshots` (internal API key auth)
- The main backend reads snapshots and computes eligible views, earnings, etc.

### Phase 1: YouTube Shorts support

- [ ] Add database connection to verifier (pgxpool)
- [ ] Add Redis client to verifier (for rate limiting platform APIs)
- [ ] Implement YouTube Data API v3 integration
  - [ ] Parse YouTube Shorts URLs (youtube.com/shorts/VIDEO_ID, youtu.be/VIDEO_ID)
  - [ ] Call `videos.list` endpoint (1 unit per request)
  - [ ] Extract: viewCount, likeCount, commentCount, publishedAt
  - [ ] Handle private/deleted/unavailable videos gracefully
  - [ ] Store YouTube API key in env config
- [ ] Implement snapshot submission pipeline
  - [ ] Poll `/internal/snapshots/list` for submissions needing verification
  - [ ] For each: fetch metrics, create normalized snapshot
  - [ ] POST snapshot to main API
  - [ ] Rate limit: max 100 YouTube API calls per day (free tier)
  - [ ] Retry failed fetches with exponential backoff
- [ ] Add polling worker (background goroutine, configurable interval)
- [ ] Write tests (mock YouTube API responses)
- [ ] Deploy verifier as separate Docker container

### Phase 2: Instagram Reels support (VERIFIED METHOD, 2026-09-04)

Instagram views are obtainable anonymously, free, plain HTTP, no browser, no
paid API. Verified working from a datacenter IP on 2026-09-04.

Method (3 HTTP calls, reference implementation `/tmp` prototype `ig_views.py`):

1. `GET https://www.instagram.com/` -> anonymous `csrftoken` cookie (desktop
   Chrome UA, session persists cookies)
2. `POST /graphql/query` doc_id `27128499623469141` with compact-JSON variables
   `{"shortcode":"...","__relay_internal__pv__PolarisAIGMMediaWebLabelEnabledrelayprovider":false}`
   -> `data.xdt_api__v1__media__shortcode__web_info.items[0]` carries
   `user.pk`, `user.username`, `like_count`, `comment_count`, `code`,
   `taken_at`. (`view_count` is null here - login-gated, ignore it.)
3. `POST /graphql/query` doc_id `27234427476213202` with variables
   `{"data":{"include_feed_video":true,"page_size":12,"target_user_id":"<pk>"}}`
   -> iterate `data.xdt_api__v1__clips__user__connection_v2.edges[].node.media`,
   match `code == shortcode`, read `play_count` (the real "views" counter).

Headers for both GraphQL calls: `X-CSRFToken` (from call 1), `X-IG-App-ID:
936619743392459`, `X-ASBD-ID: 129477`, `X-Requested-With: XMLHttpRequest`,
`Origin`/`Referer: https://www.instagram.com/`, form-urlencoded.

Verified: 3,465,648 views on reference reel, view count ticks in near-real-time
(+45 in 65s), zero authenticated cookies, garbage shortcodes cleanly rejected.

Known pitfalls (all verified):

- `doc_ids` are version-pinned relay queries that rotate. On `execution error`
  refresh from instaloader master (`instaloader/structures.py`, search
  `doc_id_graphql_query`). Disambiguate rotation vs bad shortcode by probing a
  known-good control shortcode.
- The reel must be in the owner's last N reels (page_size 12, retry 50). Older
  posts surface a clean "not in feed" error - acceptable for recent submissions.
  Not finding the reel in feed pages is NOT a hard failure: mark views
  unverified and poll again on the next cycle (rare feed-excluded reels, e.g.
  pinned/restricted, may never appear - those stay unverified, never estimate).
- Rate limits: anonymous GraphQL tolerates a few req/min; keep polling sparse
  (>=60s per reel), reuse the session, back off on 429.
- Do NOT send cookies beyond what call 1 sets. No login, no sessionid, ever.

Tasks:

- [ ] Port the 3-call method to Go in the verifier (`igclient` package)
- [ ] Parse Instagram Reel/Post URLs (reels/, reel/, p/, tv/, ?igsh junk)
- [ ] Extract: play_count (views), like_count, comment_count, username, taken_at
- [ ] Handle private/deleted/not-in-feed gracefully
- [ ] Per-session request pacing (>=60s per reel), 429 backoff
- [ ] Doc-id rotation recovery (config override + control-shortcode probe)
- [ ] Write tests (mock GraphQL responses)


### Phase 3: TikTok support

- [ ] Anonymous page scrape: public video page embeds JSON with views, likes,
  comments, shares (no auth needed - proven by albeethekid/metadata-api)
- [ ] Parse TikTok URLs
- [ ] Extract metrics
- [ ] Write tests

### Infrastructure

- [ ] Add `services/verifier/Dockerfile`
- [ ] Add to docker-compose.yml
- [ ] Add CI workflow for verifier
- [ ] Add health check monitoring

## Priority 2: Mobile App (PWA)

Clippers primarily use phones. A PWA is the fastest path to mobile without native app overhead.

### Phase 1: PWA Foundation

- [ ] Add Nuxt PWA module (`@vite-pwa/nuxt`)
- [ ] Create `public/manifest.json` with app metadata
- [ ] Add service worker for offline caching
- [ ] Add install prompt banner
- [ ] Add splash screens
- [ ] Configure push notifications (Web Push API)
- [ ] Test on Android Chrome + iOS Safari

### Phase 2: Mobile-Optimized UI

- [ ] Bottom tab bar for primary navigation (Campaigns, Submissions, Earnings, Profile)
- [ ] Pull-to-refresh on list pages
- [ ] Swipe gestures for submission approve/reject (owner)
- [ ] Haptic feedback on key actions
- [ ] Camera integration for quick clip submission
- [ ] Deep linking (share campaign URL -> open in app)

### Phase 3: Push Notifications

- [ ] Service worker push handler
- [ ] Notification permissions flow
- [ ] Backend: send push on submission approved/rejected
- [ ] Backend: send push on payout completed
- [ ] Backend: send push on new campaign matching clipper interests
- [ ] Notification preferences (opt-in/out per type)

### Phase 4: Offline Support

- [ ] Cache campaign listings for offline browsing
- [ ] Queue clip submissions when offline
- [ ] Sync when back online
- [ ] Show offline indicator

## Priority 3: Growth Features

### Community

- [ ] Discord/Telegram integration
- [ ] In-app chat or announcements
- [ ] Clipper tips and guides

### Referral System

- [ ] Unique referral links per user
- [ ] Track referral source in ledger
- [ ] Bonus earnings or reduced fees for referrals

### Trust Signals

- [ ] Clipper ratings from campaign owners
- [ ] Campaign ratings from clippers
- [ ] Verified badge for high-reputation users
- [ ] Campaign thumbnails/images
