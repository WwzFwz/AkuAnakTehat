ALTER TABLE outbox ADD COLUMN rejected_at TIMESTAMPTZ, ADD COLUMN rejection_reason TEXT;
DROP INDEX outbox_pending_idx;
CREATE INDEX outbox_pending_idx ON outbox(id) WHERE published_at IS NULL AND rejected_at IS NULL;
