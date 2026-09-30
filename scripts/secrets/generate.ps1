$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$env:GOTELEMETRY = 'off'
$env:GOTOOLCHAIN = 'local'
$env:GOWORK = 'off'
$env:GOCACHE = Join-Path $repoRoot '.local/go-cache'
$env:GOMODCACHE = Join-Path $repoRoot '.local/go-mod-cache'
Push-Location $repoRoot
try {
    go run ./scripts/secrets/generate.go
    if ($LASTEXITCODE -ne 0) { throw 'Secret bootstrap failed.' }
} finally { Pop-Location }
