#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo bash scripts/setup-services.sh"
  exit 1
fi

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT_DIR="$(cd "${SCRIPT_DIR}/.." && pwd)"

install -d -m 0755 /opt/mediaplayer/bin
rsync -a --delete "${ROOT_DIR}/backend/" /opt/mediaplayer/backend/
rsync -a --delete "${ROOT_DIR}/ui/" /opt/mediaplayer/ui/
if [[ -f "${ROOT_DIR}/bin/mediaplayer-backend" ]]; then
  install -m 0755 "${ROOT_DIR}/bin/mediaplayer-backend" /opt/mediaplayer/bin/mediaplayer-backend
fi

install -d -m 0755 /etc/wireplumber/wireplumber.conf.d
install -m 0644 "${ROOT_DIR}/deploy/pipewire/wireplumber-bluetooth.lua.d/51-mediaplayer-policy.lua" \
  /etc/wireplumber/wireplumber.conf.d/51-mediaplayer-policy.lua

install -m 0644 "${ROOT_DIR}/deploy/systemd/mediaplayer-mpv.service" /etc/systemd/system/mediaplayer-mpv.service
install -m 0644 "${ROOT_DIR}/deploy/systemd/mediaplayer-backend.service" /etc/systemd/system/mediaplayer-backend.service
install -m 0644 "${ROOT_DIR}/deploy/systemd/mediaplayer-ui.service" /etc/systemd/system/mediaplayer-ui.service
install -m 0644 "${ROOT_DIR}/deploy/systemd/mediaplayer-kiosk.service" /etc/systemd/system/mediaplayer-kiosk.service

systemctl daemon-reload
systemctl enable mediaplayer-mpv.service
systemctl enable mediaplayer-backend.service
systemctl enable mediaplayer-ui.service
systemctl enable mediaplayer-kiosk.service

echo "Services installed and enabled."
