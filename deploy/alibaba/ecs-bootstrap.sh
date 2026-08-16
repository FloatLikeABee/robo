#!/usr/bin/env bash
# First-boot helper for an Alibaba ECS host that will run robo APIs.
# Run as root after syncing a deploy package to /opt/robo (or pass ROBO_ROOT).
#
#   sudo bash /opt/robo/deploy/alibaba/ecs-bootstrap.sh
#
set -euo pipefail

ROBO_ROOT="${ROBO_ROOT:-/opt/robo}"
INSTALL_DOCKER="${INSTALL_DOCKER:-1}"
INSTALL_NGINX="${INSTALL_NGINX:-1}"

echo "==> robo ECS bootstrap (root=${ROBO_ROOT})"

if [[ "$(id -u)" -ne 0 ]]; then
  echo "Run as root (sudo)." >&2
  exit 1
fi

export DEBIAN_FRONTEND=noninteractive
apt-get update -y
apt-get install -y --no-install-recommends \
  ca-certificates curl rsync tar gzip ufw \
  python3

if [[ "$INSTALL_NGINX" == "1" ]]; then
  apt-get install -y --no-install-recommends nginx
fi

if [[ "$INSTALL_DOCKER" == "1" ]] && ! command -v docker >/dev/null 2>&1; then
  echo "==> Installing Docker Engine…"
  curl -fsSL https://get.docker.com | sh
  systemctl enable --now docker
fi

mkdir -p \
  /var/lib/robo/morph-data \
  /var/lib/robo/formx-uploads \
  /var/lib/robo/composerx-storage \
  /var/log/robo \
  /etc/robo

if [[ -d "${ROBO_ROOT}/deploy/alibaba/systemd" ]]; then
  cp -v "${ROBO_ROOT}/deploy/alibaba/systemd/"*.service /etc/systemd/system/ || true
fi

if [[ -f "${ROBO_ROOT}/deploy/env.production.example" && ! -f /etc/robo/env.production ]]; then
  cp "${ROBO_ROOT}/deploy/env.production.example" /etc/robo/env.production
  echo "Created /etc/robo/env.production — EDIT SECRETS before starting services."
fi

if [[ -f "${ROBO_ROOT}/deploy/nginx/spa-proxy.conf.example" && "$INSTALL_NGINX" == "1" ]]; then
  cp "${ROBO_ROOT}/deploy/nginx/spa-proxy.conf.example" /etc/nginx/sites-available/robo.example
  echo "Nginx example: /etc/nginx/sites-available/robo.example"
fi

systemctl daemon-reload

cat <<EOF

Bootstrap complete.

Next steps:
  1) Edit /etc/robo/env.production
  2) Place Linux binaries in ${ROBO_ROOT}/bin/ (or use Docker images)
  3) systemctl enable --now userspanel-api morph-api formx-api composerx-api
  4) Configure ALB/SSL or enable the nginx site example
  5) SharpReport: use SharpReport/deploy/docker-compose.yml on a larger ECS

See ${ROBO_ROOT}/DEPLOY-README.md (or repo DEPLOY-README.md)
EOF
