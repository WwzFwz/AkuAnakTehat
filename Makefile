.PHONY: secrets up down logs check config smoke

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

config: secrets
	docker compose config --quiet
