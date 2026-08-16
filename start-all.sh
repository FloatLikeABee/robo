#!/usr/bin/env bash
# Dev launcher for robo platform apps (academi excluded).
# Folder map: morph, formx, composerx, booki, UsersPanel, SharpReport, …
#
# Usage:
#   ./start-all.sh                         start everything (foreground; Ctrl+C stops all)
#   ./start-all.sh --install               install deps, then start all
#   ./start-all.sh stop | --stop           stop everything
#   ./start-all.sh start                   start everything (one-shot; skips already running)
#   ./start-all.sh restart                 stop + start everything
#   ./start-all.sh restart <service>       restart one app (or alias)
#   ./start-all.sh start <service>         start one app
#   ./start-all.sh stop <service>          stop one app
#   ./start-all.sh status [service]        show running / stopped
#   ./start-all.sh logs <service>          tail -f log file
#   ./start-all.sh list                    list service names + aliases
#
# Aliases (API + UI): morph, morph-utils, bk, formx, composerx, booki,
#   morph-engi, academi, sharpreport — or `all` for every service below.
# Neo4j: full-stack start/restart ensures bolt port 7687 is up when `neo4j` CLI exists.
# Deprecated alias: userspanel (UsersPanel UI/API — auth moved into Morph).
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="${ROOT}/.robo-dev"
LOG_DIR="${RUN_DIR}/logs"
PID_FILE="${RUN_DIR}/pids"

# Default stack — UsersPanel removed (auth + user admin live in Morph).
# Deprecated: userspanel-api / userspanel-admin still startable via alias `userspanel`.
ALL_SERVICES=(
  morph-api
  formx-api
  composerx-api
  booki-api
  morph-engi-api
  bk-api
  academi-api
  sharpreport-api
  morph-ui
  morph-utils-ui
  bk-ui
  formx-ui
  composerx-ui
  booki-ui
  morph-engi-ui
  academi-ui
  sharpreport-ui
)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
DIM='\033[2m'
NC='\033[0m'

mkdir -p "$LOG_DIR"

log()  { echo -e "${CYAN}▶${NC} $*"; }
ok()   { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}!${NC} $*"; }
err()  { echo -e "${RED}✗${NC} $*" >&2; }

# Load KEY=VALUE lines from .env into the current shell (CRLF-safe; no bash `source`).
apply_dotenv() {
  local envfile="$1"
  [[ -f "$envfile" ]] || return 0
  local line key value
  while IFS= read -r line || [[ -n "$line" ]]; do
    line="${line%$'\r'}"
    [[ -z "$line" || "$line" =~ ^[[:space:]]*# ]] && continue
    if [[ "$line" =~ ^([A-Za-z_][A-Za-z0-9_]*)=(.*)$ ]]; then
      key="${BASH_REMATCH[1]}"
      value="${BASH_REMATCH[2]}"
      value="${value%$'\r'}"
      # Strip optional surrounding quotes
      if [[ "$value" =~ ^\"(.*)\"$ ]]; then
        value="${BASH_REMATCH[1]}"
      elif [[ "$value" =~ ^\'(.*)\'$ ]]; then
        value="${BASH_REMATCH[1]}"
      fi
      printf -v "$key" '%s' "$value"
      export "$key"
    fi
  done <"$envfile"
}

# Parent .env first, then app dir (app overrides parent — e.g. booki/.env over backend/).
load_app_env() {
  local workdir="$1"
  local parent
  parent="$(dirname "$workdir")"
  if [[ -f "${parent}/.env" ]]; then
    apply_dotenv "${parent}/.env"
  fi
  if [[ -f "${workdir}/.env" ]]; then
    apply_dotenv "${workdir}/.env"
  fi
}

ensure_morph_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building morph-api (macOS — avoids BadgerDB LC_UUID issue)..."
    (cd "${ROOT}/morph" && go build -o "${RUN_DIR}/morph-server" main.go)
  fi
}

ensure_composerx_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building composerx-api..."
    (cd "${ROOT}/composerx/backend" && go build -o "${RUN_DIR}/composerx-server" .)
  fi
}

ensure_formx_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building formx-api..."
    (cd "${ROOT}/formx/backend" && go build -o "${RUN_DIR}/formx-server" ./cmd/server)
  fi
}

pid_of() {
  local name="$1"
  [[ -f "$PID_FILE" ]] || return 1
  local line pid
  line="$(grep -E "^${name}:" "$PID_FILE" 2>/dev/null | tail -1 || true)"
  [[ -n "$line" ]] || return 1
  pid="${line##*:}"
  if kill -0 "$pid" 2>/dev/null; then
    echo "$pid"
    return 0
  fi
  return 1
}

remove_pid_entry() {
  local name="$1"
  [[ -f "$PID_FILE" ]] || return 0
  grep -v "^${name}:" "$PID_FILE" > "${PID_FILE}.tmp" 2>/dev/null || true
  mv "${PID_FILE}.tmp" "$PID_FILE"
  [[ -s "$PID_FILE" ]] || rm -f "$PID_FILE"
}

record_pid() {
  local name="$1"
  local pid="$2"
  touch "$PID_FILE"
  remove_pid_entry "$name"
  echo "${name}:${pid}" >>"$PID_FILE"
}

kill_pid() {
  local pid="$1"
  kill "$pid" 2>/dev/null || true
  sleep 0.5
  kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null || true
}

# Kill any process listening on a TCP port (orphaned dev servers not tracked in pids).
free_listening_port() {
  local port="$1"
  local pids
  pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  [[ -n "$pids" ]] || return 0
  warn "Freeing port ${port} (stale listener pid(s): ${pids//$'\n'/ })"
  # shellcheck disable=SC2086
  kill $pids 2>/dev/null || true
  sleep 0.5
  pids="$(lsof -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
  [[ -z "$pids" ]] || kill -9 $pids 2>/dev/null || true
}

neo4j_port_listening() {
  lsof -tiTCP:7687 -sTCP:LISTEN >/dev/null 2>&1
}

ensure_neo4j() {
  if neo4j_port_listening; then
    ok "Neo4j already listening on port 7687"
    return 0
  fi
  if ! command -v neo4j >/dev/null 2>&1; then
    warn "Neo4j not installed (port 7687 closed). Morph graph features may be unavailable."
    return 0
  fi
  log "Starting Neo4j (bolt port 7687)..."
  if neo4j start >/dev/null 2>&1; then
    sleep 2
    if neo4j_port_listening; then
      ok "Neo4j started"
    else
      warn "Neo4j start did not open port 7687 yet; graph features may be unavailable."
    fi
  else
    warn "Failed to start Neo4j; graph features may be unavailable."
  fi
}

bk_python() {
  if [[ -x "${ROOT}/bk/.venv/bin/python" ]]; then
    echo "${ROOT}/bk/.venv/bin/python"
  elif command -v python3.12 >/dev/null 2>&1; then
    echo python3.12
  elif command -v python3.11 >/dev/null 2>&1; then
    echo python3.11
  elif command -v python3 >/dev/null 2>&1; then
    echo python3
  else
    echo python
  fi
}

bk_venv_python() {
  if command -v python3.12 >/dev/null 2>&1; then
    echo python3.12
  elif command -v python3.11 >/dev/null 2>&1; then
    echo python3.11
  elif command -v python3 >/dev/null 2>&1; then
    echo python3
  else
    echo python
  fi
}

ensure_bk_venv() {
  local venv_dir="${ROOT}/bk/.venv"
  local venv_py="${venv_dir}/bin/python"
  local venv_pip="${venv_dir}/bin/pip"
  local bootstrap_py
  bootstrap_py="$(bk_venv_python)"
  if ! command -v "$bootstrap_py" >/dev/null 2>&1; then
    warn "python3 not found — bk-api needs Python 3.11+ and pip install -r bk/requirements.txt"
    return 1
  fi
  if [[ -x "$venv_py" ]] && ! "$venv_py" -c "import uvicorn" >/dev/null 2>&1; then
    if [[ ! -x "$venv_pip" ]] || ! "$venv_pip" --version >/dev/null 2>&1; then
      warn "bk .venv is broken (e.g. after moving the repo) — recreating…"
      rm -rf "$venv_dir"
    elif ! "$venv_py" -c 'import sys; assert sys.version_info[:2] in {(3,11),(3,12),(3,13)}' 2>/dev/null; then
      warn "bk .venv uses Python $($venv_py -c 'import sys; print(f\"{sys.version_info.major}.{sys.version_info.minor}\")' 2>/dev/null || echo unknown) — recreating with $bootstrap_py…"
      rm -rf "$venv_dir"
    fi
  fi
  if [[ ! -x "$venv_py" ]]; then
    log "Creating bk Python venv (.venv) with ${bootstrap_py}..."
    "$bootstrap_py" -m venv "$venv_dir"
  fi
  if ! "$venv_py" -c "import uvicorn" >/dev/null 2>&1; then
    log "Installing bk Python requirements (first run may take a minute)..."
    if ! "$venv_pip" install -q -r "${ROOT}/bk/requirements.txt"; then
      err "bk pip install failed — try: rm -rf bk/.venv && ./start-all.sh restart bk-api"
      return 1
    fi
    ok "bk Python environment ready"
  fi
}

start_service() {
  local name="$1"
  local workdir="$2"
  shift 2
  local logfile="${LOG_DIR}/${name}.log"
  local existing_pid
  if existing_pid="$(pid_of "$name")"; then
    warn "${name} already running (pid ${existing_pid})"
    return 0
  fi
  remove_pid_entry "$name"
  log "Starting ${name} → ${logfile}"
  (
    cd "$workdir"
    if [[ "$name" == "sharpreport-ui" ]]; then
      # SharpReport/.env is for the Rust backend (PORT, DATABASE_URL, JVM opts, etc.).
      if [[ -f "${workdir}/.env" ]]; then
        apply_dotenv "${workdir}/.env"
      fi
    else
      load_app_env "$workdir"
    fi
    exec "$@"
  ) >>"$logfile" 2>&1 &
  disown 2>/dev/null || true
  record_pid "$name" "$!"
  ok "${name} (pid $(pid_of "$name"))"
}

stop_service() {
  local name="$1"
  local pid
  if ! pid="$(pid_of "$name")"; then
    warn "${name} is not running"
    remove_pid_entry "$name"
    return 0
  fi
  log "Stopping ${name} (pid ${pid})"
  kill_pid "$pid"
  remove_pid_entry "$name"
  ok "${name} stopped"
}

start_one() {
  local name="$1"
  case "$name" in
    userspanel-api)
      start_service userspanel-api "${ROOT}/UsersPanel/backend" cargo run
      ;;
    morph-api)
      ensure_morph_binary
      if [[ "$(uname -s)" == "Darwin" ]]; then
        start_service morph-api "${ROOT}/morph" "${RUN_DIR}/morph-server"
      else
        start_service morph-api "${ROOT}/morph" go run main.go
      fi
      ;;
    formx-api)
      load_app_env "${ROOT}/formx/backend"
      free_listening_port "${SERVER_PORT:-29909}"
      if [[ "$(uname -s)" == "Darwin" ]]; then
        ensure_formx_binary
        start_service formx-api "${ROOT}/formx/backend" "${RUN_DIR}/formx-server"
      else
        start_service formx-api "${ROOT}/formx/backend" go run ./cmd/server
      fi
      ;;
    composerx-api)
      if [[ "$(uname -s)" == "Darwin" ]]; then
        ensure_composerx_binary
        start_service composerx-api "${ROOT}/composerx/backend" "${RUN_DIR}/composerx-server"
      else
        start_service composerx-api "${ROOT}/composerx/backend" go run .
      fi
      ;;
    booki-api)
      start_service booki-api "${ROOT}/booki/backend" go run ./cmd/server
      ;;
    morph-engi-api)
      start_service morph-engi-api "${ROOT}/morph-engi/backend" cargo run
      ;;
    bk-api)
      ensure_bk_venv || { err "bk-api: Python environment not ready"; return 1; }
      start_service bk-api "${ROOT}/bk" "$(bk_python)" main.py
      ;;
    academi-api)
      start_service academi-api "${ROOT}/academi/backend" go run ./cmd/main.go
      ;;
    sharpreport-api)
      start_service sharpreport-api "${ROOT}/SharpReport/backend" cargo run
      ;;
    userspanel-admin)
      start_service userspanel-admin "${ROOT}/UsersPanel/admin" npm run dev
      ;;
    morph-ui)
      start_service morph-ui "${ROOT}/morph/frontend" npm start
      ;;
    morph-utils-ui)
      start_service morph-utils-ui "${ROOT}/morph-utils/frontend" npm run dev
      ;;
    bk-ui)
      start_service bk-ui "${ROOT}/bk/frontend" npm start
      ;;
    formx-ui)
      start_service formx-ui "${ROOT}/formx/frontend" npm run dev
      ;;
    composerx-ui)
      start_service composerx-ui "${ROOT}/composerx/frontend" npm run dev
      ;;
    booki-ui)
      start_service booki-ui "${ROOT}/booki/frontend" npm run dev
      ;;
    morph-engi-ui)
      start_service morph-engi-ui "${ROOT}/morph-engi/frontend" npm run dev
      ;;
    academi-ui)
      start_service academi-ui "${ROOT}/academi/web" python3 -m http.server 8765 --bind 127.0.0.1
      ;;
    sharpreport-ui)
      start_service sharpreport-ui "${ROOT}/SharpReport/frontend" npm run dev
      ;;
    *)
      err "Unknown service: ${name}"
      return 1
      ;;
  esac
}

restart_one() {
  local name="$1"
  stop_service "$name"
  start_one "$name"
}

resolve_services() {
  local input="$1"
  case "$input" in
    all|"")
      echo "${ALL_SERVICES[*]}"
      ;;
    userspanel|users-panel)
      echo "userspanel-api userspanel-admin"
      ;;
    morph)
      echo "morph-api morph-ui"
      ;;
    morph-utils|utils)
      echo "morph-utils-ui"
      ;;
    bk|ground-control)
      echo "bk-api bk-ui"
      ;;
    formx|formsx)
      echo "formx-api formx-ui"
      ;;
    composerx|tranmail)
      echo "composerx-api composerx-ui"
      ;;
    booki)
      echo "booki-api booki-ui"
      ;;
    morph-engi|engi)
      echo "morph-engi-api morph-engi-ui"
      ;;
    academi)
      echo "academi-api academi-ui"
      ;;
    sharpreport|datapulse)
      echo "sharpreport-api sharpreport-ui"
      ;;
    *)
      echo "$input"
      ;;
  esac
}

start_services_list() {
  local name
  touch "$PID_FILE"
  while IFS= read -r name; do
    [[ -z "$name" ]] && continue
    start_one "$name"
  done
}

restart_all() {
  stop_all
  ensure_neo4j
  log "Starting all robo apps…"
  printf '%s\n' "${ALL_SERVICES[@]}" | start_services_list
  echo ""
  ok "All apps restarted."
  print_status
}

expand_services() {
  local arg name expanded=()
  for arg in "$@"; do
    # shellcheck disable=SC2207
    expanded+=($(resolve_services "$arg"))
  done
  printf '%s\n' "${expanded[@]}" | awk '!seen[$0]++'
}

stop_all() {
  if [[ ! -f "$PID_FILE" ]]; then
    warn "No PID file at ${PID_FILE}; nothing to stop."
    return 0
  fi
  log "Stopping all robo dev processes..."
  local line name pid
  while IFS= read -r line; do
    [[ -z "$line" ]] && continue
    name="${line%%:*}"
    pid="${line##*:}"
    if kill -0 "$pid" 2>/dev/null; then
      echo "  stopping ${name} (pid ${pid})"
      kill_pid "$pid"
    fi
  done < "$PID_FILE"
  rm -f "$PID_FILE"
  ok "All stopped."
}

service_url() {
  case "$1" in
    userspanel-api)     echo "http://127.0.0.1:5001/swagger-ui" ;;
    userspanel-admin)   echo "http://localhost:5173" ;;
    morph-api)          echo "http://localhost:9090" ;;
    morph-ui)           echo "http://localhost:3031" ;;
    morph-utils-ui)     echo "http://localhost:3040" ;;
    bk-api)             echo "http://localhost:8000/docs" ;;
    bk-ui)              echo "http://localhost:3000" ;;
    formx-api)          echo "http://localhost:29909/swagger/index.html" ;;
    formx-ui)           echo "http://localhost:19909" ;;
    composerx-api)      echo "http://localhost:8043/health" ;;
    composerx-ui)       echo "http://localhost:8044" ;;
    booki-api)          echo "http://127.0.0.1:9095/health" ;;
    booki-ui)           echo "http://localhost:5174" ;;
    morph-engi-api)     echo "http://127.0.0.1:9096/health" ;;
    morph-engi-ui)      echo "http://localhost:5179" ;;
    academi-api)        echo "http://127.0.0.1:8978/health" ;;
    academi-ui)         echo "http://localhost:8765" ;;
    sharpreport-api)    echo "http://127.0.0.1:3050" ;;
    sharpreport-ui)     echo "http://localhost:5178" ;;
    *)                  echo "" ;;
  esac
}

print_status() {
  local name pid url state
  for name in "${ALL_SERVICES[@]}"; do
    url="$(service_url "$name")"
    if pid="$(pid_of "$name")"; then
      state="${GREEN}running${NC} (pid ${pid})"
    else
      state="${DIM}stopped${NC}"
      remove_pid_entry "$name"
    fi
    if [[ -n "$url" ]]; then
      echo -e "  ${name}: ${state}  ${DIM}${url}${NC}"
    else
      echo -e "  ${name}: ${state}"
    fi
  done
}

print_list() {
  cat <<'EOF'
Services (use with start | stop | restart | logs):

  userspanel-api       UsersPanel Rust API
  userspanel-admin     UsersPanel Svelte admin UI
  morph-api            Morph / MorphData backend
  morph-ui             Morph React frontend
  morph-utils-ui       Morph Utils shell (FormsX + ComposerX + DataX)
  bk-api               AI tools API (Ground Control / Python FastAPI)
  bk-ui                AI tools UI (Ground Control / React)
  formx-api            FormsX backend
  formx-ui             FormsX frontend
  composerx-api        ComposerX (TranMail) backend
  composerx-ui         ComposerX frontend
  booki-api            Booki backend
  booki-ui             Booki frontend
  morph-engi-api       Morph Engi civil engineering API (Rust)
  morph-engi-ui        Morph Engi frontend (Svelte)
  academi-api          Academi study assistant API (Go)
  academi-ui           Academi web frontend (static)
  sharpreport-api      SharpReport / DataPulse backend
  sharpreport-ui       SharpReport frontend

Aliases (API + UI together):

  userspanel, morph, morph-utils, bk, formx, composerx, booki, morph-engi, academi, sharpreport
  all                  every service above (same as start/stop/restart with no args)

Examples:

  ./start-all.sh restart              # stop + start everything
  ./start-all.sh stop                 # stop everything
  ./start-all.sh start                # start everything
  ./start-all.sh restart morph
  ./start-all.sh restart formx-api
  ./start-all.sh logs composerx-ui
  ./start-all.sh status
EOF
}

do_install() {
  log "Installing dependencies..."
  (cd "${ROOT}/UsersPanel/admin" && npm install)
  (cd "${ROOT}/UsersPanel/backend" && cargo fetch)
  (cd "${ROOT}/morph/frontend" && npm install)
  (cd "${ROOT}/morph-utils/frontend" && npm install)
  (cd "${ROOT}/bk/frontend" && npm install)
  ensure_bk_venv || warn "bk-api Python deps not installed — run: python3 -m venv bk/.venv && bk/.venv/bin/pip install -r bk/requirements.txt"
  (cd "${ROOT}/morph" && go mod download)
  (cd "${ROOT}/formx/frontend" && npm install)
  (cd "${ROOT}/formx/backend" && go mod download)
  (cd "${ROOT}/composerx/frontend" && npm install)
  (cd "${ROOT}/composerx/backend" && go mod download)
  (cd "${ROOT}/booki/frontend" && npm install)
  (cd "${ROOT}/booki/backend" && go mod download)
  (cd "${ROOT}/morph-engi/frontend" && npm install)
  (cd "${ROOT}/morph-engi/backend" && cargo fetch)
  (cd "${ROOT}/academi/backend" && go mod download)
  (cd "${ROOT}/SharpReport/frontend" && npm install)
  (cd "${ROOT}/SharpReport/backend" && cargo fetch)
  ok "Dependencies ready"
}

start_all() {
  if [[ -f "$PID_FILE" ]]; then
    err "PID file already exists. Run ./start-all.sh --stop first."
    exit 1
  fi
  ensure_neo4j
  touch "$PID_FILE"

  for name in "${ALL_SERVICES[@]}"; do
    start_one "$name"
  done

  echo ""
  ok "All apps started (academi excluded)."
  echo ""
  print_status
  echo ""
  echo "Logs: ${LOG_DIR}/"
  echo "Per-app: ./start-all.sh restart <service>"
  echo "Stop all: ./start-all.sh --stop   (or Ctrl+C)"
  echo ""
  warn "Press Ctrl+C to stop all services."
}

usage() {
  sed -n '3,18p' "$0" | sed 's/^# \{0,1\}//'
}

# --- main ---

CMD="${1:-all}"
shift || true

case "$CMD" in
  -h|--help|help)
    usage
    exit 0
    ;;
  list)
    print_list
    exit 0
    ;;
  status)
    print_status
    exit 0
    ;;
  logs)
    [[ -n "${1:-}" ]] || { err "Usage: ./start-all.sh logs <service>"; exit 1; }
    name="$(expand_services "$1" | head -1)"
    logfile="${LOG_DIR}/${name}.log"
    [[ -f "$logfile" ]] || { err "No log yet: ${logfile}"; exit 1; }
    exec tail -f "$logfile"
    ;;
  stop|--stop)
    if [[ -z "${1:-}" || "${1:-}" == "all" ]]; then
      stop_all
    else
      while IFS= read -r name; do
        stop_service "$name"
      done < <(expand_services "$@")
    fi
    exit 0
    ;;
  start)
    if [[ -z "${1:-}" || "${1:-}" == "all" ]]; then
      ensure_neo4j
      log "Starting all robo apps…"
      printf '%s\n' "${ALL_SERVICES[@]}" | start_services_list
      echo ""
      print_status
    else
      while IFS= read -r name; do
        start_one "$name"
      done < <(expand_services "$@")
    fi
    exit 0
    ;;
  restart)
    if [[ -z "${1:-}" || "${1:-}" == "all" ]]; then
      restart_all
    else
      while IFS= read -r name; do
        restart_one "$name"
      done < <(expand_services "$@")
    fi
    exit 0
    ;;
  --install)
    do_install
    start_all
    cleanup_trap() { echo ""; stop_all; exit 0; }
    # INT/TERM only — not EXIT. disown'd jobs make `wait` return immediately,
    # which would fire EXIT and kill everything right after startup.
    trap cleanup_trap INT TERM
    while true; do sleep 86400; done
    ;;
  all|"")
    start_all
    cleanup_trap() { echo ""; stop_all; exit 0; }
    trap cleanup_trap INT TERM
    while true; do sleep 86400; done
    ;;
  *)
    err "Unknown command: ${CMD}"
    echo ""
    usage
    exit 1
    ;;
esac
