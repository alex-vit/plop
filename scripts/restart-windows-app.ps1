#Requires -Version 5.1
[CmdletBinding()]
param(
    [switch]$NoBuild
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot   = Split-Path -Parent $PSScriptRoot
$OutExe     = Join-Path $RepoRoot "out\plop.exe"
$InstallDir = Join-Path $env:LOCALAPPDATA "Plop"
$InstallExe = Join-Path $InstallDir "plop.exe"

# 1. Kill running instance (best effort — not an error if not running)
$proc = Get-Process -Name plop -ErrorAction SilentlyContinue
if ($proc) {
    Write-Host "Stopping running plop instance..."
    $proc | Stop-Process -Force
    $proc | Wait-Process -Timeout 10 -ErrorAction SilentlyContinue
}

# 2. Build (unless -NoBuild)
if (-not $NoBuild) {
    Write-Host "Building..."
    Push-Location $RepoRoot
    try {
        $ldflags = "-H=windowsgui -X main.version=dev"
        go build -tags noassets -ldflags $ldflags -o $OutExe .
        if ($LASTEXITCODE -ne 0) { throw "go build failed" }
    } finally {
        Pop-Location
    }
}

# 3. Copy to install dir
if (-not (Test-Path $InstallDir)) {
    throw "Plop is not installed at $InstallDir - run the installer first"
}
Copy-Item $OutExe $InstallExe -Force

# 4. Relaunch from install dir
Start-Process $InstallExe
Write-Host "Restarted $InstallExe"
