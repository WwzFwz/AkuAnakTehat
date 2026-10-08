param([switch]$RefreshAssets)
$ErrorActionPreference = 'Stop'
$reportRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '..'))
$repoRoot = [IO.Path]::GetFullPath((Join-Path $reportRoot '../..'))
$buildDir = Join-Path $reportRoot 'build'
New-Item -ItemType Directory -Force -Path $buildDir | Out-Null
Push-Location $reportRoot
try {
    if ($RefreshAssets) {
        py scripts/refresh-assets.py
        if ($LASTEXITCODE -ne 0) { throw 'Asset generation failed.' }
    }
    $engine = $env:TECTONIC_BINARY
    if (-not $engine) {
        $command = Get-Command tectonic -ErrorAction SilentlyContinue
        if ($command) { $engine = $command.Source }
    }
    if (-not $engine) { $engine = Join-Path $repoRoot '.local/report-tools/tectonic/tectonic.exe' }
    if (-not (Test-Path -LiteralPath $engine)) {
        throw 'Tectonic not found. Run scripts/bootstrap-tools.ps1 or set TECTONIC_BINARY.'
    }
    & $engine --keep-logs --keep-intermediates --synctex --outdir $buildDir main.tex
    if ($LASTEXITCODE -ne 0) { throw 'LaTeX compilation failed; inspect build/main.log.' }
    $output = Join-Path $reportRoot 'IF4031_M1_AkuAnakTehat.pdf'
    Copy-Item -LiteralPath (Join-Path $buildDir 'main.pdf') -Destination $output -Force
    Write-Output "Report built: $output"
} finally { Pop-Location }
