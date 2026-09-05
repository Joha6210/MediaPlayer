# MediaPlayer

Lightweight Raspberry Pi 5 media player optimized for fast boot and local touchscreen control.

## git cmd

git tag v0.0.1-rc8

git push origin v0.0.1-rc8

## Update on rpi

wget https://github.com/Joha6210/MediaPlayer/releases/download/v0.0.1-rc8/mediaplayer_0.0.1-rc8_linux_arm64.deb
sudo apt install ./mediaplayer_0.0.1-rc8_linux_arm64.deb
sudo systemctl daemon-reload
sudo systemctl restart mediaplayer-mpv mediaplayer-backend

## Stack

- Backend: Go
- UI: SvelteKit (kiosk mode in Chromium)
- Playback engine: mpv JSON IPC daemon
- Audio/Bluetooth: PipeWire + WirePlumber + BlueZ

## Hardware target

- Raspberry Pi 5 (4GB)
- HifiBerry DAC2
- DSI touchscreen

## Quick start (Raspberry Pi OS Lite 64-bit)

1. Clone this repository to the Pi.
2. Run setup scripts:
   - `sudo bash scripts/setup-audio.sh`
   - `sudo bash scripts/setup-services.sh`
3. Build and install apps:
   - `bash scripts/build-backend.sh`
   - `bash scripts/build-ui.sh`
4. Reboot and use kiosk UI on touchscreen.

## Runtime components

- `mediaplayer-backend.service`
- `mediaplayer-mpv.service`
- `mediaplayer-kiosk.service`

All service units are in [deploy/systemd/](/C:/Users/Jensen/Skrivebord/Projekter/MediaPlayer/deploy/systemd).

## Windows local test mode

You can test most application behavior on Windows with mock runtime mode:

- Backend runs without Linux-only dependencies (`mpv` unix socket and `bluetoothctl`).
- DR/Plexamp adapters skip external availability checks.
- Source switching, API responses, WebSocket state updates, and UI flows remain testable.

Steps:

1. Start backend (test mode config):
   - `powershell -ExecutionPolicy Bypass -File scripts/run-backend-windows.ps1`
2. Start UI:
   - `powershell -ExecutionPolicy Bypass -File scripts/run-ui-windows.ps1`
3. Open `http://127.0.0.1:4173`
4. Run backend tests:
   - `powershell -ExecutionPolicy Bypass -File scripts/test-backend-windows.ps1`

## API summary

- `GET /health`
- `GET /api/state`
- `POST /api/source/select`
- `POST /api/player/volume`
- `GET /ws`

`/api/source/select` payload examples:

- Internet radio:
  - `{ "source": "internet-radio", "url": "https://..." }`
- DR:
  - `{ "source": "dr-radio", "meta": { "station": "p3" } }`
- Plexamp:
  - `{ "source": "plexamp", "meta": { "path": "/audio/:/transcode/..." } }`
  - API documentation: https://developer.plex.tv/pms/#section/API-Info
- Bluetooth sink:
  - `{ "source": "bluetooth" }`

## Validation checklist on Pi

1. `bash scripts/build-backend.sh`
2. `bash scripts/build-ui.sh`
3. `sudo bash scripts/setup-audio.sh`
4. `sudo bash scripts/setup-services.sh`
5. Reboot
6. `bash scripts/verify-audio.sh`
7. `bash scripts/measure-boot.sh` and verify total boot target under 30 seconds
