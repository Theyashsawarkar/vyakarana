package keymap

import (
	"strings"
	"testing"

	"github.com/Theyashsawarkar/vyakarana/mode"
)

func testMap() *Keymap {
	return New("demo",
		Common{Noun: "port", Search: "Search the ports", Refresh: "Rescan the ports", Tabs: true}.Binds(),
		[]Binding{
			{Seq: "d d", Act: "stop", Desc: "Stop the server", Where: "ports"},
			{Seq: "Space c r", Act: "restart", Desc: "Restart the container", Where: "ports"},
			{Seq: "Space y u", Act: "copy-url", Desc: "Copy the URL"},
			{Seq: "1", Act: "one", Desc: "Switch to the first tab", Where: "tabs"},
			{Seq: "x", Moved: "d d"},
			{Seq: "Esc", Act: "normal", Desc: "Leave the filter", Modes: []mode.Mode{mode.Insert}},
			{Seq: "Ctrl+j", Act: "down", Desc: "Move down", Modes: []mode.Mode{mode.Insert}},
		}).Group("Space c", "Container")
}

func feed(e *Engine, where string, keys ...string) Result {
	var r Result
	for _, k := range keys {
		r = e.Feed(k, where)
	}
	return r
}

func TestSimpleAndAlt(t *testing.T) {
	e := NewEngine(testMap(), mode.Normal)
	if r := feed(&e, "", "j"); r.Act != ActDown || r.Count != 1 || r.Counted {
		t.Fatalf("j: %+v", r)
	}
	if r := feed(&e, "", "down"); r.Act != ActDown {
		t.Fatalf("down: %+v", r)
	}
	if r := feed(&e, "", "tab"); r.Act != ActNextTab {
		t.Fatalf("tab: %+v", r)
	}
}

func TestCount(t *testing.T) {
	e := NewEngine(testMap(), mode.Normal)
	if r := feed(&e, "", "1", "2", "j"); r.Act != ActDown || r.Count != 12 || !r.Counted {
		t.Fatalf("12j: %+v", r)
	}
	if r := feed(&e, "", "5", "g", "g"); r.Act != ActTop || r.Count != 5 {
		t.Fatalf("5gg: %+v", r)
	}
	// A view that binds 1 itself gets it.
	if r := feed(&e, "tabs", "1"); r.Act != "one" {
		t.Fatalf("1 in tabs: %+v", r)
	}
	// Once a count has started, a digit the view binds is part of it.
	if r := feed(&e, "tabs", "2", "1", "j"); r.Act != ActDown || r.Count != 21 {
		t.Fatalf("21j in tabs: %+v", r)
	}
	// 0 alone is not a count.
	if r := feed(&e, "", "0"); r.Pending {
		t.Fatalf("0: %+v", r)
	}
}

func TestSequencesAndLeader(t *testing.T) {
	e := NewEngine(testMap(), mode.Normal)
	if r := feed(&e, "ports", "d"); !r.Pending || e.Pending() != "d" {
		t.Fatalf("d: %+v %q", r, e.Pending())
	}
	if r := feed(&e, "ports", "d"); r.Act != "stop" {
		t.Fatalf("dd: %+v", r)
	}
	if r := feed(&e, "ports", " ", "c", "r"); r.Act != "restart" {
		t.Fatalf("Space c r: %+v", r)
	}
	// d d only works on ports.
	if r := feed(&e, "", "d"); r.Pending || r.Unknown != "d" {
		t.Fatalf("d elsewhere: %+v", r)
	}
	// Esc cancels.
	feed(&e, "ports", " ")
	if r := feed(&e, "ports", "esc"); !r.Cancel || e.Waiting() {
		t.Fatalf("esc: %+v", r)
	}
	// A wrong second key says so.
	if r := feed(&e, "ports", " ", "c", "z"); r.Unknown != "Space c z" || r.Message() == "" {
		t.Fatalf("Space c z: %+v", r)
	}
}

func TestMovedAndInsert(t *testing.T) {
	e := NewEngine(testMap(), mode.Normal)
	if r := feed(&e, "", "x"); r.Moved != "d d" || !strings.Contains(r.Message(), "moved to d d") {
		t.Fatalf("x: %+v", r)
	}
	e.SetMode(mode.Insert)
	if r := feed(&e, "", "j"); r.Handled {
		t.Fatalf("j while typing is the field's: %+v", r)
	}
	if r := feed(&e, "", "ctrl+j"); r.Act != "down" {
		t.Fatalf("ctrl+j: %+v", r)
	}
	if r := feed(&e, "", "esc"); r.Act != "normal" {
		t.Fatalf("esc: %+v", r)
	}
}

func TestHints(t *testing.T) {
	e := NewEngine(testMap(), mode.Normal)
	feed(&e, "ports", " ")
	hs := e.Hints("ports")
	got := map[string]Hint{}
	for _, h := range hs {
		got[h.Key] = h
	}
	if h := got["c"]; !h.Group || h.Desc != "Container" {
		t.Errorf("c: %+v", h)
	}
	if h := got["y"]; !h.Group || h.Desc != "Copy" {
		t.Errorf("y: %+v", h)
	}
	if h := got["Space"]; h.Group || h.Desc == "" {
		t.Errorf("Space: %+v", h)
	}
}

func TestGenerated(t *testing.T) {
	k := testMap()
	f := k.KeybindingsFile("Tmux", "go test -update")
	for _, want := range []string{
		"Tmux/demo:normal | j / Down | Move to the next port",
		"Tmux/demo ports:normal | d d | Stop the server",
		"Tmux/demo:insert | Esc | Leave the filter",
		"# keybind-group: Tmux/demo | Space c | Container",
	} {
		if !strings.Contains(f, want) {
			t.Errorf("missing %q in\n%s", want, f)
		}
	}
	if strings.Contains(f, "moved") {
		t.Error("moved keys are listed")
	}
	md := k.Markdown()
	if !strings.Contains(md, "<kbd>Space</kbd> <kbd>c</kbd>") || !strings.Contains(md, "*Container…*") {
		t.Errorf("markdown:\n%s", md)
	}
}

func TestToken(t *testing.T) {
	for in, want := range map[string]string{" ": "Space", "ctrl+d": "Ctrl+d", "alt+enter": "Alt+Enter", "G": "G", "shift+tab": "Shift+Tab", "pgdown": "PgDn"} {
		if got := Token(in); got != want {
			t.Errorf("Token(%q) = %q, want %q", in, got, want)
		}
	}
}
