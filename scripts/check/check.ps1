$ErrorActionPreference = 'Stop'
$repoRoot = [System.IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../..'))
$env:GOTELEMETRY = 'off'
$env:GOTOOLCHAIN = 'local'
$env:GOWORK = 'off'
$env:GOCACHE = Join-Path $repoRoot '.local/go-cache'
$env:GOMODCACHE = Join-Path $repoRoot '.local/go-mod-cache'
Push-Location $repoRoot
try {
    go test ./scripts/secrets/generate.go ./scripts/secrets/generate_test.go
    if ($LASTEXITCODE -ne 0) { throw 'Secret generator tests failed.' }
    foreach ($module in (Get-ChildItem services -Filter go.mod -Recurse)) {
        Push-Location $module.DirectoryName
        try {
            Write-Host "Checking $($module.Directory.Name)"
            go test ./...
            if ($LASTEXITCODE -ne 0) { throw 'Tests failed.' }
            go vet ./...
            if ($LASTEXITCODE -ne 0) { throw 'Vet failed.' }
        } finally { Pop-Location }
    }
} finally { Pop-Location }
