#!/bin/sh
# Starts as root so a fresh disk at /data (root-owned) can be handed to the server.
# ponytail: chown -R on every start. Ceiling is a large /data tree; upgrade
# path is to skip the walk when the mount is already uid 65532.
set -eu

if [ "$(id -u)" = "0" ]; then
  mkdir -p /data
  chown -R sharpreport:sharpreport /data
  exec setpriv --reuid=sharpreport --regid=sharpreport --init-groups --inh-caps=-all "$@"
fi

exec "$@"
