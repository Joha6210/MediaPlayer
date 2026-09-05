#!/bin/sh
set -e

systemctl daemon-reload
systemctl enable mediaplayer-mpv.service mediaplayer-backend.service mediaplayer-kiosk.service

if systemctl is-active --quiet mediaplayer-mpv.service; then
  systemctl restart mediaplayer-mpv.service
else
  systemctl start mediaplayer-mpv.service
fi

if systemctl is-active --quiet mediaplayer-backend.service; then
  systemctl restart mediaplayer-backend.service
else
  systemctl start mediaplayer-backend.service
fi