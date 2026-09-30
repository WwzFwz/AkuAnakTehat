CREATE TABLE hazard_events (
 hazard_id UUID PRIMARY KEY,
 source TEXT NOT NULL CHECK(source IN ('BMKG','PVMBG')),
 source_ref_id TEXT NOT NULL,
 hazard_type TEXT NOT NULL CHECK(hazard_type IN ('SEISMIC','VOLCANIC')),
 severity TEXT NOT NULL CHECK(severity IN ('NORMAL','WASPADA','SIAGA','AWAS')),
 area_name TEXT NOT NULL,
 latitude DOUBLE PRECISION NOT NULL CHECK(latitude BETWEEN -90 AND 90),
 longitude DOUBLE PRECISION NOT NULL CHECK(longitude BETWEEN -180 AND 180),
 occurred_at TIMESTAMPTZ NOT NULL,
 ingested_at TIMESTAMPTZ NOT NULL,
 attributes JSONB NOT NULL DEFAULT '{}',
 version BIGINT NOT NULL DEFAULT 1 CHECK(version>0),
 content_hash BYTEA NOT NULL,
 updated_at TIMESTAMPTZ NOT NULL,
 last_seen_at TIMESTAMPTZ NOT NULL,
 UNIQUE(source,source_ref_id)
);
CREATE INDEX hazard_type_time_idx ON hazard_events(hazard_type,occurred_at DESC,hazard_id DESC);
CREATE INDEX hazard_time_idx ON hazard_events(occurred_at DESC,hazard_id DESC);
CREATE TABLE outbox (
 id BIGSERIAL PRIMARY KEY,event_id UUID NOT NULL UNIQUE,hazard_id UUID NOT NULL REFERENCES hazard_events(hazard_id),
 version BIGINT NOT NULL CHECK(version>0),payload JSONB NOT NULL,created_at TIMESTAMPTZ NOT NULL,published_at TIMESTAMPTZ,
 UNIQUE(hazard_id,version)
);
CREATE INDEX outbox_pending_idx ON outbox(id) WHERE published_at IS NULL;
CREATE TABLE checkpoints(endpoint TEXT PRIMARY KEY,watermark TIMESTAMPTZ NOT NULL);
CREATE TABLE tsunami_warnings(warning_id TEXT PRIMARY KEY,related_event_id TEXT NOT NULL,payload JSONB NOT NULL,updated_at TIMESTAMPTZ NOT NULL);
CREATE INDEX tsunami_related_idx ON tsunami_warnings(related_event_id);
CREATE TABLE quarantine(id BIGSERIAL PRIMARY KEY,source TEXT NOT NULL,endpoint TEXT NOT NULL,reason TEXT NOT NULL,correlation_id TEXT NOT NULL,payload JSONB NOT NULL,observed_at TIMESTAMPTZ NOT NULL);
CREATE TABLE source_status(source TEXT PRIMARY KEY,status TEXT NOT NULL CHECK(status IN ('HEALTHY','DEGRADED','DOWN')),last_success_at TIMESTAMPTZ,last_attempt_at TIMESTAMPTZ,consecutive_failures INTEGER NOT NULL DEFAULT 0,last_error TEXT);
-- Durable per-endpoint health avoids one successful BMKG endpoint masking the other.
CREATE TABLE source_endpoint_status(endpoint TEXT PRIMARY KEY,source TEXT NOT NULL REFERENCES source_status(source),healthy BOOLEAN NOT NULL DEFAULT false,last_success_at TIMESTAMPTZ,last_attempt_at TIMESTAMPTZ,consecutive_failures INTEGER NOT NULL DEFAULT 0,last_error TEXT);
INSERT INTO source_status(source,status) VALUES ('BMKG','DOWN'),('PVMBG','DOWN');
INSERT INTO source_endpoint_status(endpoint,source) VALUES ('bmkg.seismic-events','BMKG'),('bmkg.tsunami-warnings','BMKG'),('pvmbg.volcanic-reports','PVMBG');
