---
title: CI Release on Tag Push
created: 2026-02-26
updated: 2026-02-26
tags: [ci, github-actions, release, windows, macos]
status: idea
sections:
  - Mirror monibright tag-trigger release
  - Build Windows and macOS release assets on tag push
  - Add a release command that tags and pushes safely
---

# CI Release on Tag Push

## Goal

When a tag is pushed, GitHub Actions should build release artifacts and publish a GitHub Release automatically.

Target artifacts:

- Windows executable (`plop.exe`)
- Windows "no-setup" artifact (portable package, e.g. zipped exe)
- macOS app bundle artifact (`Plop.app` packaged as zip for release upload)

## Neighbor Reference (`monibright`)

`../monibright/.github/workflows/release.yml` is the baseline pattern:

- Trigger: `push` on tags matching `v*.*.*`
- Runner: `windows-2022`
- Build: `go build ... -o monibright.exe`
- Package: `iscc /DAppVersion=${{ github.ref_name }} installer.iss`
- Publish: `softprops/action-gh-release` uploading built artifacts

Use this as the starting point, then extend for `plop` to include a macOS `.app` artifact.

## macOS Version Strictness (Important)

macOS bundle versions are stricter than free-form tags.

- `1.01` is invalid
- `1.1` is valid

Rules to enforce for macOS version fields (`CFBundleVersion`, `CFBundleShortVersionString`):

- No `v` prefix in plist values
- Numeric dot-separated components only
- No leading zeros in numeric components (except the value `0`)

Implementation note:

- Keep Git tag as `vX.Y.Z`
- Derive macOS plist version as `X.Y.Z` (strip leading `v`)
- Fail release command/workflow early if version does not satisfy macOS rules

## Proposed Deliverable

Add a release command/script (for example `scripts/release-tag.sh`) that:

1. Validates the requested tag format (`vX.Y.Z`) and macOS-safe numeric components.
2. Creates an annotated tag.
3. Pushes the tag (`git push origin <tag>`).
4. Relies on GitHub Actions tag trigger to build artifacts and create the release.

This gives a single local command to trigger a full release pipeline.

## CI Sketch For `plop`

1. Add `.github/workflows/release.yml` with `on.push.tags: [\"v*.*.*\"]`.
2. `windows` job:
   - build `plop.exe`
   - produce portable/no-setup artifact
   - optionally build installer if/when `installer.iss` is added
3. `macos` job:
   - run `scripts/build-mac-app.sh` with validated `VERSION`
   - zip `out/Plop.app`
4. publish job:
   - create/update GitHub Release for the tag
   - upload all artifacts

## Acceptance Criteria

- Pushing `v1.1.0` creates a GitHub Release with Windows + macOS artifacts.
- Pushing a tag like `v1.01.0` is rejected before publishing artifacts.
