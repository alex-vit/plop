---
title: Paid Peers
created: 2026-02-26
updated: 2026-02-26
tags: [business, cloud, peers]
status: idea
sections:
  - Hosted peer-as-a-service — always-on backup/server node with storage SLAs
---

# Paid Peers (Idea)

Monetization angle: sell hosted peers that act as always-on backup/server nodes.

## Concept

User pays a fee, gets a peer ID to add to their peers.txt. That peer is a cloud-hosted plop instance with:
- Guaranteed uptime / availability SLAs
- A storage quota (e.g. 10GB, 50GB, 100GB tiers)
- Always online — ensures sync works even when all personal devices are off

Effectively a "Dropbox server" without the Dropbox. Your files stay synced to a reliable cloud peer.

## Why it works

- Zero friction — just another peer ID in peers.txt
- No new UX concepts — users already understand peers
- Solves the "all my devices are off" problem
- Natural upsell from free P2P usage

## Untrusted / encrypted storage

Use Syncthing's "untrusted device" feature — data is encrypted client-side before syncing to the paid peer. The server stores opaque encrypted blobs and can never see file contents. Zero-knowledge backup by default.

## Open Questions

- How mature is Syncthing's untrusted device support? Any limitations (e.g. file size, metadata leakage)?
- Hosting infrastructure — VPS fleet, or something more managed?
- Storage backend — local disk, object storage, or Syncthing on top of something?
- Billing and provisioning flow — how does the user pay and receive the peer ID?
- Multi-region? Peer closest to user?
- Data retention / deletion policy when subscription lapses
- Encryption at rest on the server side
