#!/bin/sh
set -eu
repo_root=$(CDPATH= cd -- "$(dirname -- "$0")/../.." && pwd)
cd "$repo_root"
export GOTELEMETRY=off GOTOOLCHAIN=local GOWORK=off
export GOCACHE="$repo_root/.local/go-cache" GOMODCACHE="$repo_root/.local/go-mod-cache"
go run ./scripts/secrets/generate.go
