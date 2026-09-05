#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

cd "${ROOT_DIR}/backend"
go mod tidy
GOOS=linux GOARCH=arm64 go build -o "${ROOT_DIR}/bin/mediaplayer-backend" ./cmd/mediaplayer

echo "Backend built at bin/mediaplayer-backend"
