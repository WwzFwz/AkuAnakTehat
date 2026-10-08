DROP INDEX outbox_pending_idx;
ALTER TABLE outbox DROP COLUMN rejected_at, DROP COLUMN rejection_reason;
CREATE INDEX outbox_pending_idx ON outbox(id) WHERE published_at IS NULL;
