#!/usr/bin/env bash
set -euo pipefail

systemd-analyze
systemd-analyze critical-chain
systemd-analyze blame | head -n 25
