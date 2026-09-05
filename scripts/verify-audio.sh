#!/usr/bin/env bash
set -euo pipefail

echo "Checking ALSA devices..."
aplay -l

echo "Checking PipeWire status..."
systemctl --user is-active pipewire
systemctl --user is-active wireplumber

echo "Checking Bluetooth service..."
systemctl is-active bluetooth

echo "Checking mpv availability..."
command -v mpv >/dev/null

echo "Audio baseline verification complete."
