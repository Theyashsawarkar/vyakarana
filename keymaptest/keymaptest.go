// Package keymaptest checks a tool's keymap against the grammar in its
// tests: verb-first descriptions, the reserved Space groups, no key that
// is both a key and the start of a longer one, and an up-to-date
// scripts/keybindings.
package keymaptest

import (
	"os"
	"strings"
	"testing"

	"github.com/Theyashsawarkar/vyakarana/keymap"
	"github.com/Theyashsawarkar/vyakarana/mode"
)

// Verbs is the first words Vayu's keybinding search accepts
// (keybind-search.py's VERBS).
var Verbs = strings.Fields(`add alternate approve begin build cancel clear close confirm connect
continue copy create cut cycle decline delete describe detach disconnect edit enter evaluate exit
focus follow force format go grow hover install invert jump kill leave lock lower merge move normal
open paste pause pick play post pull quit raise rate record reload remove rename reopen request
rerun rescan reset resize restart restore rotate run save schedule scroll search seek select send
set show shrink shuffle split start step stop swap switch take toggle trash undo update`)

// VerbFirst fails for a description that doesn't start with a verb the
// search knows.
func VerbFirst(t *testing.T, k *keymap.Keymap) {
	t.Helper()
	ok := map[string]bool{}
	for _, v := range Verbs {
		ok[v] = true
	}
	for _, b := range k.Binds {
		if b.Hide || b.Moved != "" {
			continue
		}
		f := strings.Fields(b.Desc)
		if len(f) == 0 || !ok[strings.ToLower(strings.Split(f[0], "/")[0])] {
			t.Errorf("%s: %q doesn't start with a verb the keybinding search knows", b.Seq, b.Desc)
		}
	}
}

// Reserved fails when the tool renames a reserved Space group, or names
// a group outside Space with a reserved name's prefix.
func Reserved(t *testing.T, k *keymap.Keymap) {
	t.Helper()
	for prefix, name := range k.Groups() {
		if want, ok := keymap.Reserved[prefix]; ok && want != name {
			t.Errorf("%s is the reserved group %q in every tool; this tool calls it %q", prefix, want, name)
		}
	}
	free := map[string]bool{}
	for _, l := range strings.Fields(keymap.Free) {
		free[l] = true
	}
	for _, b := range k.Binds {
		for _, s := range append([]string{b.Seq}, b.Alt...) {
			f := strings.Fields(s)
			if len(f) >= 2 && f[0] == "Space" {
				if _, reserved := keymap.Reserved["Space "+f[1]]; !reserved && !free[f[1]] {
					t.Errorf("%s: Space %s is neither a reserved group nor a free letter (%s)", s, f[1], keymap.Free)
				}
			}
		}
	}
}

// NoPrefixClash fails when a key is bound and also starts a longer key in
// the same mode and view: the longer one would wait forever for it. Like
// NoDuplicate, it leaves out bindings that only document a key the tool
// reads itself (no Act, no Moved): the engine never sees them.
func NoPrefixClash(t *testing.T, k *keymap.Keymap) {
	t.Helper()
	live := func(b keymap.Binding) bool { return b.Act != "" || b.Moved != "" }
	for i, a := range k.Binds {
		for _, b := range k.Binds[i:] {
			if !live(a) || !live(b) {
				continue
			}
			for _, sa := range append([]string{a.Seq}, a.Alt...) {
				for _, sb := range append([]string{b.Seq}, b.Alt...) {
					short, long := sa, sb
					if len(short) > len(long) {
						short, long = long, short
					}
					if !strings.HasPrefix(long, short+" ") || !overlap(a, b) {
						continue
					}
					t.Errorf("%q (%s) is also the start of %q (%s)", short, a.Desc, long, b.Desc)
				}
			}
		}
	}
}

// NoDuplicate fails when two bindings have the same key in the same mode
// and view: the engine runs the first, and the second never works. A
// binding with neither Act nor Moved only documents a key the tool reads
// itself, and is left out.
func NoDuplicate(t *testing.T, k *keymap.Keymap) {
	t.Helper()
	live := func(b keymap.Binding) bool { return b.Act != "" || b.Moved != "" }
	for i, a := range k.Binds {
		for _, b := range k.Binds[i+1:] {
			if !live(a) || !live(b) {
				continue
			}
			for _, sa := range append([]string{a.Seq}, a.Alt...) {
				for _, sb := range append([]string{b.Seq}, b.Alt...) {
					if sa == sb && overlap(a, b) {
						t.Errorf("%q is bound twice: %q and %q (only the first works)", sa, a.Desc+a.Moved, b.Desc+b.Moved)
					}
				}
			}
		}
	}
}

// overlap says whether two bindings can be live at once.
func overlap(a, b keymap.Binding) bool {
	modes := func(x keymap.Binding) []mode.Mode {
		if len(x.Modes) == 0 {
			return []mode.Mode{mode.Normal}
		}
		return x.Modes
	}
	shareMode := false
	for _, ma := range modes(a) {
		for _, mb := range modes(b) {
			if ma == mb {
				shareMode = true
			}
		}
	}
	if !shareMode {
		return false
	}
	if a.Where == "" || b.Where == "" {
		return true
	}
	for _, w := range strings.Fields(a.Where) {
		for _, v := range strings.Fields(b.Where) {
			if w == v {
				return true
			}
		}
	}
	return false
}

// FileUpToDate fails when the file at path isn't want; with update it
// rewrites it first. For scripts/keybindings:
//
//	keymaptest.FileUpToDate(t, "../../scripts/keybindings", km.KeybindingsFile(...), *update)
func FileUpToDate(t *testing.T, path, want string, update bool) {
	t.Helper()
	if update {
		if err := os.WriteFile(path, []byte(want), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, _ := os.ReadFile(path)
	if string(got) != want {
		t.Fatalf("%s is out of date: run the tests with -update", path)
	}
}

// All runs VerbFirst, Reserved, NoPrefixClash and NoDuplicate.
func All(t *testing.T, k *keymap.Keymap) {
	t.Helper()
	VerbFirst(t, k)
	Reserved(t, k)
	NoPrefixClash(t, k)
	NoDuplicate(t, k)
}

// Start and End mark the generated key tables in a README.
const (
	Start = "<!-- keys: generated from the keymap; run the tests with -update -->"
	End   = "<!-- /keys -->"
)

// SectionUpToDate fails when the text between Start and End in the file
// at path isn't want (the README's key tables, keymap.Markdown()); with
// update it rewrites that part.
func SectionUpToDate(t *testing.T, path, want string, update bool) {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	s := string(b)
	i, j := strings.Index(s, Start), strings.Index(s, End)
	if i < 0 || j < i {
		t.Fatalf("%s has no %q ... %q markers around its key tables", path, Start, End)
	}
	body := "\n" + strings.TrimRight(want, "\n") + "\n"
	got := s[i+len(Start) : j]
	if update && got != body {
		s = s[:i+len(Start)] + body + s[j:]
		if err := os.WriteFile(path, []byte(s), 0o644); err != nil {
			t.Fatal(err)
		}
		got = body
	}
	if got != body {
		t.Fatalf("%s's key tables are out of date: run the tests with -update", path)
	}
}
