#!/bin/sh
# Starts as root so a fresh volume at /data (root-owned) can be handed to the
# server. The Event Logs process itself is uid 65532.
# PORT is what Render and the healthcheck use. The binary listens on
# SERVER_PORT so a checkout .env PORT=9090 cannot steal this port.
# ponytail: chown -R on every start. Ceiling is a large /data tree; upgrade
# path is to skip the walk when the mount is already uid 65532.
set -eu

PORT="${PORT:-29909}"
case "$PORT" in
  ''|*[!0-9]*)
    echo "PORT must be a number" >&2
    exit 1
    ;;
esac
export SERVER_PORT="$PORT"

if [ "$(id -u)" = "0" ]; then
  mkdir -p /data/uploads /data/formsx_badger
  chown -R formx:formx /data
  exec su-exec formx "$@"
fi

exec "$@"
