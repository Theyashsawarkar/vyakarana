# Changelog

Newest first. Releases are tags (`vX.Y.Z`).

## Unreleased (v0.2.0)

- `Backspace` goes up a level of a pending sequence (`Space m` → `Space`; on `Space` alone it cancels), as which-key does in Doom and Spacemacs; `Result.Back` says it did. With nothing pending it is an ordinary key. narada is the first to use it.
- `keymaptest.NoPrefixClash` leaves out bindings that only document keys a screen reads itself (no `Act`, no `Moved`), as `NoDuplicate` already did: a dialog's `d` no longer clashes with NORMAL's `d d`.

## v0.1.0 -- 2026-11-09

- First version: the key grammar every Vayu terminal tool shares, as a Go
  module. `keymap` (bindings as data, an engine with counts, sequences and
  the `Space` leader, the reserved groups, generated keybinding-search
  lines and README tables), `mode`, `popup` (which-key box, help sheet,
  palette, a confirm that takes only `y`), `ui` (the drawing kit pattan,
  nauka, vaikhari and darshan each had a copy of), `text` (kitty's
  diacritics) and `keymaptest`.
- Not pushed yet: pattan, nauka, vaikhari and darshan use it through a
  `replace ... => ../vyakarana` line in their go.mod until it is published
  and tagged.
- Review fixes before the first tag: the help sheet scrolls back up after
  `G` (or `j` past the end) even though Bubble Tea draws it from a copy of
  the model; `Help.Box` now puts the current view's section first, as it
  said it did (`Keymap.SectionsHere`); `keymaptest.NoDuplicate` (in `All`)
  catches a key bound twice in the same mode and view, and `Reserved` checks
  the `Alt` keys too. Once a count has started, every digit is part of it
  (`72j` in a view that binds `2`). `popup.Keys` splits keys Bubble Tea
  delivers together ("gg") so the engine can take them one at a time. The
  README has its shloka (Pāṇinīya Śikṣā 57).
