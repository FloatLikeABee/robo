#!/usr/bin/env bash
# Dev launcher for robo platform apps.
# Folders: morph, morph-utils, formx, composerx, morph-engi, SharpReport, bk.
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
# Stop and restart signal the process group this launcher recorded for that
# service (TERM, then KILL). Bash job control creates that group, so macOS
# bash 3.2 does not need a setsid binary. Children from `go run`, cargo, and
# npm are not left listening or holding the Badger lock. The next start waits
# until that service's port is free. PID 0, PID 1, and an empty PID are never
# signaled. A listener this launcher did not record is stopped as that
# process tree only, never as a process group. morph-api start and restart
# wait until /health succeeds. Service names match exactly (not as a regex).
# Port checks use lsof and warn if it is missing.
#
# Aliases (API + UI): morph, morph-utils (shell + Data Access), bk, formx, composerx,
#   morph-engi, sharpreport — or `all` for every service below.
# Neo4j: full-stack start/restart ensures bolt port 7687 is up when `neo4j` CLI exists.
# Auth: hosted by Morph (the standalone UsersPanel project has been removed).
#
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
RUN_DIR="${ROOT}/.robo-dev"
LOG_DIR="${RUN_DIR}/logs"
PID_FILE="${RUN_DIR}/pids"
# Warn once per process when port checks cannot see listeners.
LSOF_MISSING_WARNED=0

# Default stack — auth + user admin live in Morph (UsersPanel project removed).
ALL_SERVICES=(
  morph-api
  formx-api
  composerx-api
  morph-engi-api
  bk-api
  sharpreport-api
  morph-ui
  morph-utils-ui
  invite-signup-ui
  bk-ui
  formx-ui
  composerx-ui
  morph-engi-ui
  sharpreport-ui
)

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
DIM='\033[2m'
NC='\033[0m'

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

# Single local config: repo-root .env only (nested app .env files are ignored).
load_root_env() {
  apply_dotenv "${ROOT}/.env"
}

ensure_morph_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building morph-api (macOS — avoids BadgerDB LC_UUID issue)..."
    ensure_run_dirs
    (cd "${ROOT}/morph" && go build -o "${RUN_DIR}/morph-server" main.go)
  fi
}

ensure_composerx_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building composerx-api..."
    ensure_run_dirs
    (cd "${ROOT}/composerx/backend" && go build -o "${RUN_DIR}/composerx-server" .)
  fi
}

ensure_formx_binary() {
  if [[ "$(uname -s)" == "Darwin" ]]; then
    log "Building formx-api..."
    ensure_run_dirs
    (cd "${ROOT}/formx/backend" && go build -o "${RUN_DIR}/formx-server" ./cmd/server)
  fi
}

# A PID we may signal. Rejects empty, non-numeric, 0, and 1.
# kill -0 0 is true (it checks the caller's process group). child_pids 0
# matches PPID 0 and returns PID 1. kill -TERM 0 signals that group.
valid_target_pid() {
  local pid="$1"
  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  [[ "$pid" -gt 1 ]] || return 1
}

known_service() {
  local name="$1"
  local candidate
  for candidate in "${ALL_SERVICES[@]}"; do
    [[ "$candidate" == "$name" ]] && return 0
  done
  return 1
}

pid_of() {
  local name="$1"
  [[ -f "$PID_FILE" ]] || return 1
  local line pid
  # Exact service name. grep -E would treat `.*` as a regex and select a real row.
  line="$(awk -F: -v n="$name" '$1 == n { found=$0 } END { if (found != "") print found }' "$PID_FILE" 2>/dev/null || true)"
  [[ -n "$line" ]] || return 1
  pid="${line#*:}"
  pid="${pid%%:*}"
  valid_target_pid "$pid" || return 1
  if kill -0 "$pid" 2>/dev/null; then
    echo "$pid"
    return 0
  fi
  return 1
}

remove_pid_entry() {
  local name="$1"
  [[ -f "$PID_FILE" ]] || return 0
  awk -F: -v n="$name" '$1 != n { print }' "$PID_FILE" > "${PID_FILE}.tmp" 2>/dev/null || true
  mv "${PID_FILE}.tmp" "$PID_FILE"
  [[ -s "$PID_FILE" ]] || rm -f "$PID_FILE"
}

record_pid() {
  local name="$1"
  local pid="$2"
  if ! valid_target_pid "$pid"; then
    err "refusing to record unsafe pid '${pid}' for ${name}"
    return 1
  fi
  ensure_run_dirs
  touch "$PID_FILE"
  remove_pid_entry "$name"
  echo "${name}:${pid}" >>"$PID_FILE"
}

# True when this launcher wrote the PID into the pid file.
pid_is_recorded() {
  local pid="$1"
  [[ -f "$PID_FILE" ]] || return 1
  valid_target_pid "$pid" || return 1
  awk -F: -v p="$pid" '$2 == p { found=1 } END { exit (found ? 0 : 1) }' "$PID_FILE"
}

ensure_run_dirs() {
  mkdir -p "$LOG_DIR"
}

# PGID of a live process, or empty. BSD and GNU ps both accept this form.
pgid_of() {
  local pid="$1"
  ps -p "$pid" -o pgid= 2>/dev/null | awk 'NR==1 { print $1; exit }' || true
}

child_pids() {
  local pid="$1"
  # Empty and 0 both become 0 in awk arithmetic, which lists PID 1.
  valid_target_pid "$pid" || return 0
  ps -A -o pid= -o ppid= 2>/dev/null | awk -v p="$pid" '$2+0 == p+0 && $1+0 > 1 { print $1 }' || true
}

group_alive() {
  local pgid="$1"
  valid_target_pid "$pgid" || return 1
  ps -A -o pgid= 2>/dev/null | awk -v g="$pgid" '$1+0 == g+0 { found=1; exit } END { exit (found ? 0 : 1) }'
}

# Negative pid means "this process group" on bash 3.2 and on BSD/GNU kill.
signal_group() {
  local sig="$1"
  local pgid="$2"
  valid_target_pid "$pgid" || return 0
  kill -"$sig" -"$pgid" 2>/dev/null || true
}

wait_group_exit() {
  local pgid="$1"
  local i=0
  while [[ "$i" -lt 25 ]]; do
    if ! group_alive "$pgid"; then
      return 0
    fi
    sleep 0.2
    i=$((i + 1))
  done
  return 1
}

# Kill a dedicated process group. Refuses the launcher's own group so a
# legacy shared group cannot take down every other service.
kill_process_group() {
  local pgid="$1"
  local self
  valid_target_pid "$pgid" || return 0
  self="$(pgid_of "$$")"
  # Unknown launcher group: fail closed. An empty self used to compare unequal
  # and signal the target group anyway.
  if [[ -z "$self" || "$pgid" == "$self" ]]; then
    return 1
  fi
  signal_group TERM "$pgid"
  if ! wait_group_exit "$pgid"; then
    signal_group KILL "$pgid"
    wait_group_exit "$pgid" || true
  fi
}

# Descendants first, so a child is signaled before its parent can exit and
# reparent it out of the tree (the `go run` orphan case).
kill_tree() {
  local pid="$1"
  local sig="$2"
  local child
  valid_target_pid "$pid" || return 0
  for child in $(child_pids "$pid"); do
    kill_tree "$child" "$sig"
  done
  kill -"$sig" "$pid" 2>/dev/null || true
}

# TERM the descendant tree, then KILL if the PID is still alive.
kill_pid_tree() {
  local pid="$1"
  local i=0
  valid_target_pid "$pid" || return 0
  kill_tree "$pid" TERM
  while [[ "$i" -lt 25 ]]; do
    if ! kill -0 "$pid" 2>/dev/null; then
      return 0
    fi
    sleep 0.2
    i=$((i + 1))
  done
  kill_tree "$pid" KILL
}

# Stop one recorded process. If it leads its own process group (how
# start_service launches), signal the whole group. Otherwise kill only its
# descendant tree — never the launcher group shared by older runs.
# Group-kill only when this launcher's own PGID is known.
kill_recorded_pid() {
  local pid="$1"
  local pgid self
  valid_target_pid "$pid" || return 0
  if ! kill -0 "$pid" 2>/dev/null; then
    return 0
  fi
  pgid="$(pgid_of "$pid")"
  self="$(pgid_of "$$")"
  if [[ -n "$self" && -n "$pgid" && "$pgid" == "$pid" && "$pgid" != "$self" ]]; then
    kill_process_group "$pgid"
    return 0
  fi
  kill_pid_tree "$pid"
}

listening_pids() {
  local port="$1"
  [[ -n "$port" ]] || return 0
  if ! command -v lsof >/dev/null 2>&1; then
    if [[ "$LSOF_MISSING_WARNED" != 1 ]]; then
      # stderr: callers capture stdout as a PID list.
      echo -e "${YELLOW}!${NC} lsof is not installed; port checks cannot see listeners. Stop cannot prove a port is free, and a leftover process may keep the port or the Badger lock. Install lsof (macOS: already in the system; Debian/Ubuntu: apt install lsof)." >&2
      LSOF_MISSING_WARNED=1
    fi
    return 0
  fi
  lsof -nP -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true
}

wait_until_port_free() {
  local port="$1"
  local i=0
  while [[ "$i" -lt 25 ]]; do
    if [[ -z "$(listening_pids "$port")" ]]; then
      return 0
    fi
    sleep 0.2
    i=$((i + 1))
  done
  return 1
}

# Clear a leftover listener. Group-kill only a PID this launcher recorded.
# Anything else is one process tree: group-killing an unrelated leader on
# :3000 or :8000 would take down that application's other processes.
free_listening_port() {
  local port="$1"
  local pids pid
  pids="$(listening_pids "$port")"
  [[ -n "$pids" ]] || return 0
  warn "Freeing port ${port} (stale listener pid(s): ${pids//$'\n'/ })"
  for pid in $pids; do
    if pid_is_recorded "$pid"; then
      kill_recorded_pid "$pid"
    else
      warn "port ${port}: listener ${pid} was not started by this launcher; stopping that process tree only, not its process group"
      kill_pid_tree "$pid"
    fi
  done
  if ! wait_until_port_free "$port"; then
    pids="$(listening_pids "$port")"
    for pid in $pids; do
      valid_target_pid "$pid" || continue
      if pid_is_recorded "$pid"; then
        kill_recorded_pid "$pid"
      else
        kill_tree "$pid" KILL
      fi
    done
    wait_until_port_free "$port" || true
  fi
}

# True when the command exits 0 within about two seconds. Rejects a macOS
# python stub that `command -v` can see but that does not run.
interpreter_runs() {
  local bin="$1"
  shift
  command -v "$bin" >/dev/null 2>&1 || return 1
  "$bin" "$@" >/dev/null 2>&1 &
  local probe=$!
  local i=0
  while [[ "$i" -lt 20 ]]; do
    if ! kill -0 "$probe" 2>/dev/null; then
      wait "$probe"
      return $?
    fi
    sleep 0.1
    i=$((i + 1))
  done
  kill -9 "$probe" 2>/dev/null || true
  wait "$probe" 2>/dev/null || true
  return 1
}

# New session without forking, so the PID we record stays the group leader.
# Used only when bash job control could not be enabled. Perl before python:
# macOS ships perl, and python3 may be a stub.
exec_in_new_session() {
  if interpreter_runs perl -MPOSIX -e 'exit 0'; then
    exec perl -MPOSIX -e 'eval { POSIX::setsid(); }; exec @ARGV or die "exec: $!"' -- "$@"
  fi
  if interpreter_runs python3 -c 'import os'; then
    exec python3 -c 'import os,sys
try:
    os.setsid()
except OSError:
    pass
os.execvp(sys.argv[1], sys.argv[1:])' "$@"
  fi
  if interpreter_runs python -c 'import os'; then
    exec python -c 'import os,sys
try:
    os.setsid()
except OSError:
    pass
os.execvp(sys.argv[1], sys.argv[1:])' "$@"
  fi
  if interpreter_runs setsid true; then
    exec setsid "$@"
  fi
  exec "$@"
}

# TCP port the service binds, if it has one. Override hooks are for scripts/test-start-all-process-group.sh.
service_port() {
  local name="$1"
  if [[ -n "${START_ALL_PORT_OVERRIDE_NAME:-}" && "$name" == "$START_ALL_PORT_OVERRIDE_NAME" && -n "${START_ALL_PORT_OVERRIDE:-}" ]]; then
    printf '%s\n' "$START_ALL_PORT_OVERRIDE"
    return 0
  fi
  case "$name" in
    morph-api)          printf '%s\n' "${PORT:-9090}" ;;
    morph-ui)           printf '%s\n' 3031 ;;
    morph-utils-ui)     printf '%s\n' 3040 ;;
    invite-signup-ui)   printf '%s\n' 3051 ;;
    bk-api)             printf '%s\n' 8000 ;;
    bk-ui)              printf '%s\n' 3000 ;;
    formx-api)          printf '%s\n' "${SERVER_PORT:-29909}" ;;
    formx-ui)           printf '%s\n' 19909 ;;
    composerx-api)      printf '%s\n' 8043 ;;
    composerx-ui)       printf '%s\n' 8044 ;;
    morph-engi-api)     printf '%s\n' 9096 ;;
    morph-engi-ui)      printf '%s\n' 5179 ;;
    sharpreport-api)    printf '%s\n' "${SHARPREPORT_PORT:-3050}" ;;
    sharpreport-ui)     printf '%s\n' 5178 ;;
    *)                  return 0 ;;
  esac
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
  local existing_pid port use_job_control=0 pid
  ensure_run_dirs
  if existing_pid="$(pid_of "$name")"; then
    warn "${name} already running (pid ${existing_pid})"
    return 0
  fi
  remove_pid_entry "$name"
  port="$(service_port "$name")"
  if [[ -n "$port" ]]; then
    free_listening_port "$port"
    if ! wait_until_port_free "$port"; then
      err "${name}: port ${port} is still in use"
      return 1
    fi
  fi
  log "Starting ${name} → ${logfile}"
  # Bash job control (including bash 3.2) makes this PID its own process
  # group, so stop can signal `go run` / npm children with the parent.
  # No setsid binary required. Helpers run only if monitor mode fails.
  if set -m 2>/dev/null; then
    use_job_control=1
  fi
  (
    trap '' HUP
    cd "$workdir"
    load_root_env
    if [[ "$use_job_control" == 1 ]]; then
      exec "$@"
    fi
    exec_in_new_session "$@"
  ) >>"$logfile" 2>&1 &
  pid=$!
  disown 2>/dev/null || true
  if [[ "$use_job_control" == 1 ]]; then
    set +m
  fi
  record_pid "$name" "$pid"
  ok "${name} (pid ${pid})"
}

stop_service() {
  local name="$1"
  local pid port did=0
  if pid="$(pid_of "$name")"; then
    log "Stopping ${name} (pid ${pid})"
    kill_recorded_pid "$pid"
    did=1
  else
    warn "${name} is not running"
  fi
  remove_pid_entry "$name"
  port="$(service_port "$name")"
  if [[ -n "$port" && -n "$(listening_pids "$port")" ]]; then
    free_listening_port "$port"
    did=1
  fi
  if [[ -n "$port" ]] && ! wait_until_port_free "$port"; then
    err "${name}: port ${port} still in use after stop"
    return 1
  fi
  if [[ "$did" == 1 ]]; then
    ok "${name} stopped"
  fi
}

# morph-api restart is not done until GET /health succeeds.
wait_morph_api_healthy() {
  local port url i=0 attempts interval
  port="$(service_port morph-api)"
  url="http://127.0.0.1:${port}/health"
  attempts="${START_ALL_HEALTH_ATTEMPTS:-90}"
  interval="${START_ALL_HEALTH_INTERVAL:-1}"
  if ! command -v curl >/dev/null 2>&1; then
    err "curl is not installed; cannot confirm morph-api is healthy at ${url}"
    return 1
  fi
  while [[ "$i" -lt "$attempts" ]]; do
    if curl -fsS --max-time 2 "$url" >/dev/null 2>&1; then
      ok "morph-api healthy (${url})"
      return 0
    fi
    if ! pid_of morph-api >/dev/null; then
      err "morph-api exited before ${url} returned success"
      return 1
    fi
    sleep "$interval"
    i=$((i + 1))
  done
  err "morph-api did not become healthy at ${url}"
  return 1
}

start_one() {
  local name="$1"
  case "$name" in
    morph-api)
      ensure_morph_binary
      if [[ "$(uname -s)" == "Darwin" ]]; then
        start_service morph-api "${ROOT}/morph" "${RUN_DIR}/morph-server"
      else
        start_service morph-api "${ROOT}/morph" go run main.go
      fi
      wait_morph_api_healthy
      ;;
    formx-api)
      load_root_env
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
    morph-engi-api)
      start_service morph-engi-api "${ROOT}/morph-engi/backend" cargo run
      ;;
    bk-api)
      ensure_bk_venv || { err "bk-api: Python environment not ready"; return 1; }
      start_service bk-api "${ROOT}/bk" "$(bk_python)" main.py
      ;;
    sharpreport-api)
      start_service sharpreport-api "${ROOT}/SharpReport/backend" cargo run
      ;;
    morph-ui)
      start_service morph-ui "${ROOT}/morph/frontend" npm start
      ;;
    morph-utils-ui)
      start_service morph-utils-ui "${ROOT}/morph-utils/frontend" npm run dev
      ;;
    invite-signup-ui)
      start_service invite-signup-ui "${ROOT}/invite-signup/frontend" npm run dev
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
    morph-engi-ui)
      start_service morph-engi-ui "${ROOT}/morph-engi/frontend" npm run dev
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
    morph)
      echo "morph-api morph-ui"
      ;;
    morph-utils|utils)
      echo "morph-utils-ui sharpreport-api sharpreport-ui"
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
    morph-engi|engi)
      echo "morph-engi-api morph-engi-ui"
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
  ensure_run_dirs
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
  local line name names=() n
  while IFS= read -r line || [[ -n "$line" ]]; do
    [[ -z "$line" ]] && continue
    name="${line%%:*}"
    names+=("$name")
  done < "$PID_FILE"
  for n in "${names[@]+"${names[@]}"}"; do
    stop_service "$n"
  done
  rm -f "$PID_FILE"
  ok "All stopped."
}

service_url() {
  case "$1" in
    morph-api)          echo "http://localhost:${PORT:-9090}" ;;
    morph-ui)           echo "http://localhost:3031" ;;
    morph-utils-ui)     echo "http://localhost:3040" ;;
    invite-signup-ui)   echo "http://localhost:3051" ;;
    bk-api)             echo "http://localhost:8000/docs" ;;
    bk-ui)              echo "http://localhost:3000" ;;
    formx-api)          echo "http://localhost:29909/swagger/index.html" ;;
    formx-ui)           echo "http://localhost:19909" ;;
    composerx-api)      echo "http://localhost:8043/health" ;;
    composerx-ui)       echo "http://localhost:8044" ;;
    morph-engi-api)     echo "http://127.0.0.1:9096/health" ;;
    morph-engi-ui)      echo "http://localhost:5179" ;;
    sharpreport-api)    echo "http://127.0.0.1:${SHARPREPORT_PORT:-3050}" ;;
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

  morph-api            Morph / MorphData backend
  morph-ui             Morph React frontend
  morph-utils-ui       MorphUtils shell (Event Logs, Content Maker, Data Access, Project)
  invite-signup-ui     Invite Signup (admin codes + user redeem)
  bk-api               AI tools API
  bk-ui                AI tools UI
  formx-api            Event Logs backend
  formx-ui             Event Logs frontend
  composerx-api        Content Maker backend
  composerx-ui         Content Maker frontend
  morph-engi-api       Project API (Rust)
  morph-engi-ui        Project frontend (Svelte)
  sharpreport-api      Data Access backend
  sharpreport-ui       Data Access frontend

Aliases (API + UI together):

  morph, morph-utils (shell + Data Access API/UI), bk, formx, composerx, morph-engi, sharpreport
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
  (cd "${ROOT}/morph/frontend" && npm install)
  (cd "${ROOT}/morph-utils/frontend" && npm install)
  (cd "${ROOT}/invite-signup/frontend" && npm install)
  (cd "${ROOT}/bk/frontend" && npm install)
  ensure_bk_venv || warn "bk-api Python deps not installed — run: python3 -m venv bk/.venv && bk/.venv/bin/pip install -r bk/requirements.txt"
  (cd "${ROOT}/morph" && go mod download)
  (cd "${ROOT}/formx/frontend" && npm install)
  (cd "${ROOT}/formx/backend" && go mod download)
  (cd "${ROOT}/composerx/frontend" && npm install)
  (cd "${ROOT}/composerx/backend" && go mod download)
  (cd "${ROOT}/morph-engi/frontend" && npm install)
  (cd "${ROOT}/morph-engi/backend" && cargo fetch)
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
  ensure_run_dirs
  touch "$PID_FILE"

  for name in "${ALL_SERVICES[@]}"; do
    start_one "$name"
  done

  echo ""
  ok "All apps started."
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
  # Header comments up to, but not including, the first non-comment line.
  sed -n '3,/^[^#]/p' "$0" | sed '$d; s/^# \{0,1\}//'
}

# --- main ---

if [[ "${BASH_SOURCE[0]}" != "$0" ]]; then
  return 0 2>/dev/null || exit 0
fi

load_root_env

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
    if ! known_service "$name"; then
      err "Unknown service: ${name}"
      exit 1
    fi
    logfile="${LOG_DIR}/${name}.log"
    [[ -f "$logfile" ]] || { err "No log yet: ${logfile}"; exit 1; }
    exec tail -f "$logfile"
    ;;
  stop|--stop)
    if [[ -z "${1:-}" || "${1:-}" == "all" ]]; then
      stop_all
    else
      while IFS= read -r name; do
        [[ -z "$name" ]] && continue
        if ! known_service "$name"; then
          err "Unknown service: ${name}"
          exit 1
        fi
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
        [[ -z "$name" ]] && continue
        if ! known_service "$name"; then
          err "Unknown service: ${name}"
          exit 1
        fi
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
        [[ -z "$name" ]] && continue
        if ! known_service "$name"; then
          err "Unknown service: ${name}"
          exit 1
        fi
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
