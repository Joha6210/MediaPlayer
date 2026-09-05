#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo bash scripts/optimize-boot.sh"
  exit 1
fi

systemctl disable apt-daily.service apt-daily-upgrade.service
systemctl mask apt-daily.service apt-daily-upgrade.service

systemctl set-default graphical.target
systemctl daemon-reload

echo "Boot optimization baseline applied."
echo "Measure with: systemd-analyze && systemd-analyze blame | head -n 20"
