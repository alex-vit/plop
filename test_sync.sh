#!/usr/bin/env bash
set -euo pipefail

# End-to-end sync test for gosync.
# Runs two instances on localhost, creates a file on A, waits for it on B.

GOSYNC="./gosync"
HOME_A="/tmp/gosync-test-a"
HOME_B="/tmp/gosync-test-b"
SYNC_A="/tmp/gosync-test-sync-a"
SYNC_B="/tmp/gosync-test-sync-b"
TIMEOUT=90

cleanup() {
    echo "Cleaning up..."
    [ -n "${PID_A:-}" ] && kill "$PID_A" 2>/dev/null || true
    [ -n "${PID_B:-}" ] && kill "$PID_B" 2>/dev/null || true
    wait "$PID_A" 2>/dev/null || true
    wait "$PID_B" 2>/dev/null || true
    rm -rf "$HOME_A" "$HOME_B" "$SYNC_A" "$SYNC_B"
}
trap cleanup EXIT

# Build
echo "=== Building gosync ==="
go build -tags noassets -o "$GOSYNC" .

# Clean slate
rm -rf "$HOME_A" "$HOME_B" "$SYNC_A" "$SYNC_B"

# Init
echo "=== Initializing instances ==="
$GOSYNC init --home "$HOME_A" "$SYNC_A"
$GOSYNC init --home "$HOME_B" "$SYNC_B"

ID_A=$($GOSYNC id --home "$HOME_A")
ID_B=$($GOSYNC id --home "$HOME_B")
echo "Device A: $ID_A"
echo "Device B: $ID_B"

# Pair
echo "=== Pairing ==="
$GOSYNC pair --home "$HOME_A" "$ID_B"
$GOSYNC pair --home "$HOME_B" "$ID_A"

# Start both daemons
echo "=== Starting daemons ==="
$GOSYNC run --home "$HOME_A" > /tmp/gosync-test-log-a.txt 2>&1 &
PID_A=$!
$GOSYNC run --home "$HOME_B" > /tmp/gosync-test-log-b.txt 2>&1 &
PID_B=$!

# Wait for connection
echo "=== Waiting for peer connection ==="
CONNECTED=false
for i in $(seq 1 30); do
    if grep -q "Established secure connection" /tmp/gosync-test-log-a.txt 2>/dev/null && \
       grep -q "Established secure connection" /tmp/gosync-test-log-b.txt 2>/dev/null; then
        CONNECTED=true
        break
    fi
    sleep 1
done

if [ "$CONNECTED" != "true" ]; then
    echo "FAIL: devices did not connect within 30s"
    echo "--- Log A ---"
    cat /tmp/gosync-test-log-a.txt
    echo "--- Log B ---"
    cat /tmp/gosync-test-log-b.txt
    exit 1
fi
echo "Connected."

# Create test file on A
echo "=== Creating test file on A ==="
echo "hello from gosync" > "$SYNC_A/test.txt"

# Wait for file to appear on B
echo "=== Waiting for sync (up to ${TIMEOUT}s) ==="
SYNCED=false
for i in $(seq 1 "$TIMEOUT"); do
    if [ -f "$SYNC_B/test.txt" ]; then
        CONTENT=$(cat "$SYNC_B/test.txt")
        if [ "$CONTENT" = "hello from gosync" ]; then
            SYNCED=true
            break
        fi
    fi
    sleep 1
done

echo ""
echo "--- Log A (last 20 lines) ---"
tail -20 /tmp/gosync-test-log-a.txt
echo ""
echo "--- Log B (last 20 lines) ---"
tail -20 /tmp/gosync-test-log-b.txt
echo ""

echo "--- Sync folder A ---"
ls -la "$SYNC_A/"
echo "--- Sync folder B ---"
ls -la "$SYNC_B/"

if [ "$SYNCED" = "true" ]; then
    echo ""
    echo "PASS: file synced successfully in ${i}s"
    exit 0
else
    echo ""
    echo "FAIL: file did not sync within ${TIMEOUT}s"
    exit 1
fi
