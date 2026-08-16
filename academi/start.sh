#!/usr/bin/env bash
#
# Academi dev launcher — starts the Go backend and the static web frontend
# together, and shuts both down cleanly on Ctrl+C.
#
# Usage:
#   ./start.sh                 # default ports (frontend 8765, backend from .env)
#   FRONTEND_PORT=3000 ./start.sh
#
set -euo pipefail

# Resolve repo root (directory containing this script) so it works from anywhere.
ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BACKEND_DIR="$ROOT_DIR/backend"
WEB_DIR="$ROOT_DIR/web"

FRONTEND_PORT="${FRONTEND_PORT:-8765}"

# Colors (fall back to no-op if not a TTY).
if [ -t 1 ]; then
    BOLD="\033[1m"; GREEN="\033[32m"; BLUE="\033[34m"; YELLOW="\033[33m"; RESET="\033[0m"
else
    BOLD=""; GREEN=""; BLUE=""; YELLOW=""; RESET=""
fi

log() { printf "%b\n" "${BLUE}[academi]${RESET} $*"; }
err() { printf "%b\n" "${YELLOW}[academi]${RESET} $*" >&2; }

# --- Preflight checks --------------------------------------------------------
command -v go >/dev/null 2>&1 || { err "Go is not installed or not on PATH. Install from https://go.dev/dl/"; exit 1; }

PY_BIN=""
if command -v python3 >/dev/null 2>&1; then
    PY_BIN="python3"
elif command -v python >/dev/null 2>&1; then
    PY_BIN="python"
else
    err "python3 is required to serve the web frontend (ES modules need an HTTP server)."
    exit 1
fi

# Ensure backend/.env exists (copy from the example on first run).
if [ ! -f "$BACKEND_DIR/.env" ]; then
    if [ -f "$BACKEND_DIR/.env.example" ]; then
        cp "$BACKEND_DIR/.env.example" "$BACKEND_DIR/.env"
        log "Created backend/.env from .env.example — add your AI API key before using chat."
    else
        err "backend/.env.example not found; cannot create backend/.env."
        exit 1
    fi
fi

# Read SERVER_PORT from backend/.env (default 8978) so URLs we print are accurate.
BACKEND_PORT="$(grep -E '^[[:space:]]*SERVER_PORT[[:space:]]*=' "$BACKEND_DIR/.env" 2>/dev/null | tail -n1 | cut -d'=' -f2 | tr -d '[:space:]')"
BACKEND_PORT="${BACKEND_PORT:-8978}"

# --- Process management -------------------------------------------------------
BACKEND_PID=""
FRONTEND_PID=""

_cleaned=0
cleanup() {
    [ "$_cleaned" = "1" ] && return
    _cleaned=1
    log "Shutting down…"
    for pid in "$FRONTEND_PID" "$BACKEND_PID"; do
        [ -n "$pid" ] && kill -TERM "$pid" 2>/dev/null || true
    done
    # Give them a moment to exit gracefully, then force-kill stragglers.
    for _ in 1 2 3 4 5 6; do
        still=0
        for pid in "$FRONTEND_PID" "$BACKEND_PID"; do
            [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null && still=1
        done
        [ "$still" = "0" ] && break
        sleep 0.5
    done
    for pid in "$FRONTEND_PID" "$BACKEND_PID"; do
        [ -n "$pid" ] && kill -0 "$pid" 2>/dev/null && kill -9 "$pid" 2>/dev/null || true
    done
    wait 2>/dev/null || true
}
trap cleanup EXIT INT TERM

# --- Start backend -----------------------------------------------------------
# Build to a binary first, then run it directly. `go run` forks a separate
# compiled child, so killing it would orphan the real server and leak the port.
BACKEND_BIN="$BACKEND_DIR/bin/academi-backend"
log "Building backend…"
( cd "$BACKEND_DIR" && go build -o "$BACKEND_BIN" ./cmd/main.go )

log "Starting backend on port ${BOLD}${BACKEND_PORT}${RESET}…"
(
    cd "$BACKEND_DIR"
    exec "$BACKEND_BIN"
) &
BACKEND_PID=$!

# --- Start frontend ----------------------------------------------------------
log "Starting frontend (static) on port ${BOLD}${FRONTEND_PORT}${RESET}…"
(
    cd "$WEB_DIR"
    exec "$PY_BIN" -m http.server "$FRONTEND_PORT" --bind 0.0.0.0
) &
FRONTEND_PID=$!

# --- Ready banner ------------------------------------------------------------
sleep 1
printf "%b\n" ""
printf "%b\n" "${GREEN}${BOLD}Academi is running.${RESET}"
printf "%b\n" "  Frontend : ${BOLD}http://localhost:${FRONTEND_PORT}${RESET}"
printf "%b\n" "  Backend  : ${BOLD}http://localhost:${BACKEND_PORT}${RESET}/health"
printf "%b\n" "  ${YELLOW}On a phone/LAN device, open http://<this-machine-ip>:${FRONTEND_PORT}${RESET}"
printf "%b\n" "  Press ${BOLD}Ctrl+C${RESET} to stop both."
printf "%b\n" ""

# Wait on whichever child exits first; if one dies, tear everything down.
# (Portable poll — macOS ships bash 3.2, which lacks `wait -n`.)
while kill -0 "$BACKEND_PID" 2>/dev/null && kill -0 "$FRONTEND_PID" 2>/dev/null; do
    sleep 1
done
err "A process exited — stopping the other."
