// Package keymap is the key grammar every Vayu terminal tool shares, as
// data: each key is a Binding (its keys, the action it runs, a verb-first
// description, where and in which mode it works), and one Engine turns
// key presses into actions: counts (5j), sequences (g g, d d) and the
// Space leader, whose groups mean the same in every tool.
//
// The tool keeps its own state and switches on the action name the
// engine returns; Bubble Tea models are values, so a binding names its
// action instead of holding a closure over a stale copy of the model.
package keymap

import (
	"sort"
	"strings"

	"github.com/Theyashsawarkar/vyakarana/mode"
)

// Binding is one key of a tool.
type Binding struct {
	Seq   string      // tokens separated by spaces: "j", "g g", "Space p m", "Ctrl+d"
	Alt   []string    // other keys that do the same ("Tab" for "L"); listed after Seq
	Act   string      // the action the tool runs (it switches on this name)
	Desc  string      // verb first, as the keybinding search wants: "Move to the next port"
	Where string      // the views it works in, space-separated ids ("" everywhere)
	Modes []mode.Mode // nil: NORMAL only
	Count bool        // takes a count (5j); the engine passes it either way
	Hide  bool        // works, but help, the palette and the search don't list it
	Moved string      // an old key: it does nothing and says "moved to <Moved>"
}

// Hint is a key that can follow what was typed so far, for the which-key
// box: either a group (more keys follow) or a binding.
type Hint struct {
	Key, Desc string
	Group     bool
}

// Reserved are the Space groups that mean the same in every tool. A tool
// may leave one empty but never use its letter for anything else
// (keymaptest.Reserved checks it).
var Reserved = map[string]string{
	"Space Space": "Palette",
	"Space ?":     "Keys",
	"Space q":     "Close",
	"Space t":     "Tabs",
	"Space u":     "Toggles",
	"Space y":     "Copy",
	"Space f":     "Find",
	"Space s":     "Sort",
	"Space r":     "Refresh",
	"Space a":     "Add",
	"Space d":     "Delete",
	"Space e":     "Edit",
	"Space o":     "Open",
	"Space g":     "Go to",
	"Space m":     "Mark",
	"Space ,":     "Settings",
	"Space k":     "Ask the AI",
}

// Free are the letters after Space a tool may give its own nouns.
const Free = "b c h i j l n p v w x z"

// Keymap is every key of one tool.
type Keymap struct {
	Tool   string
	Binds  []Binding
	groups map[string]string
	wheres []whereName
}

type whereName struct{ id, label string }

// New makes a tool's keymap.
func New(tool string, binds ...[]Binding) *Keymap {
	k := &Keymap{Tool: tool, groups: map[string]string{}}
	for _, b := range binds {
		k.Binds = append(k.Binds, b...)
	}
	return k
}

// Group names a prefix of the tool's own ("Space c" → "Container").
// The reserved groups are named already.
func (k *Keymap) Group(prefix, name string) *Keymap {
	k.groups[prefix] = name
	return k
}

// Groups is the names the tool gave its own prefixes.
func (k *Keymap) Groups() map[string]string { return k.groups }

// Where names a view for the help sheet ("ports" → "On a port"), in the
// order the sheet lists them. Keys for everywhere come first.
func (k *Keymap) Where(id, label string) *Keymap {
	k.wheres = append(k.wheres, whereName{id, label})
	return k
}

// GroupName is what a prefix is called: the tool's name for it, else the
// reserved name, else "".
func (k *Keymap) GroupName(prefix string) string {
	if n, ok := k.groups[prefix]; ok {
		return n
	}
	return Reserved[prefix]
}

// Keys is a binding's keys as the help and the search show them:
// "j / Down".
func (b Binding) Keys() string {
	return strings.Join(append([]string{b.Seq}, b.Alt...), " / ")
}

// In says whether the binding works in a mode and view.
func (b Binding) In(m mode.Mode, where string) bool {
	modes := b.Modes
	if len(modes) == 0 {
		modes = []mode.Mode{mode.Normal}
	}
	ok := false
	for _, x := range modes {
		if x == m {
			ok = true
		}
	}
	if !ok {
		return false
	}
	return b.Where == "" || where != "" && hasWord(b.Where, where)
}

func hasWord(list, w string) bool {
	for _, f := range strings.Fields(list) {
		if f == w {
			return true
		}
	}
	return false
}

// Leader says whether the binding starts with Space.
func (b Binding) Leader() bool { return strings.HasPrefix(b.Seq, "Space ") }

func (b Binding) seqs() []string { return append([]string{b.Seq}, b.Alt...) }

// lookup finds the binding seq names in a mode and view, and whether a
// longer one starts with it.
func (k *Keymap) lookup(seq string, m mode.Mode, where string) (*Binding, bool) {
	var exact *Binding
	longer := false
	for i := range k.Binds {
		b := &k.Binds[i]
		if !b.In(m, where) {
			continue
		}
		for _, s := range b.seqs() {
			if s == seq && exact == nil {
				exact = b
			}
			if strings.HasPrefix(s, seq+" ") {
				longer = true
			}
		}
	}
	return exact, longer
}

// Next is what can follow the keys typed so far, for the which-key box:
// one hint per next key, groups named by GroupName.
func (k *Keymap) Next(pending string, m mode.Mode, where string) []Hint {
	var out []Hint
	seen := map[string]int{}
	for _, b := range k.Binds {
		if b.Hide || b.Moved != "" || !b.In(m, where) {
			continue
		}
		for _, s := range b.seqs() {
			if !strings.HasPrefix(s, pending+" ") {
				continue
			}
			rest := strings.Fields(strings.TrimPrefix(s, pending+" "))
			if len(rest) == 0 {
				continue
			}
			h := Hint{Key: rest[0], Desc: b.Desc}
			if len(rest) > 1 {
				h.Group = true
				h.Desc = k.GroupName(pending + " " + rest[0])
				if h.Desc == "" {
					h.Desc = "more"
				}
			}
			if i, ok := seen[h.Key]; ok {
				if h.Group && !out[i].Group {
					out[i] = h
				}
				continue
			}
			seen[h.Key] = len(out)
			out = append(out, h)
		}
	}
	return out
}

// Row is a line of the help sheet; a heading has no keys.
type Row struct {
	Keys, Desc string
	Head       bool
}

// Section is one view's keys on the help sheet.
type Section struct {
	Title string
	Rows  []Row
}

// Sections is the help sheet: everywhere first, then each named view;
// in each, the direct keys, then the Space keys by group, then the keys
// that work while typing.
func (k *Keymap) Sections() []Section {
	order := []string{""}
	for _, w := range k.wheres {
		order = append(order, w.id)
	}
	seen := map[string]bool{}
	for _, w := range order {
		seen[w] = true
	}
	for _, b := range k.Binds { // views nobody named, as they come
		if !seen[b.Where] {
			seen[b.Where] = true
			order = append(order, b.Where)
		}
	}
	title := k.title
	var out []Section
	for _, where := range order {
		var direct, leader, typing []Binding
		for _, b := range k.Binds {
			if b.Hide || b.Moved != "" || b.Where != where {
				continue
			}
			w := firstWord(where)
			switch {
			case !b.In(mode.Normal, w) && !b.In(mode.Visual, w):
				typing = append(typing, b)
			case b.Leader():
				leader = append(leader, b)
			default:
				direct = append(direct, b)
			}
		}
		sort.SliceStable(leader, func(i, j int) bool { return groupOf(leader[i].Seq) < groupOf(leader[j].Seq) })
		var rows []Row
		for _, b := range direct {
			rows = append(rows, Row{Keys: b.Keys(), Desc: b.Desc})
		}
		group := ""
		for _, b := range leader {
			if g := groupOf(b.Seq); g != group && strings.Count(b.Seq, " ") > 1 {
				group = g
				rows = append(rows, Row{Keys: g, Desc: k.GroupName(g), Head: true})
			}
			rows = append(rows, Row{Keys: b.Keys(), Desc: b.Desc})
		}
		if len(typing) > 0 {
			rows = append(rows, Row{Desc: "While typing", Head: true})
			for _, b := range typing {
				rows = append(rows, Row{Keys: b.Keys(), Desc: b.Desc})
			}
		}
		if len(rows) > 0 {
			out = append(out, Section{Title: title(where), Rows: rows})
		}
	}
	return out
}

// SectionsHere is Sections with the view here's own sections first (any
// whose views include here), then the rest in their order.
func (k *Keymap) SectionsHere(here string) []Section {
	all := k.Sections()
	if here == "" {
		return all
	}
	titles := map[string]bool{}
	for _, b := range k.Binds {
		if b.Where != "" && hasWord(b.Where, here) {
			titles[k.title(b.Where)] = true
		}
	}
	var first, rest []Section
	for _, s := range all {
		if titles[s.Title] {
			first = append(first, s)
		} else {
			rest = append(rest, s)
		}
	}
	return append(first, rest...)
}

func firstWord(s string) string {
	if f := strings.Fields(s); len(f) > 0 {
		return f[0]
	}
	return ""
}

// title is a section's title: the views' labels, joined.
func (k *Keymap) title(where string) string {
	if where == "" {
		return "Everywhere"
	}
	var parts []string
	for _, w := range strings.Fields(where) {
		l := w
		for _, n := range k.wheres {
			if n.id == w {
				l = n.label
				break
			}
		}
		parts = append(parts, l)
	}
	return strings.Join(parts, " and ")
}

// groupOf is a leader sequence's group: "Space p m" → "Space p".
func groupOf(seq string) string {
	f := strings.Fields(seq)
	if len(f) < 2 {
		return seq
	}
	return f[0] + " " + f[1]
}

// Palette is what Space Space lists in a view: every NORMAL key with a
// description, except the palette itself.
func (k *Keymap) Palette(where string) []Binding {
	var out []Binding
	for _, b := range k.Binds {
		if b.Hide || b.Moved != "" || b.Act == "" || b.Seq == "Space Space" || !b.In(mode.Normal, where) {
			continue
		}
		out = append(out, b)
	}
	return out
}

// Token turns a key as Bubble Tea spells it (tea.KeyMsg.String()) into
// the grammar's spelling: " " → "Space", "ctrl+d" → "Ctrl+d",
// "enter" → "Enter".
func Token(s string) string {
	if n, ok := names[s]; ok {
		return n
	}
	for _, p := range []struct{ tea, ours string }{{"ctrl+", "Ctrl+"}, {"alt+", "Alt+"}, {"shift+", "Shift+"}} {
		if strings.HasPrefix(s, p.tea) && len(s) > len(p.tea) {
			return p.ours + Token(s[len(p.tea):])
		}
	}
	return s
}

var names = map[string]string{
	" ": "Space", "space": "Space", "enter": "Enter", "esc": "Esc", "tab": "Tab",
	"shift+tab": "Shift+Tab", "backspace": "Backspace", "delete": "Delete",
	"up": "Up", "down": "Down", "left": "Left", "right": "Right",
	"home": "Home", "end": "End", "pgup": "PgUp", "pgdown": "PgDn",
}
