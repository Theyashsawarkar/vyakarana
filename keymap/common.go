package keymap

// The actions of the keys every tool has.
const (
	ActDown     = "down"
	ActUp       = "up"
	ActTop      = "top"
	ActBottom   = "bottom"
	ActHalfDown = "half-down"
	ActHalfUp   = "half-up"
	ActNextTab  = "next-tab"
	ActPrevTab  = "prev-tab"
	ActSearch   = "search"
	ActHelp     = "help"
	ActPalette  = "palette"
	ActQuit     = "quit"
	ActEsc      = "esc"
	ActRefresh  = "refresh"
)

// Common is the keys every tool has, with the grammar's letters. noun
// names what the list holds ("port"); tabs adds H / L (and Tab) for the
// tool's tabs; search and refresh describe / and r ("" leaves them out).
type Common struct {
	Noun, Search, Refresh string
	Tabs                  bool
}

// Binds is the common keys.
func (c Common) Binds() []Binding {
	n := c.Noun
	if n == "" {
		n = "row"
	}
	b := []Binding{
		{Seq: "j", Alt: []string{"Down"}, Act: ActDown, Desc: "Move to the next " + n, Count: true},
		{Seq: "k", Alt: []string{"Up"}, Act: ActUp, Desc: "Move to the previous " + n, Count: true},
		{Seq: "g g", Alt: []string{"Home"}, Act: ActTop, Desc: "Jump to the first " + n + " (5gg: the fifth)", Count: true},
		{Seq: "G", Alt: []string{"End"}, Act: ActBottom, Desc: "Jump to the last " + n + " (5G: the fifth)", Count: true},
		{Seq: "Ctrl+d", Alt: []string{"PgDn"}, Act: ActHalfDown, Desc: "Scroll half a page down"},
		{Seq: "Ctrl+u", Alt: []string{"PgUp"}, Act: ActHalfUp, Desc: "Scroll half a page up"},
	}
	if c.Tabs {
		b = append(b,
			Binding{Seq: "L", Alt: []string{"Tab"}, Act: ActNextTab, Desc: "Switch to the next tab"},
			Binding{Seq: "H", Alt: []string{"Shift+Tab"}, Act: ActPrevTab, Desc: "Switch to the previous tab"},
		)
	}
	if c.Search != "" {
		b = append(b, Binding{Seq: "/", Alt: []string{"Space f f"}, Act: ActSearch, Desc: c.Search})
	}
	if c.Refresh != "" {
		b = append(b, Binding{Seq: "r", Alt: []string{"Space r"}, Act: ActRefresh, Desc: c.Refresh})
	}
	return append(b,
		Binding{Seq: "?", Alt: []string{"Space ?"}, Act: ActHelp, Desc: "Show every key of this view"},
		Binding{Seq: "Space Space", Act: ActPalette, Desc: "Search every action and run one (the palette)"},
		Binding{Seq: "q", Alt: []string{"Space q"}, Act: ActQuit, Desc: "Quit"},
		Binding{Seq: "Esc", Act: ActEsc, Desc: "Go back one step: clear the filter, then quit"},
	)
}
