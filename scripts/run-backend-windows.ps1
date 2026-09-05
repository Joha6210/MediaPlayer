$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$configPath = Join-Path $backendDir "config\windows-test.yaml"
$goExe = "C:\Program Files\Go\bin\go.exe"

if (-not (Test-Path $goExe)) {
    throw "Go not found at $goExe"
}

Set-Location $backendDir
$env:MEDIAPLAYER_CONFIG = $configPath

& $goExe mod tidy
if (-not $?) {
    throw "go mod tidy failed"
}

& $goExe run .\cmd\mediaplayer
