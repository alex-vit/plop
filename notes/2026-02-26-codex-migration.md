---
title: Codex Migration Notes
created: 2026-02-26
updated: 2026-02-26
tags: [codex, claude, workflow]
status: done
sections:
  - Claude concepts mapped to Codex concepts and files
  - Day-to-day command equivalents and permission workflow
---

# Codex Migration Notes (Done)

## Concept Mapping
- Repo instructions:
  - Claude Code: `CLAUDE.md`
  - Codex: `AGENTS.md`
- Global skill location:
  - Claude plugin skills
  - Codex skills: `~/.codex/skills/<name>/SKILL.md`
- Permission model:
  - Claude: `~/.claude/settings*.json` with `permissions.allow`
  - Codex: `~/.codex/rules/default.rules` with `prefix_rule(pattern=[...], decision="allow")`
- Session history:
  - Codex sessions: `~/.codex/sessions/...` (can be mined by `rule-forge`)

## Command/Workflow Equivalents
- Promote and clean permission rules:
  - `python3 ~/.codex/skills/rule-forge/scripts/rule_forge.py from-session`
  - `python3 ~/.codex/skills/rule-forge/scripts/rule_forge.py from-session --write`
  - `python3 ~/.codex/skills/rule-forge/scripts/rule_forge.py inspect`
- Add one new rule from a command:
  - `python3 ~/.codex/skills/rule-forge/scripts/rule_forge.py add --cmd 'git push origin master'`
- Bisect regressions:
  - Use `bisect` skill (natural language: "run bisect with <test command>")
- Sentry inspection:
  - Use `sentry` skill

## Installed Skills for Codex
- `rule-forge`
- `sentry`
- `bisect`
- `android`
- `go`

## Suggested Operating Pattern
- Keep architecture/runtime gotchas in `AGENTS.md` so agents do not depend on `CLAUDE.md`.
- Keep design history and implementation notes in `notes/`.
- Keep permission rules broad but safe; prune covered variants regularly.
