# field-cli

`field-cli` is a standalone Go CLI for Tim Lapangan. It communicates only with
`auth-service` and `client-api`; it does not import service modules or access a
database.

## Build

From this directory:

```powershell
go test ./...
go build -o field-cli.exe ./cmd/field-cli
```

## Configuration

Credentials should be supplied through environment variables and are never
written to stdout or logs:

```powershell
$env:FIELD_CLI_CLIENT_ID = "field-team"
$env:FIELD_CLI_CLIENT_SECRET = "<local secret>"
$env:FIELD_CLI_AUTH_URL = "http://127.0.0.1:8090"
$env:FIELD_CLI_API_URL = "http://127.0.0.1:8080"
```

`FIELD_CLI_CLIENT_SECRET` may be replaced by `--client-secret` for local
one-off use. The CLI refreshes access tokens serially because refresh-token
rotation invalidates the previous token. A single retry is made after a `401`.

## Commands

```powershell
./field-cli.exe list --type VOLCANIC --limit 20
./field-cli.exe list --cursor <opaque-cursor>
./field-cli.exe list --raw
./field-cli.exe get <hazard-id>
./field-cli.exe get <hazard-id> --raw
```

Responses are printed as indented JSON. Service error codes are reported
without including response bodies, access tokens, or client secrets.
