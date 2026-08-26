CREATE TABLE metric_snapshots (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    submission_id   UUID NOT NULL REFERENCES submissions(id),
    platform        TEXT NOT NULL,
    views           BIGINT NOT NULL CHECK (views >= 0),
    likes           BIGINT NOT NULL DEFAULT 0 CHECK (likes >= 0),
    comments        BIGINT NOT NULL DEFAULT 0 CHECK (comments >= 0),
    shares          BIGINT NOT NULL DEFAULT 0 CHECK (shares >= 0),
    captured_at     TIMESTAMPTZ NOT NULL DEFAULT now(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_metric_snapshots_submission_id ON metric_snapshots(submission_id);
CREATE INDEX idx_metric_snapshots_captured_at ON metric_snapshots(captured_at);
