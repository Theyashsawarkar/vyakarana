# 1. One key grammar for every tool, as a shared module

Date: 2026-11-08 · Status: accepted

## Context

pattan, nauka, vaikhari, darshan, narada and sangam each grew their own
keys. The same letter meant different things (`r` was rescan, restart,
resend, reload, request changes and reply), tabs switched six different
ways, some single keys did things that can't be undone, and the help
sheets closed on any key. The drawing kit (`internal/ui`) was copied into
four repos.

## Decision

One grammar, written down in the README: NORMAL / INSERT modes, vim's
direct keys, `Space` as the leader with groups that mean the same in every
tool (reserved letters; others free for a tool's nouns), a which-key box
after 300 ms and no timeout, a confirm that takes only `y`, and moved keys
that say where they went for one release.

It ships as this module, so the tools share code and not a description:
bindings are data (one table feeds the engine, the help sheet, the
palette, Vayu's keybinding search and the README), and `keymaptest` checks
each tool against the grammar in its own tests. Bindings name their action
instead of holding a function, because Bubble Tea models are values.

## Consequences

- A tool's keys can't drift from its docs: both are generated.
- Changing a reserved group touches every tool: it needs an ADR here.
- Until the module is published, tools use a `replace` line to `../vyakarana`.
