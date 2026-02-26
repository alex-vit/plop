#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
APP_BUNDLE="${ROOT_DIR}/out/Plop.app"
APP_EXEC="${APP_BUNDLE}/Contents/MacOS/plop"
DO_BUILD=1

usage() {
  cat <<'EOF'
Usage: ./scripts/restart-mac-app.sh [--no-build]

Restarts Plop.app for local testing.
- default: rebuild app bundle, stop running app process, relaunch app
- --no-build: skip rebuild and only restart
EOF
}

for arg in "$@"; do
  case "${arg}" in
    --no-build)
      DO_BUILD=0
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      echo "unknown argument: ${arg}" >&2
      usage
      exit 1
      ;;
  esac
done

cd "${ROOT_DIR}"

if [[ "${DO_BUILD}" -eq 1 ]]; then
  ./scripts/build-mac-app.sh
fi

# Ask the app to quit (best effort), then ensure the app-bundle process is gone.
osascript -e 'tell application "Plop" to quit' >/dev/null 2>&1 || true

if pgrep -f "${APP_EXEC}" >/dev/null 2>&1; then
  pkill -TERM -f "${APP_EXEC}" || true
  for _ in {1..25}; do
    if ! pgrep -f "${APP_EXEC}" >/dev/null 2>&1; then
      break
    fi
    sleep 0.2
  done
fi

if pgrep -f "${APP_EXEC}" >/dev/null 2>&1; then
  echo "plop app process did not exit after SIGTERM" >&2
  exit 1
fi

open "${APP_BUNDLE}"
echo "Restarted ${APP_BUNDLE}"
