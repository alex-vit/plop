# Own Discovery Server — Independence from Syncthing Infra

## Idea

Figure out what it takes to run a minimum viable peer discovery server so plop can eventually stop depending on Syncthing's public global announce and relay infrastructure.

## Why

- Syncthing's public discovery and relay servers are a free community resource — relying on them long-term for a separate product is not great
- If Syncthing's infra goes down or changes policy, plop breaks
- Own infra means control over availability, performance, and branding
- Needed before plop can be a "real" product

## What Syncthing Infra Plop Uses Today

1. **Global discovery servers** (`discovery-v4.syncthing.net`, `discovery-v6.syncthing.net`) — devices announce their addresses here so peers can find them outside LAN
2. **Relay servers** (`relays.syncthing.net` relay pool) — when NAT traversal fails, traffic flows through relays
3. **LAN discovery** — local broadcast, no external dependency, keep as-is

## Minimum Viable Scope

The MVP is a single small server that handles global discovery. Relays are a harder problem (bandwidth costs) and can come later — NAT traversal + LAN discovery cover many cases without relays.

### Discovery server

Syncthing already ships `stdiscosrv` as a standalone binary. Options:
- Run `stdiscosrv` as-is on a cheap VPS — it's a single Go binary, minimal resources
- Or implement a simpler protocol-compatible server if `stdiscosrv` is too heavy

### Relay server (later)

Syncthing ships `strelaysrv`. Running one is straightforward but relays carry bandwidth, so this is a cost question more than a technical one. Could start with one relay in a single region.

## Open Questions

- What are the actual resource requirements for `stdiscosrv` serving a small number of devices (tens to hundreds)?
- Can plop point to a custom discovery server while keeping Syncthing's as a fallback?
- Is the discovery protocol stable/documented enough to implement a minimal version from scratch?
- What's the monthly cost of running discovery + one relay on a small VPS?
- Should discovery be a paid-tier feature (own infra) vs free-tier (Syncthing public infra)?
- How does Syncthing's relay protocol work — is it just TCP proxying or something more complex?
- Could a single binary serve both discovery and relay?
