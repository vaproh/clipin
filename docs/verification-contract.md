# Verification Contract

The `services/verifier` is a separate worker that submits snapshots to the main API. This document defines the write contract.

The verifier is implemented at `services/verifier/` with YouTube and Instagram providers. It polls the internal list endpoint, fetches metrics, and writes snapshots through the internal snapshot endpoint.

## The verifier's only job

Given a submitted social post URL, retrieve and normalize available public metrics.

The verifier does **not** own:

- balances
- campaign budgets
- payout decisions
- campaign eligibility
- earnings calculations

The main backend decides what verified metrics mean financially.

## Write target: `metric_snapshots` table

The main API writes rows to the `public.metric_snapshots` table from verifier submissions.

### Table schema

```sql
CREATE TABLE metric_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id   UUID NOT NULL REFERENCES submissions(id),
    views           BIGINT NOT NULL CHECK (views >= 0),
    likes           BIGINT NOT NULL DEFAULT 0 CHECK (likes >= 0),
    comments        BIGINT NOT NULL DEFAULT 0 CHECK (comments >= 0),
    shares          BIGINT NOT NULL DEFAULT 0 CHECK (shares >= 0),
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_metric_snapshots_submission_id ON metric_snapshots(submission_id);
CREATE INDEX idx_metric_snapshots_captured_at ON metric_snapshots(captured_at);
```

### Write rules

1. **Append-only.** Never update or delete existing snapshots. A new snapshot always adds a new row.
2. **One row per successful fetch.** A poll that cannot retrieve metrics does not write a snapshot and is retried later.
3. **Values are point-in-time.** Views/likes/comments/shares are the platform-reported values at `captured_at`, not deltas.
4. **`submission_id` is a foreign key.** The verifier receives submission IDs and writes snapshots against them. The verifier does not create or modify submissions.
5. **Never NULL on core metrics.** `views` must always be provided. `likes`, `comments`, `shares` default to 0 if the platform doesn't expose them.

### What the main backend computes from snapshots

The main backend (not the verifier) reads snapshots and computes:

- **Growth deltas**: `snapshot[n].views - snapshot[n-1].views` for each submission.
- **Engagement ratios**: `likes / views`, `comments / views` to detect anomalies.
- **Eligible views per window**: new verified views that meet the campaign's rules (view floor, per-clip cap, remaining pool).
- **Earnings**: eligible views × campaign CPM rate.

### Polling cadence (recommended)

| Clip view tier | Poll interval |
|---|---|
| Cold (0–999 views) | Every 12 hours |
| Warm (1,000–9,999) | Every 6 hours |
| Hot (10,000–99,999) | Every 4 hours |
| Viral (100,000+) | Every 2 hours |

Adjust based on platform rate limits and cost. These are starting points.

### Platform-specific notes

| Platform | Method | Reliability |
|---|---|---|
| YouTube Shorts | Data API v3 `videos.list` (1 unit per call, exact integer counts) | High |
| Instagram Reels | Unauthenticated GraphQL (reverse-engineered, no API key) | Best-effort; may break |

If a platform fetch fails, the verifier should not write a snapshot. The main backend handles missing snapshots gracefully.

The verifier exposes unauthenticated operational endpoints for container and
Prometheus monitoring:

- `GET /health` — process liveness
- `GET /metrics` — poll cycles, provider failures, snapshot writes, and latest poll timestamps

### Error handling

- If the verifier cannot fetch metrics (network error, rate limit, private post, deleted post), it should **not** write a snapshot.
- If a post becomes private or deleted, the verifier stops polling that submission and optionally writes a final `post_unavailable` signal (future extension — not required for V1).
- The main backend tolerates gaps in snapshot history gracefully (eligible views for windows with no new snapshots are zero).
