.PHONY: secrets up down logs check config smoke ingest-check events-check query-check load-check

# POSIX Make entry point: let containers read the current user's 0600 secrets.
export LOCAL_UID := $(shell id -u)
export LOCAL_GID := $(shell id -g)

secrets:
	sh scripts/secrets/generate.sh

up: secrets
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f --tail=100

check:
	sh scripts/check/check.sh

smoke:
	GOTELEMETRY=off GOTOOLCHAIN=local GOWORK=off go test ./scripts/check/foundation_test.go -v -count=1 -timeout=8m

ingest-check:
	docker compose run --build --rm --env-from-file ./env/aggregator.env ingest-test
	GOTELEMETRY=off GOTOOLCHAIN=local GOWORK=off go test ./scripts/check/foundation_test.go ./scripts/check/ingest_test.go -run TestIngestPipeline -v -count=1 -timeout=5m

config: secrets
	docker compose config --quiet

events-check:
	docker compose --profile demo build pemda-portal
	GOTELEMETRY=off GOTOOLCHAIN=local GOWORK=off go test ./scripts/check/foundation_test.go ./scripts/check/events_test.go -run TestEventPipeline -v -count=1 -timeout=10m

query-check:
	GOTELEMETRY=off GOTOOLCHAIN=local GOWORK=off go test ./scripts/check/foundation_test.go ./scripts/check/query_test.go -run 'TestQueryIntegration|TestDynamicSchemaAndStaleAPI|TestNaturalExpiryAndFieldCLI|TestIndependentRebuild' -v -count=1 -timeout=8m

load-check:
	python3 scripts/loadtest/run.py
