package keymaptest

import (
	"testing"

	"github.com/Theyashsawarkar/vyakarana/keymap"
)

func TestCommonFollowsTheGrammar(t *testing.T) {
	All(t, keymap.New("demo", keymap.Common{Noun: "row", Search: "Search the rows", Refresh: "Reload the rows", Tabs: true}.Binds()))
}

func TestCatchesMistakes(t *testing.T) {
	k := keymap.New("demo", []keymap.Binding{
		{Seq: "g", Act: "x", Desc: "Go somewhere"},
		{Seq: "g g", Act: "y", Desc: "Jump to the top"},
		{Seq: "Space y z", Act: "z", Desc: "Copy it"},
		{Seq: "Space Q", Act: "q", Desc: "Quit"},
		{Seq: "a", Act: "a", Desc: "Archive it"},
		{Seq: "a", Act: "b", Desc: "Add one"},
	}).Group("Space y", "Yank")
	var errs int
	check := func(f func(*testing.T, *keymap.Keymap)) {
		inner := &testing.T{}
		done := make(chan struct{})
		go func() { defer close(done); f(inner, k) }()
		<-done
		if inner.Failed() {
			errs++
		}
	}
	check(VerbFirst)
	check(Reserved)
	check(NoPrefixClash)
	check(NoDuplicate)
	if errs != 4 {
		t.Fatalf("want all four checks to fail, %d did", errs)
	}
}
