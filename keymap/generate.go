package keymap

import (
	"fmt"
	"strings"

	"github.com/Theyashsawarkar/vyakarana/mode"
)

// KeybindingsFile is the file Vayu's keybinding search reads for a tmux
// plugin (scripts/keybindings): one annotation per key, its scope
// "<tool> <where>:<mode>" ("pattan ports:normal"), so the search can show
// and filter keys by view and mode. source is the search's source name
// ("Tmux"); regen is the command that rewrites the file.
func (k *Keymap) KeybindingsFile(source, regen string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "# Keys inside the %s screen, for Vayu's keybinding search\n", k.Tool)
	b.WriteString("# (keybind-search.py reads ~/.tmux/plugins/*/scripts). Generated from\n")
	fmt.Fprintf(&b, "# the keymap by `%s`; don't edit.\n", regen)
	// The marker is kept apart: the binary holds this code, and the
	// keybinding search scans plugins' bin/ folders.
	marker := strings.Join([]string{"#", "keybind"}, " ")
	// Each Space group's name, so the search can show it and check it
	// against the reserved meanings.
	seen := map[string]bool{}
	for _, bd := range k.Binds {
		if bd.Hide || bd.Moved != "" {
			continue
		}
		for _, s := range bd.seqs() {
			f := strings.Fields(s)
			if len(f) < 3 || f[0] != "Space" {
				continue
			}
			g := f[0] + " " + f[1]
			if name := k.GroupName(g); name != "" && !seen[g] {
				seen[g] = true
				fmt.Fprintf(&b, "%s-group: %s/%s | %s | %s\n", marker, source, k.Tool, g, name)
			}
		}
	}
	for _, bd := range k.Binds {
		if bd.Hide || bd.Moved != "" {
			continue
		}
		modes := bd.Modes
		if len(modes) == 0 {
			modes = []mode.Mode{mode.Normal}
		}
		wheres := strings.Fields(bd.Where)
		if len(wheres) == 0 {
			wheres = []string{""}
		}
		for _, w := range wheres {
			scope := k.Tool
			if w != "" {
				scope += " " + w
			}
			fmt.Fprintf(&b, "%s: %s/%s:%s | %s | %s\n", marker, source, scope, modes[0].Lower(), bd.Keys(), bd.Desc)
		}
	}
	return b.String()
}

// Markdown is the README's key tables: one per section of the help sheet.
func (k *Keymap) Markdown() string {
	var b strings.Builder
	for i, s := range k.Sections() {
		if i > 0 {
			b.WriteString("\n")
		}
		fmt.Fprintf(&b, "**%s**\n\n| Keys | What it does |\n|---|---|\n", s.Title)
		for _, r := range s.Rows {
			if r.Head {
				if r.Keys == "" {
					fmt.Fprintf(&b, "| | *%s* |\n", r.Desc)
				} else {
					fmt.Fprintf(&b, "| %s | *%s…* |\n", caps(r.Keys), r.Desc)
				}
				continue
			}
			fmt.Fprintf(&b, "| %s | %s |\n", caps(r.Keys), strings.ReplaceAll(r.Desc, "|", "\\|"))
		}
	}
	return b.String()
}

// caps writes keys as <kbd> caps: "Space p m / Ctrl+s".
func caps(keys string) string {
	var alts []string
	for _, alt := range strings.Split(keys, " / ") {
		var toks []string
		for _, t := range strings.Fields(alt) {
			t = strings.ReplaceAll(t, "|", "\\|")
			toks = append(toks, "<kbd>"+t+"</kbd>")
		}
		alts = append(alts, strings.Join(toks, " "))
	}
	return strings.Join(alts, " or ")
}
