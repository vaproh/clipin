# TODO - Next Phase

## Priority 1: Verifier Service

The verifier is the critical missing piece. Without it, no views get verified and nobody gets paid.

### Architecture

- Verifier runs as a separate Go service (at `services/verifier/`)
- It polls for approved submissions via `POST /internal/snapshots/list` (internal API key auth)
- For each submission, fetches the post URL from the social platform
- Extracts view/like/comment/share counts
- Normalizes metrics into a standard schema
- Writes snapshots via `POST /internal/snapshots` (internal API key auth)
- The main backend reads snapshots and computes eligible views, earnings, etc.

### Phase 1: YouTube Shorts support (DONE)

- [x] Implement YouTube Data API v3 integration
- [x] Parse YouTube Shorts URLs (youtube.com/shorts/VIDEO_ID, youtube.com/watch?v=, youtu.be/VIDEO_ID)
- [x] Call `videos.list` endpoint (1 unit per request)
- [x] Extract: viewCount, likeCount, commentCount
- [x] Handle private/deleted/unavailable videos gracefully
- [x] Store YouTube API key in env config (`YOUTUBE_API_KEY`)
- [x] Implement snapshot submission pipeline
- [x] Poll `/internal/snapshots/list` for submissions needing verification
- [x] For each: fetch metrics, create normalized snapshot
- [x] POST snapshot to main API
- [x] Add polling worker (background goroutine, configurable interval)
- [x] Write tests (mock YouTube API responses)
- [ ] Deploy verifier as separate Docker container

### Phase 2: Instagram Reels support (DONE)

Ported the 3-call anonymous GraphQL protocol from `reference/ig_views.py` to Go.

- [x] Port the 3-call method to Go (`internal/provider/instagram.go`)
- [x] Parse Instagram Reel/Post URLs (reels/, reel/, p/, tv/, ?igsh junk)
- [x] Extract: play_count (views), like_count, comment_count
- [x] Handle private/deleted/not-in-feed gracefully
- [x] 429 backoff and retry
- [x] Write tests (mock GraphQL responses via httptest)

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
