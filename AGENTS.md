# vyakarana -- notes for agents (same as CLAUDE.md)

The key grammar every Vayu terminal tool shares, as a Go module (no
binary). Used by pattan, nauka, vaikhari and darshan; it must work for any
Bubble Tea screen without Vayu.

- The grammar is in the README; it is the contract. A change to a reserved
  group, the confirm (`y` only) or the modes is a breaking change for every
  tool: write an ADR in docs/adr/ (next number; supersede, never rewrite).
- `keymap` stays free of Bubble Tea: tools feed it key strings
  (`tea.KeyMsg.String()`). Drawing and timers live in `popup`.
- Generated keybinding lines must never put the literal "# keybind:" in the
  compiled code (the search scans plugins' bin/ folders): keep the marker
  joined at run time.
- `ui` is the drawing kit: theme colours only (varna), every corner rounded.
- Every change gets a CHANGELOG.md entry; releases are tags (`vX.Y.Z`).
  `go vet ./...`, `gofmt -l .` (empty) and `go test ./...` must pass.
- The identity: assets/logo.svg and the Pāṇinīya Śikṣā verse in the
  README (checked against sanskritdocuments.org). See the tool-identity
  skill; never write or adapt a verse.
