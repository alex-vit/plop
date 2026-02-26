#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage:
  ./scripts/release-tag.sh v1.0.0 [--watch]

Description:
  Creates and pushes an annotated git tag for release automation.
  Optionally waits for the GitHub Actions "Release" workflow run.

Examples:
  ./scripts/release-tag.sh v1.0.0
  ./scripts/release-tag.sh v1.0.1 --watch
EOF
}

if [[ $# -lt 1 || $# -gt 2 ]]; then
  usage
  exit 1
fi

version="$1"
watch=false

if [[ $# -eq 2 ]]; then
  if [[ "$2" != "--watch" ]]; then
    usage
    exit 1
  fi
  watch=true
fi

if [[ ! "$version" =~ ^v[0-9]+\.[0-9]+\.[0-9]+([-.][0-9A-Za-z.-]+)?$ ]]; then
  echo "invalid version tag: $version"
  echo "expected format like v1.0.0"
  exit 1
fi

if ! git diff --quiet || ! git diff --cached --quiet; then
  echo "working tree is not clean; commit or stash changes first"
  exit 1
fi

git fetch --tags origin

if git rev-parse -q --verify "refs/tags/$version" >/dev/null; then
  echo "tag already exists: $version"
  exit 1
fi

git tag -a "$version" -m "Release $version"
git push origin "$version"
echo "pushed tag: $version"

if [[ "$watch" != true ]]; then
  exit 0
fi

if ! command -v gh >/dev/null 2>&1; then
  echo "gh CLI is not installed; cannot watch workflow run"
  exit 0
fi

if ! gh auth status >/dev/null 2>&1; then
  echo "gh CLI is not authenticated; cannot watch workflow run"
  exit 0
fi

echo "waiting for Release workflow run for tag $version..."
run_id=""
for _ in $(seq 1 30); do
  run_id="$(gh run list --workflow release.yml --limit 30 --json databaseId,headBranch,event --jq ".[] | select(.event == \"push\" and .headBranch == \"$version\") | .databaseId" | head -n1 || true)"
  if [[ -n "$run_id" ]]; then
    break
  fi
  sleep 5
done

if [[ -z "$run_id" ]]; then
  echo "no matching workflow run found yet. inspect manually with: gh run list --workflow release.yml"
  exit 0
fi

echo "watching run id: $run_id"
gh run watch "$run_id" --exit-status
