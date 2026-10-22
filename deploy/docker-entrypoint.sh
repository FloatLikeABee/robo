#!/bin/sh
# Starts as root so a fresh named volume at /data (root-owned) can be handed
# to the server. The Dorph process itself is uid 65532.
# ponytail: chown -R on every start. Ceiling is a large /data tree; upgrade
# path is to skip the walk when the mount is already uid 65532.
set -eu

if [ "$(id -u)" = "0" ]; then
  mkdir -p /data/badger /data/entity_details /data/knowledge /data/uploads/entity_attachments
  chown -R dorph:dorph /data
  exec su-exec dorph "$@"
fi

exec "$@"
