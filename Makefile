.PHONY: secrets up down logs check config

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

config: secrets
	docker compose config --quiet
