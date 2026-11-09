package popup

import (
	"strings"
	"testing"

	"github.com/Theyashsawarkar/vyakarana/keymap"
	"github.com/Theyashsawarkar/vyakarana/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func TestConfirmOnlyY(t *testing.T) {
	ran := false
	c := Confirm{Act: func() tea.Msg { ran = true; return nil }}
	_ = ran
	for _, k := range []string{"enter", "Y", "n", " ", "esc"} {
		if c.Answer(k) != nil {
			t.Errorf("%q confirmed", k)
		}
	}
	if c.Answer("y") == nil {
		t.Error("y didn't confirm")
	}
}

func TestPalette(t *testing.T) {
	k := keymap.New("demo", keymap.Common{Noun: "row", Refresh: "Reload the rows"}.Binds())
	p := NewPalette(k, "")
	for _, r := range "reload" {
		p.Key(string(r))
	}
	m := p.Matches()
	if len(m) == 0 || m[0].Act != keymap.ActRefresh {
		t.Fatalf("matches: %+v", m)
	}
	if act, closed := p.Key("enter"); act != keymap.ActRefresh || !closed {
		t.Fatalf("enter: %q %v", act, closed)
	}
}

func TestBoxesDraw(t *testing.T) {
	st := ui.New(ui.Mocha)
	k := keymap.New("demo", keymap.Common{Noun: "row", Tabs: true}.Binds())
	var h Help
	if b := h.Box(st, k, "", 100, 30); len(b) < 5 || !strings.Contains(strings.Join(b, "\n"), "EVERYWHERE") {
		t.Errorf("help: %v", b)
	}
	if h.Key("x") {
		t.Error("a stray key closed the help")
	}
	if !h.Key("q") {
		t.Error("q didn't close the help")
	}
	if b := WhichKey(st, "Space", k.Next("Space", 0, ""), 100); len(b) < 3 {
		t.Errorf("which-key: %v", b)
	}
}

// The tool draws the sheet from a copy of its model (Bubble Tea's View),
// so the clamp Box does must still reach the next Key: after G, one k
// scrolls up one line, not from a million.
func TestHelpScrollsBackAfterG(t *testing.T) {
	st := ui.New(ui.Mocha)
	k := keymap.New("demo", keymap.Common{Noun: "row", Tabs: true, Search: "Search the rows", Refresh: "Reload the rows"}.Binds())
	var h Help
	draw := func() string { c := h; return strings.Join(c.Box(st, k, "", 100, 8), "\n") }
	top := draw()
	h.Key("G")
	bottom := draw()
	if bottom == top {
		t.Fatal("G didn't scroll")
	}
	h.Key("k")
	if up := draw(); up == bottom {
		t.Fatal("k after G didn't scroll up")
	}
	for range 50 {
		h.Key("j")
	}
	draw()
	h.Key("k")
	if up := draw(); up == bottom {
		t.Fatal("k after many j didn't scroll up")
	}
}

// The view the tool is on comes first on the sheet.
func TestHelpHereFirst(t *testing.T) {
	st := ui.New(ui.Mocha)
	k := keymap.New("demo", keymap.Common{Noun: "row"}.Binds(), []keymap.Binding{
		{Seq: "d d", Act: "stop", Desc: "Stop the server", Where: "ports"},
	}).Where("ports", "On a port")
	var h Help
	b := strings.Join(h.Box(st, k, "ports", 100, 60), "\n")
	if i, j := strings.Index(b, "ON A PORT"), strings.Index(b, "EVERYWHERE"); i < 0 || j < 0 || i > j {
		t.Fatalf("ports isn't first:\n%s", b)
	}
}

func TestKeysSplitsBursts(t *testing.T) {
	got := Keys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("gg")})
	if strings.Join(got, ",") != "g,g" {
		t.Errorf("burst: %q", got)
	}
	if got := Keys(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("gg"), Paste: true}); len(got) != 1 {
		t.Errorf("paste split: %q", got)
	}
	if got := Keys(tea.KeyMsg{Type: tea.KeyCtrlD}); strings.Join(got, ",") != "ctrl+d" {
		t.Errorf("ctrl+d: %q", got)
	}
}
