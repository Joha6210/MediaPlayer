$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$backendDir = Join-Path $repoRoot "backend"
$goExe = "C:\Program Files\Go\bin\go.exe"

if (-not (Test-Path $goExe)) {
    throw "Go not found at $goExe"
}

Set-Location $backendDir
& $goExe test .\...
