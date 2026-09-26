#!/usr/bin/env bash
set -euo pipefail

if [[ "${EUID}" -ne 0 ]]; then
  echo "Run as root: sudo bash scripts/setup-audio.sh"
  exit 1
fi

apt-get update
apt-get install -y --no-install-recommends \
  pipewire \
  pipewire-audio \
  wireplumber \
  bluez \
  bluez-tools \
  mpv \
  xserver-xorg \
  xinit \
  chromium-browser \
  jq \
  curl

BOOT_CONFIG="/boot/firmware/config.txt"
if [[ -f "/boot/config.txt" ]]; then
  BOOT_CONFIG="/boot/config.txt"
fi

if ! grep -q "^dtoverlay=hifiberry-dacplus$" "${BOOT_CONFIG}"; then
  printf "\n# MediaPlayer\n" >> "${BOOT_CONFIG}"
  printf "dtoverlay=hifiberry-dacplus\n" >> "${BOOT_CONFIG}"
fi

if ! grep -q "^dtparam=audio=off$" "${BOOT_CONFIG}"; then
  printf "dtparam=audio=off\n" >> "${BOOT_CONFIG}"
fi

install -d -m 0755 /etc/bluetooth
cat >/etc/bluetooth/main.conf <<'EOF'
[General]
Name = MediaPlayer
Class = 0x20041C
DiscoverableTimeout = 0
PairableTimeout = 0
AutoEnable=true

[Policy]
AutoEnable=true
EOF

systemctl enable bluetooth

install -d -m 0755 /etc/X11
cat >/etc/X11/Xwrapper.config <<'EOF'
allowed_users=anybody
needs_root_rights=yes
EOF

echo "Audio setup complete. Reboot required."
