# Syncthing Interop — Pair with "Real" Syncthing

## Idea

Make it easy and seamless to add a regular Syncthing instance (e.g. the Android Syncthing app) as a peer to plop. The main use case is syncing the plop folder to/from a phone.

## Why

- Syncthing has a mature Android app (Syncthing-Fork is actively maintained)
- plop already speaks BEP and uses Syncthing's libraries — it *is* Syncthing under the hood
- Users shouldn't have to understand Syncthing internals to pair their phone

## What "Hard to Get Wrong" Means

Syncthing's UI exposes a lot of knobs. A user pairing with plop should only need to:

1. Exchange device IDs (plop already has `plop id` + `plop pair`)
2. Share exactly the one default folder (plop's sync folder)
3. Get the folder ID and label right on both sides

Things that can go wrong today if someone tries manually:
- Folder ID mismatch — plop uses a hardcoded folder ID; Syncthing defaults to generating a new one
- Wrong folder path on the remote side
- Encryption / untrusted device confusion
- Ignoring `.stignore` differences
- Introducer / auto-accept settings causing unexpected behavior

## Scope

Focus on the "plop folder only" case — one folder, send-receive, no encryption passwords, no introducer chains. The goal is a short, concrete guide (and maybe a `plop pair --help` enhancement) that walks someone through adding Syncthing Android as a peer with the right folder ID and settings.

## Open Questions

- Should `plop pair` detect that it's being paired with a full Syncthing (vs another plop) and adjust behavior?
- Should plop print/display the folder ID somewhere prominent to make Android setup easier?
- Is there a way to generate a QR code or deep link that Syncthing Android can scan to auto-configure?
- Should there be a `plop pair --syncthing` mode that prints step-by-step instructions?
- What about iOS? (No official Syncthing app, but Möbius Sync exists)
