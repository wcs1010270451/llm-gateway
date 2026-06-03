CREATE TABLE request_log_bodies (
    id BIGSERIAL PRIMARY KEY,
    request_log_id BIGINT NOT NULL UNIQUE REFERENCES request_logs(id) ON DELETE CASCADE,
    request_body_path TEXT NOT NULL DEFAULT '',
    response_body_path TEXT NOT NULL DEFAULT '',
    request_body_size BIGINT NOT NULL DEFAULT 0,
    response_body_size BIGINT NOT NULL DEFAULT 0,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_request_log_bodies_expires_at ON request_log_bodies(expires_at);

COMMENT ON TABLE request_log_bodies IS 'Large request and response body file references for short-term debugging';
COMMENT ON COLUMN request_log_bodies.request_log_id IS 'Linked request_logs row';
COMMENT ON COLUMN request_log_bodies.request_body_path IS 'Server-side path to the stored request body';
COMMENT ON COLUMN request_log_bodies.response_body_path IS 'Server-side path to the stored response body';
COMMENT ON COLUMN request_log_bodies.expires_at IS 'Time after which files should be deleted and no longer read';
