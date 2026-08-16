#!/usr/bin/env bash
# Auto-deploy orchestrator for the robo platform (Render + Alibaba Cloud).
#
# Usage:
#   ./scripts/deploy.sh help
#   ./scripts/deploy.sh doctor
#   ./scripts/deploy.sh build --all
#   ./scripts/deploy.sh build --app=morph,formx,userspanel
#   ./scripts/deploy.sh package
#   ./scripts/deploy.sh docker build --all --tag=robo
#   ./scripts/deploy.sh render validate|apply|status
#   ./scripts/deploy.sh alibaba package|sync|push-acr|print-topology
#   ./scripts/deploy.sh env check
#   ./scripts/deploy.sh list
#
# See DEPLOY-README.md for topology, secrets, and cloud setup.
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DEPLOY_DIR="${ROOT}/deploy"
DIST_DIR="${ROOT}/dist/deploy"
ENV_FILE="${DEPLOY_DIR}/.env.production"
ENV_EXAMPLE="${DEPLOY_DIR}/env.production.example"
RENDER_YAML="${DEPLOY_DIR}/render.yaml"
DOCKER_DIR="${DEPLOY_DIR}/docker"

TAG="${DEPLOY_TAG:-latest}"
REGISTRY_PREFIX=""
APPS_FILTER=""
SYNC_HOST=""
SYNC_PATH="/opt/robo"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
NC='\033[0m'

log()  { echo -e "${CYAN}▶${NC} $*"; }
ok()   { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}!${NC} $*"; }
err()  { echo -e "${RED}✗${NC} $*" >&2; }

ALL_APPS=(
  userspanel
  morph
  formx
  composerx
  booki
  morph-engi
  morph-utils
  sharpreport
)

# Apps that produce Docker API images
DOCKER_APPS=(
  userspanel
  morph
  formx
  composerx
  booki
  morph-engi
  sharpreport
)

usage() {
  cat <<'EOF'
Auto-deploy orchestrator for the robo platform (Render + Alibaba Cloud).
See DEPLOY-README.md for topology, secrets, and cloud setup.

Usage:
  ./scripts/deploy.sh <command> [flags]

Commands:
  doctor              Check Go / Node / Rust / Docker / optional CLIs
  list                List deployable apps
  build               Build release binaries + SPA dist
  package             Pack dist/deploy/robo-deploy-<ts>.tar.gz
  docker build        Build container images (context = repo root)
  render validate     Check deploy/render.yaml exists and is readable
  render apply        render blueprints launch (requires render CLI)
  render status       render services list (requires render CLI)
  alibaba package     Linux-oriented package (same as package + docker hints)
  alibaba sync        rsync package tree to ECS (--host=user@ip [--path=/opt/robo])
  alibaba push-acr    docker push images to ACR (--registry=registry.../ns)
  alibaba print-topology  Print recommended Alibaba service map
  env check           Validate deploy/.env.production required keys
  help                Show this help

Flags:
  --all                 All apps
  --app=a,b,c           Subset of apps
  --tag=NAME            Image tag (default: latest or $DEPLOY_TAG)
  --registry=PREFIX     Image registry prefix for docker / push-acr
  --host=user@host      ECS SSH target for alibaba sync
  --path=/opt/robo      Remote path for alibaba sync
EOF
}

have() { command -v "$1" >/dev/null 2>&1; }

parse_global_flags() {
  local arg
  for arg in "$@"; do
    case "$arg" in
      --all) APPS_FILTER="*" ;;
      --app=*) APPS_FILTER="${arg#--app=}" ;;
      --tag=*) TAG="${arg#--tag=}" ;;
      --registry=*) REGISTRY_PREFIX="${arg#--registry=}" ;;
      --host=*) SYNC_HOST="${arg#--host=}" ;;
      --path=*) SYNC_PATH="${arg#--path=}" ;;
    esac
  done
}

selected_apps() {
  local want="$1"
  local app
  if [[ -z "$APPS_FILTER" || "$APPS_FILTER" == "*" ]]; then
    printf '%s\n' "${ALL_APPS[@]}"
    return
  fi
  IFS=',' read -r -a parts <<<"$APPS_FILTER"
  for app in "${parts[@]}"; do
    app="$(echo "$app" | tr '[:upper:]' '[:lower:]' | xargs)"
    [[ -n "$app" ]] || continue
    printf '%s\n' "$app"
  done
}

app_in_selection() {
  local needle="$1"
  local a
  while IFS= read -r a; do
    [[ "$a" == "$needle" ]] && return 0
  done < <(selected_apps)
  return 1
}

cmd_doctor() {
  log "Checking toolchains…"
  local missing=0
  if have go; then ok "go $(go version | awk '{print $3}')"; else err "go missing"; missing=1; fi
  if have node; then ok "node $(node -v)"; else err "node missing"; missing=1; fi
  if have npm; then ok "npm $(npm -v)"; else err "npm missing"; missing=1; fi
  if have cargo; then ok "cargo $(cargo --version | awk '{print $2}')"; else warn "cargo missing (UsersPanel / Engi / SharpReport)"; fi
  if have docker; then ok "docker $(docker --version | awk '{print $3}' | tr -d ',')"; else warn "docker missing (needed for image builds)"; fi
  if have render; then ok "render CLI present"; else warn "render CLI optional — https://render.com/docs/cli"; fi
  if have rsync; then ok "rsync present"; else warn "rsync missing (alibaba sync)"; fi
  if [[ -f "$ENV_FILE" ]]; then ok "found $ENV_FILE"; else warn "missing $ENV_FILE — copy from env.production.example"; fi
  if [[ -f "$RENDER_YAML" ]]; then ok "found render.yaml"; else err "missing render.yaml"; missing=1; fi
  if [[ -d "$DOCKER_DIR" ]]; then ok "found deploy/docker"; else err "missing deploy/docker"; missing=1; fi
  [[ "$missing" -eq 0 ]] || exit 1
  ok "doctor passed"
}

cmd_list() {
  echo "Deployable apps:"
  local a
  for a in "${ALL_APPS[@]}"; do
    echo "  - $a"
  done
  echo
  echo "Docker API images:"
  for a in "${DOCKER_APPS[@]}"; do
    echo "  - $a  (deploy/docker/${a}.Dockerfile)"
  done
}

build_userspanel() {
  log "Building UsersPanel API…"
  (cd "${ROOT}/UsersPanel/backend" && cargo build --release)
  log "Building UsersPanel Admin…"
  (cd "${ROOT}/UsersPanel/admin" && npm ci && npm run build)
  ok "userspanel"
}

build_morph() {
  log "Building Morph frontend…"
  (cd "${ROOT}/morph/frontend" && npm ci && npm run build)
  log "Building Morph API…"
  (cd "${ROOT}/morph" && go build -o "${DIST_DIR}/bin/morph-server" main.go)
  ok "morph"
}

build_formx() {
  log "Building FormsX API…"
  mkdir -p "${DIST_DIR}/bin"
  (cd "${ROOT}/formx/backend" && go build -o "${DIST_DIR}/bin/formsx-server" ./cmd/server)
  log "Building FormsX UI…"
  (cd "${ROOT}/formx/frontend" && npm ci && npm run build)
  ok "formx"
}

build_composerx() {
  log "Building ComposerX API…"
  mkdir -p "${DIST_DIR}/bin"
  (cd "${ROOT}/composerx/backend" && go build -o "${DIST_DIR}/bin/composerx-server" .)
  log "Building ComposerX UI…"
  (cd "${ROOT}/composerx/frontend" && npm ci && npm run build)
  ok "composerx"
}

build_booki() {
  log "Building Booki API…"
  mkdir -p "${DIST_DIR}/bin"
  (cd "${ROOT}/booki/backend" && go build -o "${DIST_DIR}/bin/booki-server" ./cmd/server)
  log "Building Booki UI…"
  (cd "${ROOT}/booki/frontend" && npm ci && npm run build)
  ok "booki"
}

build_morph_engi() {
  log "Building Morph Engi API…"
  (cd "${ROOT}/morph-engi/backend" && cargo build --release)
  log "Building Morph Engi UI…"
  (cd "${ROOT}/morph-engi/frontend" && npm ci && npm run build)
  ok "morph-engi"
}


build_morph_utils() {
  log "Building Morph Utils UI…"
  (cd "${ROOT}/morph-utils/frontend" && npm ci && npm run build)
  ok "morph-utils"
}

build_sharpreport() {
  log "Building SharpReport (cargo release)…"
  (cd "${ROOT}/SharpReport/backend" && cargo build --release)
  if [[ -d "${ROOT}/SharpReport/frontend" ]]; then
    (cd "${ROOT}/SharpReport/frontend" && npm ci && npm run build)
  fi
  ok "sharpreport (prefer SharpReport/deploy Compose for full Metabase stack)"
}

cmd_build() {
  mkdir -p "${DIST_DIR}/bin"
  local app
  while IFS= read -r app; do
    case "$app" in
      userspanel) build_userspanel ;;
      morph) build_morph ;;
      formx) build_formx ;;
      composerx) build_composerx ;;
      booki) build_booki ;;
      morph-engi) build_morph_engi ;;
      morph-utils) build_morph_utils ;;
      sharpreport) build_sharpreport ;;
      *) warn "unknown app: $app" ;;
    esac
  done < <(selected_apps)
  ok "build complete → ${DIST_DIR}"
}

cmd_package() {
  mkdir -p "${DIST_DIR}"
  local ts staging archive
  ts="$(date +%Y%m%d-%H%M%S)"
  staging="${DIST_DIR}/staging-${ts}"
  archive="${DIST_DIR}/robo-deploy-${ts}.tar.gz"
  rm -rf "$staging"
  mkdir -p "$staging"/{bin,apps,deploy}

  log "Assembling package tree…"
  cp -R "${DEPLOY_DIR}/." "${staging}/deploy/"
  [[ -d "${DIST_DIR}/bin" ]] && cp -R "${DIST_DIR}/bin/." "${staging}/bin/" 2>/dev/null || true

  # Copy release binaries from cargo targets when present
  [[ -f "${ROOT}/UsersPanel/backend/target/release/users-panel-api" ]] && \
    cp "${ROOT}/UsersPanel/backend/target/release/users-panel-api" "${staging}/bin/"
  [[ -f "${ROOT}/morph-engi/backend/target/release/morph-engi-api" ]] && \
    cp "${ROOT}/morph-engi/backend/target/release/morph-engi-api" "${staging}/bin/"
  [[ -f "${ROOT}/SharpReport/backend/target/release/datapulse" ]] && \
    cp "${ROOT}/SharpReport/backend/target/release/datapulse" "${staging}/bin/"

  # SPA dist folders
  copy_dist() {
    local src="$1" dest="$2"
    if [[ -d "$src" ]]; then
      mkdir -p "$(dirname "$dest")"
      cp -R "$src" "$dest"
    fi
  }
  copy_dist "${ROOT}/UsersPanel/admin/dist" "${staging}/apps/userspanel-admin"
  copy_dist "${ROOT}/morph/frontend/build" "${staging}/apps/morph-ui"
  copy_dist "${ROOT}/formx/frontend/dist" "${staging}/apps/formx-ui"
  copy_dist "${ROOT}/composerx/frontend/dist" "${staging}/apps/composerx-ui"
  copy_dist "${ROOT}/booki/frontend/dist" "${staging}/apps/booki-ui"
  copy_dist "${ROOT}/morph-engi/frontend/dist" "${staging}/apps/morph-engi-ui"
  copy_dist "${ROOT}/morph-utils/frontend/dist" "${staging}/apps/morph-utils-ui"
  copy_dist "${ROOT}/SharpReport/frontend/build" "${staging}/apps/sharpreport-ui"
  copy_dist "${ROOT}/SharpReport/frontend/.svelte-kit/output/client" "${staging}/apps/sharpreport-ui-kit" 2>/dev/null || true

  if [[ -f "$ENV_EXAMPLE" ]]; then
    cp "$ENV_EXAMPLE" "${staging}/deploy/env.production.example"
  fi
  if [[ -f "$ENV_FILE" ]]; then
    warn "Not packing deploy/.env.production (secrets). Copy onto the host separately."
  fi

  cat >"${staging}/DEPLOY.txt" <<EOF
robo deploy package ${ts}
See deploy/DEPLOY-README path in repo: DEPLOY-README.md
On ECS: sudo bash deploy/alibaba/ecs-bootstrap.sh
EOF

  tar -C "$staging" -czf "$archive" .
  rm -rf "$staging"
  ok "package → $archive"
  echo "$archive"
}

image_name() {
  local app="$1"
  if [[ -n "$REGISTRY_PREFIX" ]]; then
    echo "${REGISTRY_PREFIX%/}/${app}:${TAG}"
  else
    echo "robo/${app}:${TAG}"
  fi
}

cmd_docker_build() {
  have docker || { err "docker required"; exit 1; }
  local app df img
  local targets=()
  if [[ -z "$APPS_FILTER" || "$APPS_FILTER" == "*" ]]; then
    targets=("${DOCKER_APPS[@]}")
  else
    while IFS= read -r app; do targets+=("$app"); done < <(selected_apps)
  fi
  for app in "${targets[@]}"; do
    df="${DOCKER_DIR}/${app}.Dockerfile"
    if [[ ! -f "$df" ]]; then
      warn "skip $app — missing $df"
      continue
    fi
    img="$(image_name "$app")"
    log "docker build $img"
    docker build -f "$df" -t "$img" "$ROOT"
    ok "$img"
  done
}

cmd_render() {
  local sub="${1:-validate}"
  case "$sub" in
    validate)
      [[ -f "$RENDER_YAML" ]] || { err "missing $RENDER_YAML"; exit 1; }
      ok "render.yaml present at $RENDER_YAML"
      if have python3; then
        python3 - <<PY
import pathlib,sys
p=pathlib.Path("${RENDER_YAML}")
text=p.read_text()
assert "services:" in text, "services: missing"
print("basic yaml structure ok")
PY
      fi
      warn "Create a Blueprint in Render Dashboard pointing at deploy/render.yaml, or run: render apply"
      ;;
    apply)
      have render || { err "install Render CLI: https://render.com/docs/cli"; exit 1; }
      log "Launching Render blueprint…"
      (cd "$ROOT" && render blueprints launch "$RENDER_YAML")
      ;;
    status)
      have render || { err "install Render CLI"; exit 1; }
      render services
      ;;
    *)
      err "unknown render subcommand: $sub"
      exit 1
      ;;
  esac
}

cmd_alibaba() {
  local sub="${1:-package}"
  case "$sub" in
    package)
      cmd_package
      cat <<EOF

Next (ECS):
  ./scripts/deploy.sh alibaba sync --host=user@ecs-ip --path=/opt/robo
  ssh user@ecs-ip 'sudo bash /opt/robo/deploy/alibaba/ecs-bootstrap.sh'

Next (SAE / ACR):
  ./scripts/deploy.sh docker build --all --registry=registry.cn-hangzhou.aliyuncs.com/<ns>/robo
  ./scripts/deploy.sh alibaba push-acr --registry=registry.cn-hangzhou.aliyuncs.com/<ns>/robo
EOF
      ;;
    sync)
      [[ -n "$SYNC_HOST" ]] || { err "pass --host=user@ecs-ip"; exit 1; }
      have rsync || { err "rsync required"; exit 1; }
      local latest
      latest="$(ls -1t "${DIST_DIR}"/robo-deploy-*.tar.gz 2>/dev/null | head -1 || true)"
      if [[ -z "$latest" ]]; then
        log "No package found — building one…"
        cmd_package
        latest="$(ls -1t "${DIST_DIR}"/robo-deploy-*.tar.gz | head -1)"
      fi
      log "Syncing deploy assets to ${SYNC_HOST}:${SYNC_PATH}"
      ssh "$SYNC_HOST" "mkdir -p '${SYNC_PATH}'"
      rsync -avz --progress \
        "$latest" \
        "${DEPLOY_DIR}/" \
        "${ROOT}/DEPLOY-README.md" \
        "${SYNC_HOST}:${SYNC_PATH}/"
      # Also sync built bin/ if present
      if [[ -d "${DIST_DIR}/bin" ]]; then
        rsync -avz "${DIST_DIR}/bin/" "${SYNC_HOST}:${SYNC_PATH}/bin/"
      fi
      ok "synced. On host: cd ${SYNC_PATH} && sudo tar -xzf robo-deploy-*.tar.gz && sudo bash deploy/alibaba/ecs-bootstrap.sh"
      ;;
    push-acr)
      have docker || { err "docker required"; exit 1; }
      [[ -n "$REGISTRY_PREFIX" ]] || { err "pass --registry=registry.cn-xxx.aliyuncs.com/<ns>/robo"; exit 1; }
      local app img
      for app in "${DOCKER_APPS[@]}"; do
        img="$(image_name "$app")"
        if docker image inspect "$img" >/dev/null 2>&1; then
          log "push $img"
          docker push "$img"
        else
          warn "image not local: $img (run docker build first)"
        fi
      done
      ok "ACR push done"
      ;;
    print-topology)
      cat <<'EOF'
Alibaba recommended topology
============================
VPC
 ├─ RDS MySQL (tran)
 ├─ ApsaraDB MongoDB (athena, alterathena)
 ├─ Tair Redis
 ├─ OSS + CDN (SPA dist + uploads optional)
 ├─ ECS / SAE
 │   ├─ userspanel-api   (first)
 │   ├─ morph-api        (+ disk for Badger)
 │   ├─ formx-api        (+ uploads volume)
 │   ├─ composerx-api    (+ storage volume)
 │   ├─ booki-api
 │   ├─ morph-engi-api
 │   └─ morph-utils (static)
 └─ ECS (large) SharpReport + Metabase Compose
ALB/SLB terminates HTTPS → APIs; CDN serves SPAs
EOF
      ;;
    *)
      err "unknown alibaba subcommand: $sub"
      exit 1
      ;;
  esac
}

REQUIRED_ENV_KEYS=(
  USERS_PANEL_BASE_URL
  MORPH_AI_API_KEY
  JWT_SECRET
  TRAN_MYSQL_DSN
  TRAN_MONGO_URI
)

cmd_env_check() {
  if [[ ! -f "$ENV_FILE" ]]; then
    err "Missing $ENV_FILE"
    echo "Copy: cp $ENV_EXAMPLE $ENV_FILE"
    exit 1
  fi
  local key missing=0 val
  for key in "${REQUIRED_ENV_KEYS[@]}"; do
    val="$(grep -E "^${key}=" "$ENV_FILE" | tail -1 | cut -d= -f2- || true)"
    if [[ -z "$val" || "$val" == "changeme" || "$val" == "REPLACE_ME" ]]; then
      err "unset or placeholder: $key"
      missing=1
    else
      ok "$key set"
    fi
  done
  [[ "$missing" -eq 0 ]] || exit 1
  ok "env check passed"
}

main() {
  local cmd="${1:-help}"
  shift || true
  parse_global_flags "$@"
  # Also allow flags after subcommands (e.g. render apply --tag=…)
  case "$cmd" in
    help|-h|--help) usage ;;
    doctor) cmd_doctor ;;
    list) cmd_list ;;
    build)
      if [[ -z "$APPS_FILTER" ]]; then APPS_FILTER="*"; fi
      cmd_build
      ;;
    package) cmd_package ;;
    docker)
      local dsub="${1:-build}"
      shift || true
      parse_global_flags "$@"
      case "$dsub" in
        build)
          if [[ -z "$APPS_FILTER" ]]; then APPS_FILTER="*"; fi
          cmd_docker_build
          ;;
        *) err "usage: deploy.sh docker build [--all|--app=…]"; exit 1 ;;
      esac
      ;;
    render)
      local rsub="${1:-validate}"
      shift || true
      parse_global_flags "$@"
      cmd_render "$rsub"
      ;;
    alibaba)
      local asub="${1:-package}"
      shift || true
      parse_global_flags "$@"
      cmd_alibaba "$asub"
      ;;
    env)
      local esub="${1:-check}"
      shift || true
      case "$esub" in
        check) cmd_env_check ;;
        *) err "usage: deploy.sh env check"; exit 1 ;;
      esac
      ;;
    *)
      err "unknown command: $cmd"
      usage
      exit 1
      ;;
  esac
}

main "$@"
