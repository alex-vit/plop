param(
  [string]$Version,
  [string]$Arch = "amd64"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$RepoRoot = Split-Path -Parent $PSScriptRoot
$sysoPath = $null

Push-Location $RepoRoot
try {
  if ([string]::IsNullOrWhiteSpace($Version)) {
    $Version = (git describe --tags --always --dirty 2>$null)
    if ([string]::IsNullOrWhiteSpace($Version)) {
      $Version = "dev"
    }
  }

  $outDir = Join-Path $RepoRoot "out/windows"
  New-Item -ItemType Directory -Path $outDir -Force | Out-Null

  $goBin = (go env GOBIN).Trim()
  if ([string]::IsNullOrWhiteSpace($goBin)) {
    $goPath = (go env GOPATH).Trim()
    $goBin = Join-Path $goPath "bin"
  }
  $rsrcExe = Join-Path $goBin "rsrc.exe"
  if (-not (Test-Path $rsrcExe)) {
    go install github.com/akavel/rsrc@latest
  }
  if (-not (Test-Path $rsrcExe)) {
    throw "rsrc.exe was not found at $rsrcExe"
  }

  $sysoPath = Join-Path $RepoRoot "plop_windows_$Arch.syso"
  if (Test-Path $sysoPath) {
    Remove-Item $sysoPath -Force
  }

  & $rsrcExe -ico (Join-Path $RepoRoot "icon/icon.ico") -arch $Arch -o $sysoPath
  if ($LASTEXITCODE -ne 0) {
    throw "rsrc.exe failed to generate icon resource"
  }

  $env:GOOS = "windows"
  $env:GOARCH = $Arch
  $env:CGO_ENABLED = "0"
  $ldflags = "-X github.com/alex-vit/plop/cmd.Version=$Version -H=windowsgui"
  go build -tags noassets -ldflags $ldflags -o (Join-Path $outDir "plop.exe") .
  if ($LASTEXITCODE -ne 0) {
    throw "go build failed"
  }

  Remove-Item $sysoPath -Force
  $sysoPath = $null

  $isccCmd = Get-Command iscc.exe -ErrorAction SilentlyContinue
  $isccExe = if ($isccCmd) { $isccCmd.Source } else { $null }
  if (-not $isccExe) {
    $candidate = Join-Path ${env:ProgramFiles(x86)} "Inno Setup 6\ISCC.exe"
    if (Test-Path $candidate) {
      $isccExe = $candidate
    }
  }
  if (-not $isccExe) {
    throw "Inno Setup compiler (iscc.exe) not found in PATH"
  }

  & $isccExe "/DAppVersion=$Version" "installer.iss"
  if ($LASTEXITCODE -ne 0) {
    throw "Installer build failed"
  }

  $exePath = Join-Path $outDir "plop.exe"
  $setupPath = Join-Path $outDir "plop-setup.exe"
  if (-not (Test-Path $exePath)) {
    throw "Missing executable: $exePath"
  }
  if (-not (Test-Path $setupPath)) {
    throw "Missing installer: $setupPath"
  }

  Write-Host "Built executable: $exePath"
  Write-Host "Built installer:  $setupPath"
}
finally {
  if ($sysoPath -and (Test-Path $sysoPath)) {
    Remove-Item $sysoPath -Force
  }
  Pop-Location
}
