#!/usr/bin/env bash
# Fields2Cover v2 container entrypoint.
#
# Starts the C++ gRPC shim, waits for its Unix socket to appear, then
# starts the Go API. On any child exit (or container SIGTERM), kills both
# children and exits with the dead child's status — Docker / k8s decide
# whether to restart the whole thing.
#
# This is deliberately bash, not s6/dumb-init/supervisord. Two processes,
# linear startup, single failure mode — anything more is overkill.

set -euo pipefail

SOCKET="${F2C_GRPC_SOCKET:-/tmp/f2c.sock}"
API_PORT="${API_PORT:-8080}"
SHIM_BIN="${SHIM_BIN:-/usr/local/bin/f2c-grpc-server}"
API_BIN="${API_BIN:-/usr/local/bin/api}"
WAIT_TIMEOUT_S="${SHIM_WAIT_TIMEOUT_S:-15}"

log() { printf '[entrypoint] %s\n' "$*" >&2; }

# Clean any stale socket from a crashed previous run (matters on volume mounts).
rm -f "$SOCKET"

log "starting f2c-grpc-server on ${SOCKET}"
"$SHIM_BIN" &
SHIM_PID=$!

# Wait for the shim to create the socket.
for _ in $(seq 1 "$((WAIT_TIMEOUT_S * 10))"); do
    if [[ -S "$SOCKET" ]]; then
        break
    fi
    if ! kill -0 "$SHIM_PID" 2>/dev/null; then
        log "shim exited before opening ${SOCKET}"
        wait "$SHIM_PID" || true
        exit 1
    fi
    sleep 0.1
done

if [[ ! -S "$SOCKET" ]]; then
    log "timed out waiting for ${SOCKET} after ${WAIT_TIMEOUT_S}s"
    kill "$SHIM_PID" 2>/dev/null || true
    exit 1
fi

log "shim ready (pid ${SHIM_PID}); starting api on :${API_PORT}"
"$API_BIN" &
API_PID=$!

shutdown() {
    local sig=$1
    log "received ${sig}, stopping children"
    kill -TERM "$API_PID" 2>/dev/null || true
    kill -TERM "$SHIM_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
    wait "$SHIM_PID" 2>/dev/null || true
    exit 0
}
trap 'shutdown SIGTERM' TERM
trap 'shutdown SIGINT' INT

# Wait for either child to exit. `wait -n` returns the first one.
set +e
wait -n
EXITED=$?
set -e

# Figure out which one died and reap the other.
if kill -0 "$API_PID" 2>/dev/null; then
    log "shim exited with status ${EXITED}; stopping api"
    kill -TERM "$API_PID" 2>/dev/null || true
    wait "$API_PID" 2>/dev/null || true
else
    log "api exited with status ${EXITED}; stopping shim"
    kill -TERM "$SHIM_PID" 2>/dev/null || true
    wait "$SHIM_PID" 2>/dev/null || true
fi

exit "$EXITED"
