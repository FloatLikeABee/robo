#!/usr/bin/env bash
# Proves start-all.sh stop/restart kills the recorded process group, not only
# the parent, and that port cleanup cannot group-kill an unrecorded listener.
# No repo services are started. Safe for a later CI job.
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
UNDER_TEST="${START_ALL_SCRIPT:-${ROOT}/start-all.sh}"
# shellcheck disable=SC1090
source "$UNDER_TEST"

tmpdir="$(mktemp -d)"
RUN_DIR="${tmpdir}/run"
LOG_DIR="${RUN_DIR}/logs"
PID_FILE="${RUN_DIR}/pids"
mkdir -p "$LOG_DIR"
extra_pids="${tmpdir}/extra-pids"
: >"$extra_pids"

port="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')"
START_ALL_PORT_OVERRIDE_NAME=orphan-demo
START_ALL_PORT_OVERRIDE="$port"

fixture="${tmpdir}/forklisten.py"
marker="${tmpdir}/children"
cat >"$fixture" <<'PY'
import os, socket, sys, time
port = int(sys.argv[1])
marker = sys.argv[2]
listener = os.fork()
if listener == 0:
    s = socket.socket()
    s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    s.bind(("127.0.0.1", port))
    s.listen(1)
    while True:
        time.sleep(3600)
sleeper = os.fork()
if sleeper == 0:
    while True:
        time.sleep(3600)
with open(marker, "w") as handle:
    handle.write("%s\n%s\n" % (listener, sleeper))
while True:
    try:
        os.wait()
    except OSError:
        time.sleep(3600)
PY

foreign="${tmpdir}/foreign.py"
foreign_marker="${tmpdir}/foreign-child"
cat >"$foreign" <<'PY'
import os, socket, sys, time
port = int(sys.argv[1])
marker = sys.argv[2]
# Grandchild stays in this process group after the intermediate exits,
# but it is not a descendant of the listener. Group-kill would take it.
# Tree-kill of the listener would not.
child = os.fork()
if child == 0:
    grandchild = os.fork()
    if grandchild > 0:
        os._exit(0)
    with open(marker, "w") as handle:
        handle.write(str(os.getpid()))
    while True:
        time.sleep(3600)
os.waitpid(child, 0)
s = socket.socket()
s.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
s.bind(("127.0.0.1", port))
s.listen(1)
while True:
    time.sleep(3600)
PY

note_pid() {
  echo "$1" >>"$extra_pids"
}

cleanup() {
  set +e
  watch_sleeper=""
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
  if [[ -f "$extra_pids" ]]; then
    while IFS= read -r extra || [[ -n "$extra" ]]; do
      [[ -z "$extra" ]] && continue
      kill -9 "$extra" >/dev/null 2>&1
    done <"$extra_pids"
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

# Count real port-cleanup entries (a listener was still up). A call with no
# listener, which start_service always makes, does not count.
port_cleanup_entered=0
watch_sleeper=""
eval "$(declare -f free_listening_port | sed '1s/free_listening_port/free_listening_port_real/')"
free_listening_port() {
  local port="$1"
  local listeners=""
  listeners="$(listening_pids "$port" || true)"
  if [[ -n "$listeners" ]]; then
    port_cleanup_entered=1
    if [[ -n "${watch_sleeper:-}" ]] && kill -0 "$watch_sleeper" 2>/dev/null; then
      fail "non-listening child ${watch_sleeper} still alive when port cleanup started on ${port}"
    fi
  fi
  free_listening_port_real "$@"
}

eval "$(declare -f pgid_of | sed '1s/pgid_of/pgid_of_real/')"
pgid_of() {
  if [[ "${PRETEND_UNKNOWN_SELF:-}" == 1 && "$1" == "$$" ]]; then
    return 0
  fi
  pgid_of_real "$1"
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

wait_file() {
  local path="$1"
  local i=0
  while [[ "$i" -lt 30 ]]; do
    if [[ -s "$path" ]]; then
      return 0
    fi
    sleep 0.1
    i=$((i + 1))
  done
  fail "timed out waiting for ${path}"
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

assert_alive() {
  local pid="$1"
  local label="$2"
  if ! kill -0 "$pid" 2>/dev/null; then
    fail "${label} pid ${pid} is dead"
  fi
}

read_fixture_children() {
  wait_file "$marker"
  listener="$(sed -n '1p' "$marker")"
  sleeper="$(sed -n '2p' "$marker")"
  [[ -n "$listener" && -n "$sleeper" ]] || fail "fixture marker missing pids"
  [[ "$listener" != "$sleeper" ]] || fail "fixture listener and sleeper are the same pid"
  note_pid "$listener"
  note_pid "$sleeper"
}

# The test's start_one launches the fixture. restart_one is the real function.
start_one() {
  start_service orphan-demo "$tmpdir" python3 "$fixture" "$port" "$marker"
}

echo "case 1: restart_one kills the non-listening child before port cleanup"
rm -f "$marker"
start_one
wait_listen
assert_one_listener
parent="$(pid_of orphan-demo)"
pgid="$(pgid_of "$parent")"
read_fixture_children
live_listener="$(listening_pids "$port" | awk 'NR==1 { print $1; exit }')"
[[ "$parent" == "$pgid" ]] || fail "restart setup: recorded pid ${parent} is not the group leader (pgid ${pgid})"
[[ "$live_listener" == "$listener" ]] || fail "restart setup: listener ${live_listener} is not the forked child ${listener}"
[[ "$sleeper" != "$parent" ]] || fail "restart setup: sleeper is the parent"
if listening_pids "$port" | grep -qx "$sleeper"; then
  fail "restart setup: non-listening child ${sleeper} is listening"
fi
echo "restart_one setup: parent ${parent} pgid ${pgid} listener ${listener} sleeper ${sleeper}"
port_cleanup_entered=0
watch_sleeper="$sleeper"
restart_one orphan-demo
watch_sleeper=""
[[ "$port_cleanup_entered" == "0" ]] || fail "port cleanup ran during restart_one stop (group kill did not clear the listener first)"
assert_dead "$sleeper" "restart_one non-listening child before port cleanup"
assert_dead "$parent" "restart_one parent"
assert_dead "$listener" "restart_one listener"
wait_listen
assert_one_listener
new_parent="$(pid_of orphan-demo)"
[[ -n "$new_parent" && "$new_parent" != "$parent" ]] || fail "restart_one did not record a new pid (old ${parent} new ${new_parent:-})"
stop_service orphan-demo
assert_port_free

echo "case 1b: kill_recorded_pid reaps the non-listening child before any port cleanup"
rm -f "$marker"
start_one
wait_listen
parent="$(pid_of orphan-demo)"
read_fixture_children
port_cleanup_entered=0
watch_sleeper="$sleeper"
kill_recorded_pid "$parent"
[[ "$port_cleanup_entered" == "0" ]] || fail "port cleanup ran inside kill_recorded_pid"
assert_dead "$sleeper" "non-listening child before port cleanup"
assert_dead "$listener" "listener before port cleanup"
assert_dead "$parent" "parent before port cleanup"
watch_sleeper=""
remove_pid_entry orphan-demo
assert_port_free

echo "case 3: killing only the parent leaves the listener; port cleanup tree-kills it and leaves the sibling"
set +m
rm -f "$marker"
python3 "$fixture" "$port" "$marker" >/dev/null 2>&1 &
legacy_parent=$!
note_pid "$legacy_parent"
wait_listen
wait_file "$marker"
legacy_listener="$(sed -n '1p' "$marker")"
legacy_sleeper="$(sed -n '2p' "$marker")"
note_pid "$legacy_listener"
note_pid "$legacy_sleeper"
legacy_pgid="$(pgid_of "$legacy_parent")"
self_pgid="$(pgid_of "$$")"
echo "legacy parent ${legacy_parent} pgid ${legacy_pgid} listener ${legacy_listener} sleeper ${legacy_sleeper} launcher pgid ${self_pgid}"
kill "$legacy_parent" 2>/dev/null || true
sleep 0.3
if [[ -z "$(listening_pids "$port")" ]]; then
  fail "legacy kill of parent also dropped the listener; orphan setup did not reproduce"
fi
assert_alive "$legacy_sleeper" "legacy sleeper before port cleanup"
watch_sleeper=""
free_listening_port "$port"
wait_until_port_free "$port" || fail "port ${port} still busy after free_listening_port"
assert_port_free
assert_dead "$legacy_listener" "legacy listener"
assert_alive "$legacy_sleeper" "legacy non-listening sibling after tree-kill of listener"
assert_alive "$$" "launcher pid was killed while freeing an orphan"
kill -9 "$legacy_sleeper" >/dev/null 2>&1 || true

echo "case 6: unrecorded group leader is tree-killed; a same-group non-descendant survives"
rm -f "$foreign_marker"
assert_port_free
set -m
python3 "$foreign" "$port" "$foreign_marker" >/dev/null 2>&1 &
foreign_leader=$!
set +m
note_pid "$foreign_leader"
wait_listen
wait_file "$foreign_marker"
foreign_child="$(cat "$foreign_marker")"
note_pid "$foreign_child"
foreign_pgid="$(pgid_of "$foreign_leader")"
child_pgid="$(pgid_of "$foreign_child")"
[[ "$foreign_leader" == "$foreign_pgid" ]] || fail "foreign leader ${foreign_leader} pgid ${foreign_pgid} is not its own group"
[[ "$child_pgid" == "$foreign_pgid" ]] || fail "foreign child ${foreign_child} pgid ${child_pgid} left the leader group ${foreign_pgid}"
if pid_is_recorded "$foreign_leader"; then
  fail "foreign leader was recorded"
fi
echo "foreign leader ${foreign_leader} pgid ${foreign_pgid} same-group child ${foreign_child}"
free_out="$(free_listening_port "$port" 2>&1)"
printf '%s\n' "$free_out" | grep -q "was not started by this launcher" || fail "missing warning for unrecorded listener: ${free_out}"
assert_port_free
assert_dead "$foreign_leader" "foreign listener"
assert_alive "$foreign_child" "same-group process that is not a descendant of the unrecorded listener"
kill -9 "$foreign_child" >/dev/null 2>&1 || true

echo "case 7: PID 0, 1, and empty do not signal; unknown launcher PGID does not group-kill"
set +m
sleep 120 &
victim=$!
note_pid "$victim"
printf 'bad:0\n' >"$PID_FILE"
got="$(pid_of bad || true)"
[[ -z "$got" ]] || fail "pid_of accepted 0 (got ${got})"
printf 'bad:1\n' >"$PID_FILE"
got="$(pid_of bad || true)"
[[ -z "$got" ]] || fail "pid_of accepted 1 (got ${got})"
printf 'bad:\n' >"$PID_FILE"
got="$(pid_of bad || true)"
[[ -z "$got" ]] || fail "pid_of accepted an empty pid (got ${got})"
kill_recorded_pid 0
kill_recorded_pid 1
kill_recorded_pid ""
kill_tree 0 TERM
kill_tree 1 TERM
kill_tree "" TERM
assert_alive "$victim" "same-group process after refusing PID 0/1/empty"
kill -9 "$victim" >/dev/null 2>&1 || true
wait "$victim" 2>/dev/null || true

printf 'morph-api:%s\nbk-ui:%s\n' "$$" "$$" >"$PID_FILE"
got="$(pid_of '.*' || true)"
[[ -z "$got" ]] || fail "regex name .* selected pid ${got}"
got="$(pid_of 'm.*' || true)"
[[ -z "$got" ]] || fail "regex name m.* selected pid ${got}"
got="$(pid_of morph-api)"
[[ "$got" == "$$" ]] || fail "exact morph-api lookup got ${got}"
remove_pid_entry '.*'
grep -q '^morph-api:' "$PID_FILE" || fail "remove of .* deleted morph-api"
grep -q '^bk-ui:' "$PID_FILE" || fail "remove of .* deleted bk-ui"
remove_pid_entry morph-api
if grep -q '^morph-api:' "$PID_FILE"; then
  fail "exact remove left morph-api"
fi
grep -q '^bk-ui:' "$PID_FILE" || fail "exact remove deleted bk-ui"
rm -f "$PID_FILE"

rm -f "$foreign_marker"
assert_port_free
set -m
python3 "$foreign" "$port" "$foreign_marker" >/dev/null 2>&1 &
unknown_leader=$!
set +m
note_pid "$unknown_leader"
wait_listen
wait_file "$foreign_marker"
unknown_child="$(cat "$foreign_marker")"
note_pid "$unknown_child"
record_pid orphan-demo "$unknown_leader"
PRETEND_UNKNOWN_SELF=1
kill_recorded_pid "$unknown_leader"
PRETEND_UNKNOWN_SELF=0
assert_dead "$unknown_leader" "recorded leader under unknown launcher pgid"
assert_alive "$unknown_child" "same-group non-descendant after fail-closed stop"
assert_alive "$$" "launcher after fail-closed stop"
kill -9 "$unknown_child" >/dev/null 2>&1 || true
remove_pid_entry orphan-demo
# The listener is dead; the port is free. Drop a leftover if the tree kill missed it.
if [[ -n "$(listening_pids "$port")" ]]; then
  fail "unknown-pgid stop left a listener on ${port}"
fi

echo "case 8: missing lsof warns once on stderr and does not invent PIDs"
saved_path="$PATH"
nolsof="${tmpdir}/nolsof"
mkdir -p "$nolsof"
old_ifs="$IFS"
IFS=':'
for dir in $PATH; do
  [[ -d "$dir" ]] || continue
  for cmd in "$dir"/*; do
    [[ -e "$cmd" ]] || continue
    base="${cmd##*/}"
    [[ "$base" == "lsof" ]] && continue
    [[ -e "${nolsof}/${base}" ]] && continue
    ln -s "$cmd" "${nolsof}/${base}" 2>/dev/null || true
  done
done
IFS="$old_ifs"
PATH="$nolsof"
LSOF_MISSING_WARNED=0
lsof_err="${tmpdir}/lsof.err"
listening_pids 9 >/dev/null 2>"$lsof_err"
grep -q "lsof is not installed" "$lsof_err" || fail "missing lsof warning: $(cat "$lsof_err")"
: >"$lsof_err"
listening_pids 9 >/dev/null 2>"$lsof_err"
if [[ -s "$lsof_err" ]]; then
  fail "lsof warning repeated: $(cat "$lsof_err")"
fi
PATH="$saved_path"
LSOF_MISSING_WARNED=0

echo "case 9: morph-api restart waits for /health"
health_port="$(python3 -c 'import socket; s=socket.socket(); s.bind(("127.0.0.1", 0)); print(s.getsockname()[1]); s.close()')"
saved_override_name="$START_ALL_PORT_OVERRIDE_NAME"
saved_override_port="$START_ALL_PORT_OVERRIDE"
START_ALL_PORT_OVERRIDE_NAME=morph-api
START_ALL_PORT_OVERRIDE="$health_port"
START_ALL_HEALTH_ATTEMPTS=20
START_ALL_HEALTH_INTERVAL=0.2
cat >"${tmpdir}/health.py" <<'PY'
from http.server import BaseHTTPRequestHandler, HTTPServer
import sys
class Handler(BaseHTTPRequestHandler):
    def do_GET(self):
        if self.path.split("?", 1)[0] != "/health":
            self.send_response(404)
            self.end_headers()
            return
        body = b'{"status":"healthy"}'
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(body)))
        self.end_headers()
        self.wfile.write(body)
    def log_message(self, fmt, *args):
        return
HTTPServer(("127.0.0.1", int(sys.argv[1])), Handler).serve_forever()
PY
python3 "${tmpdir}/health.py" "$health_port" >/dev/null 2>&1 &
health_pid=$!
note_pid "$health_pid"
i=0
while [[ "$i" -lt 30 ]]; do
  if curl -fsS --max-time 1 "http://127.0.0.1:${health_port}/health" >/dev/null 2>&1; then
    break
  fi
  sleep 0.1
  i=$((i + 1))
done
curl -fsS --max-time 1 "http://127.0.0.1:${health_port}/health" >/dev/null 2>&1 || fail "health fixture did not serve /health"
record_pid morph-api "$health_pid"
wait_morph_api_healthy || fail "healthy morph-api fixture was rejected"
kill -9 "$health_pid" >/dev/null 2>&1 || true
wait "$health_pid" 2>/dev/null || true
sleep 30 &
dummy_pid=$!
note_pid "$dummy_pid"
record_pid morph-api "$dummy_pid"
START_ALL_HEALTH_ATTEMPTS=3
START_ALL_HEALTH_INTERVAL=0.2
if wait_morph_api_healthy >/dev/null 2>&1; then
  fail "unhealthy morph-api was treated as healthy"
fi
kill -9 "$dummy_pid" >/dev/null 2>&1 || true
wait "$dummy_pid" 2>/dev/null || true
remove_pid_entry morph-api
START_ALL_PORT_OVERRIDE_NAME="$saved_override_name"
START_ALL_PORT_OVERRIDE="$saved_override_port"

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
rm -f "$marker"
start_service orphan-demo "$tmpdir" /usr/bin/python3 "$fixture" "$port" "$marker" || fail "stub PATH: start_service returned an error"
PATH="$saved_path"
wait_listen
assert_one_listener
parent="$(pid_of orphan-demo)"
pgid="$(pgid_of "$parent")"
read_fixture_children
[[ "$parent" == "$pgid" ]] || fail "stub PATH: pid ${parent} is not group leader ${pgid}"
[[ "$listener" != "$parent" ]] || fail "stub PATH: listener is the parent"
if [[ -f "${LOG_DIR}/orphan-demo.log" ]] && grep -q "stub-interpreter" "${LOG_DIR}/orphan-demo.log"; then
  fail "stub PATH: launcher exec'd a stub interpreter"
fi
echo "stub PATH: parent ${parent} pgid ${pgid} listener ${listener} sleeper ${sleeper}"
watch_sleeper="$sleeper"
stop_service orphan-demo
watch_sleeper=""
assert_port_free
assert_dead "$parent" "stub PATH parent"
assert_dead "$listener" "stub PATH listener"
assert_dead "$sleeper" "stub PATH sleeper"

echo "case 4: launcher help, exact names, and status still run"
help_text="$("${ROOT}/start-all.sh" --help)"
printf '%s\n' "$help_text" | grep -q "process group" || fail "help does not document process-group stop"
printf '%s\n' "$help_text" | grep -q "port is free" || fail "help does not document the port wait"
printf '%s\n' "$help_text" | grep -q "PID 0" || fail "help does not document the PID 0 refusal"
printf '%s\n' "$help_text" | grep -q "/health" || fail "help does not document the morph-api health wait"
printf '%s\n' "$help_text" | grep -q "lsof" || fail "help does not document lsof"
real_pids="${ROOT}/.robo-dev/pids"
pid_before=""
if [[ -f "$real_pids" ]]; then
  pid_before="$(cat "$real_pids")"
fi
set +e
stop_out="$("${ROOT}/start-all.sh" stop '.*' 2>&1)"
stop_rc=$?
set -e
[[ "$stop_rc" != 0 ]] || fail "stop .* succeeded"
printf '%s\n' "$stop_out" | grep -q "Unknown service" || fail "stop .* was not rejected: ${stop_out}"
if [[ -f "$real_pids" || -n "$pid_before" ]]; then
  pid_after=""
  if [[ -f "$real_pids" ]]; then
    pid_after="$(cat "$real_pids")"
  fi
  [[ "$pid_before" == "$pid_after" ]] || fail "stop .* changed ${real_pids}"
fi
"${ROOT}/start-all.sh" status >/dev/null
"${ROOT}/start-all.sh" list >/dev/null

echo "OK: process-group stop reaps non-listening children before port cleanup"
