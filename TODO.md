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

### Phase 2: Instagram Reels support

- [ ] Instagram Basic Display API or GraphQL scraping
- [ ] Parse Instagram Reel URLs
- [ ] Extract metrics
- [ ] Handle auth (Instagram requires app review for production)
- [ ] Write tests

### Phase 3: TikTok support

- [ ] TikTok Research API (requires application)
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
