#!/usr/bin/env bash
# Proves start-all.sh stop/restart kills descendant listeners, not only the
# recorded parent. Mirrors `go run`: a parent forks a child that binds a port.
# No repo services are started. Safe for a later CI job.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
# shellcheck disable=SC1091
source "${ROOT}/start-all.sh"

tmpdir="$(mktemp -d)"
RUN_DIR="${tmpdir}/run"
LOG_DIR="${RUN_DIR}/logs"
PID_FILE="${RUN_DIR}/pids"
mkdir -p "$LOG_DIR"

port="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')"
START_ALL_PORT_OVERRIDE_NAME=orphan-demo
START_ALL_PORT_OVERRIDE="$port"

fixture="${tmpdir}/forklisten.py"
cat >"$fixture" <<'PY'
import os, socket, sys, time
port = int(sys.argv[1])
pid = os.fork()
if pid == 0:
    s = socket.socket()
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("127.0.0.1", port))
    s.listen(1)
    while True:
        time.sleep(3600)
else:
    os.wait()
PY

cleanup() {
  set +e
  if declare -F stop_service >/dev/null 2>&1; then
    stop_service orphan-demo >/dev/null 2>&1
  fi
  if [[ -n "${port:-}" ]]; then
    local_pids="$(lsof -nP -tiTCP:"${port}" -sTCP:LISTEN 2>/dev/null || true)"
    if [[ -n "$local_pids" ]]; then
      # shellcheck disable=SC2086
      kill -9 $local_pids >/dev/null 2>&1
    fi
  fi
  rm -rf "$tmpdir"
}
trap cleanup EXIT

fail() {
  echo "FAIL: $*" >&2
  if [[ -f "${LOG_DIR}/orphan-demo.log" ]]; then
    echo "--- orphan-demo.log ---" >&2
    cat "${LOG_DIR}/orphan-demo.log" >&2 || true
  fi
  exit 1
}

wait_listen() {
  local i=0
  while [[ "$i" -lt 30 ]]; do
    if [[ -n "$(listening_pids "$port")" ]]; then
      return 0
    fi
    sleep 0.1
    i=$((i + 1))
  done
  fail "timed out waiting for port ${port}"
}

assert_one_listener() {
  local n
  n="$(listening_pids "$port" | wc -l | tr -d ' ')"
  [[ "$n" == "1" ]] || fail "expected 1 listener on ${port}, got ${n}: $(listening_pids "$port" | tr '\n' ' ')"
}

assert_port_free() {
  local n
  n="$(listening_pids "$port" | wc -l | tr -d ' ')"
  [[ "$n" == "0" ]] || fail "expected port ${port} free, still: $(listening_pids "$port" | tr '\n' ' ')"
}

assert_dead() {
  local pid="$1"
  local label="$2"
  if kill -0 "$pid" 2>/dev/null; then
    fail "${label} pid ${pid} still alive"
  fi
}

# Stop must reap the forked listener, which is not the PID recorded by start_service.
cycle_stop() {
  local label="$1"
  local parent listener pgid
  start_service orphan-demo "$tmpdir" python3 "$fixture" "$port"
  wait_listen
  assert_one_listener
  parent="$(pid_of orphan-demo)"
  pgid="$(pgid_of "$parent")"
  listener="$(listening_pids "$port" | awk 'NR==1 { print $1; exit }')"
  [[ -n "$parent" && -n "$listener" ]] || fail "${label}: missing parent or listener"
  [[ "$parent" == "$pgid" ]] || fail "${label}: recorded pid ${parent} is not the process group leader (pgid ${pgid})"
  [[ "$listener" != "$parent" ]] || fail "${label}: listener is the parent; fixture did not fork"
  echo "${label}: parent ${parent} pgid ${pgid} listener ${listener}"
  stop_service orphan-demo
  assert_port_free
  assert_dead "$parent" "${label} parent"
  assert_dead "$listener" "${label} listener"
}

echo "case 1: job-control process group — three restarts"
i=1
while [[ "$i" -le 3 ]]; do
  cycle_stop "session restart ${i}"
  i=$((i + 1))
done

echo "case 2: job control is the default even when a setsid helper exists"
cycle_stop "job-control"

echo "case 3: killing only the parent leaves the listener; port cleanup reaps it"
set +m
python3 "$fixture" "$port" >/dev/null 2>&1 &
legacy_parent=$!
wait_listen
legacy_listener="$(listening_pids "$port" | awk 'NR==1 { print $1; exit }')"
legacy_pgid="$(pgid_of "$legacy_parent")"
self_pgid="$(pgid_of "$$")"
echo "legacy parent ${legacy_parent} pgid ${legacy_pgid} listener ${legacy_listener} launcher pgid ${self_pgid}"
kill "$legacy_parent" 2>/dev/null || true
sleep 0.3
if [[ -z "$(listening_pids "$port")" ]]; then
  fail "legacy kill of parent also dropped the listener; orphan setup did not reproduce"
fi
# Same-group orphans must not take the launcher down with them.
free_listening_port "$port"
wait_until_port_free "$port" || fail "port ${port} still busy after free_listening_port"
assert_port_free
assert_dead "$legacy_listener" "legacy listener"
kill -0 "$$" 2>/dev/null || fail "launcher pid was killed while freeing an orphan"

echo "case 5: stub python/perl/setsid on PATH must not be exec'd"
stubdir="${tmpdir}/stubs"
mkdir -p "$stubdir"
for stub in python3 python perl setsid; do
  cat >"${stubdir}/${stub}" <<'EOF'
#!/bin/sh
echo "stub-interpreter" >&2
exit 1
EOF
  chmod +x "${stubdir}/${stub}"
done
saved_path="$PATH"
PATH="${stubdir}:${PATH}"
start_service orphan-demo "$tmpdir" /usr/bin/python3 "$fixture" "$port" || fail "stub PATH: start_service returned an error"
PATH="$saved_path"
wait_listen
assert_one_listener
parent="$(pid_of orphan-demo)"
pgid="$(pgid_of "$parent")"
listener="$(listening_pids "$port" | awk 'NR==1 { print $1; exit }')"
[[ "$parent" == "$pgid" ]] || fail "stub PATH: pid ${parent} is not group leader ${pgid}"
[[ "$listener" != "$parent" ]] || fail "stub PATH: listener is the parent"
# The setsid wrapper must not have been the stub.
if [[ -f "${LOG_DIR}/orphan-demo.log" ]] && grep -q "stub-interpreter" "${LOG_DIR}/orphan-demo.log"; then
  fail "stub PATH: launcher exec'd a stub interpreter"
fi
echo "stub PATH: parent ${parent} pgid ${pgid} listener ${listener}"
stop_service orphan-demo
assert_port_free
assert_dead "$parent" "stub PATH parent"
assert_dead "$listener" "stub PATH listener"

echo "case 4: launcher help and status still run"
help_text="$("${ROOT}/start-all.sh" --help)"
printf '%s\n' "$help_text" | grep -q "process group" || fail "help does not document process-group stop"
printf '%s\n' "$help_text" | grep -q "port is free" || fail "help does not document the port wait"
"${ROOT}/start-all.sh" status >/dev/null
"${ROOT}/start-all.sh" list >/dev/null

echo "OK: process-group stop reaps forked listeners"
