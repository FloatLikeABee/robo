#!/bin/sh
# Starts as root so a fresh disk at /data (root-owned) can be handed to the
# server. The Project process itself is uid 65532.
# ponytail: chown -R on every start. Ceiling is a large /data tree; upgrade
# path is to skip the walk when the mount is already uid 65532.
set -eu

if [ "$(id -u)" = "0" ]; then
  mkdir -p /data/uploads
  chown -R morphengi:morphengi /data
  exec gosu morphengi "$@"
fi

exec "$@"
