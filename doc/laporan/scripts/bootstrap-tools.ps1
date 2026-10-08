# Portable tools only; no system-wide installation or PATH changes.
$ErrorActionPreference = 'Stop'
$repoRoot = [IO.Path]::GetFullPath((Join-Path $PSScriptRoot '../../..'))
$toolsDir = Join-Path $repoRoot '.local/report-tools'
New-Item -ItemType Directory -Force -Path $toolsDir | Out-Null

function Get-ReleaseAsset([string]$Repository, [string]$Tag, [string]$Name, [string]$Destination) {
    $release = Invoke-RestMethod "https://api.github.com/repos/$Repository/releases/tags/$Tag"
    $asset = $release.assets | Where-Object name -eq $Name
    if (-not $asset) { throw "Official release asset missing: $Name" }
    if (-not (Test-Path -LiteralPath $Destination)) {
        Invoke-WebRequest $asset.browser_download_url -OutFile $Destination
    }
    $actual = (Get-FileHash -LiteralPath $Destination -Algorithm SHA256).Hash.ToLower()
    if ($asset.digest -and $asset.digest.StartsWith('sha256:') -and $asset.digest -ne "sha256:$actual") {
        throw "Release checksum mismatch: $Name"
    }
    Write-Output "$Name SHA256=$actual"
}

$archive = Join-Path $toolsDir 'tectonic.zip'
Get-ReleaseAsset 'tectonic-typesetting/tectonic' 'tectonic@0.17.0' 'tectonic-0.17.0-x86_64-pc-windows-msvc.zip' $archive
Expand-Archive -LiteralPath $archive -DestinationPath (Join-Path $toolsDir 'tectonic') -Force
Get-ReleaseAsset 'plantuml/plantuml' 'v1.2026.8' 'plantuml.jar' (Join-Path $toolsDir 'plantuml.jar')
Write-Output 'Tools ready. Java and Python are needed only when regenerating diagram/chart assets.'
