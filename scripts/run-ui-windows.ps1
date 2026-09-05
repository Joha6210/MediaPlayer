$ErrorActionPreference = "Stop"

$repoRoot = Split-Path -Parent $PSScriptRoot
$uiDir = Join-Path $repoRoot "ui"

Set-Location $uiDir
npm.cmd install
if (-not $?) {
    throw "npm install failed"
}

npm.cmd run dev -- --host 127.0.0.1 --port 4173
