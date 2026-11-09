// Package popup is the boxes every Vayu terminal tool draws over its
// screen the same way: the which-key box while a sequence is under way,
// the help sheet (?), the palette (Space Space) and the confirm (y is the
// only yes). The tool keeps each as a value in its model, sends it the
// keys while it is open, and draws it with ui.Overlay.
package popup

import (
	"strings"
	"time"
	"unicode/utf8"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Theyashsawarkar/vyakarana/keymap"
	"github.com/Theyashsawarkar/vyakarana/ui"
)

// Delay is how long a sequence waits before the which-key box shows: a
// fast typist never sees it.
const Delay = 300 * time.Millisecond

// WhichMsg says the which-key delay is over for the sequence Gen.
type WhichMsg struct{ Gen int }

// WhichAfter waits Delay, then sends WhichMsg. Show the box when the
// engine's Gen still matches (nothing was typed since).
func WhichAfter(gen int) tea.Cmd {
	return tea.Tick(Delay, func(time.Time) tea.Msg { return WhichMsg{gen} })
}

// WhichKey is the which-key box: what was typed, then each next key with
// what it does; groups end in "…". w is the screen's width.
func WhichKey(st ui.Styles, pending string, hints []keymap.Hint, w int) []string {
	if len(hints) == 0 {
		return nil
	}
	maxW := min(w-4, 100)
	col := 0
	for _, h := range hints {
		col = max(col, lipgloss.Width(h.Key)+1+lipgloss.Width(h.Desc)+1+3)
	}
	col = min(max(col, 16), 34)
	per := max(1, min(len(hints), (maxW-4)/col))
	bw := per*col + 4
	var lines []string
	line := ""
	for i, h := range hints {
		d := h.Desc
		if h.Group {
			d += "…"
		}
		ds := st.Text
		if h.Group {
			ds = st.Accent
		}
		line += ui.Fit(st.Peach.Render(h.Key)+" "+ds.Render(d), col-2) + "  "
		if (i+1)%per == 0 {
			lines = append(lines, line)
			line = ""
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	lines = append(lines, st.Muted.Render("esc cancels"))
	return st.Popup(pending+" ›", lines, bw, "")
}

// ------------------------------------------------------------- confirm

// Confirm asks before something that can't be undone. y runs Act; any
// other key, Enter included, keeps things as they are.
type Confirm struct {
	Title, Body string
	YesText     string // what y does: "stop the server"
	NoText      string // what any other key does ("" → "keep it")
	Color       string // the box's colour: red for danger, peach otherwise
	Act         tea.Cmd
}

// Answer is the confirm's answer to a key: the action for y, nil else.
// Close the confirm either way.
func (c Confirm) Answer(key string) tea.Cmd {
	if key == "y" {
		return c.Act
	}
	return nil
}

// Box draws the confirm.
func (c Confirm) Box(st ui.Styles, w int) []string {
	w = min(w-4, 72)
	no := c.NoText
	if no == "" {
		no = "keep it"
	}
	lines := append(ui.Wrap(c.Body, w-4), "", st.Key("y", c.YesText)+"   "+st.Key("any other key", no))
	return st.Popup(c.Title, lines, w, c.Color)
}

// ------------------------------------------------------------- help

// Help is the help sheet: every key of the tool, by view. It scrolls
// with j / k, Ctrl+d / Ctrl+u and g / G; q, Esc, ? or Enter close it.
// Other keys do nothing (a stray key doesn't close it).
type Help struct {
	Top int
	// last is the bottom-most Top the last Box could show. A pointer,
	// because Box runs on a copy of the model (Bubble Tea's View) and
	// its clamp must still reach the next Key; Key makes it.
	last *int
}

// Key handles a key; closed says the sheet should close.
func (h *Help) Key(key string) (closed bool) {
	if h.last == nil {
		h.last = new(int)
		*h.last = -1
	}
	if *h.last >= 0 {
		h.Top = min(h.Top, *h.last)
	}
	defer func() {
		if *h.last >= 0 {
			h.Top = min(h.Top, *h.last)
		}
	}()
	switch key {
	case "j", "down":
		h.Top++
	case "k", "up":
		h.Top = max(0, h.Top-1)
	case "ctrl+d", "pgdown", " ":
		h.Top += 10
	case "ctrl+u", "pgup":
		h.Top = max(0, h.Top-10)
	case "g", "home":
		h.Top = 0
	case "G", "end":
		h.Top = 1 << 20
	case "q", "esc", "?", "enter":
		return true
	}
	return false
}

// Box draws the sheet, at most maxH rows tall, w wide. here is the view
// the tool is on: its section comes first.
func (h *Help) Box(st ui.Styles, k *keymap.Keymap, here string, w, maxH int) []string {
	w = min(w-4, 96)
	secs := k.SectionsHere(here)
	var lines []string
	kw := 0
	for _, s := range secs {
		for _, r := range s.Rows {
			if !r.Head {
				kw = max(kw, lipgloss.Width(r.Keys))
			}
		}
	}
	kw = min(kw, 24)
	for _, s := range secs {
		if len(lines) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, st.Heading(s.Title))
		for _, r := range s.Rows {
			switch {
			case r.Head && r.Keys == "":
				lines = append(lines, st.Muted.Render(r.Desc))
			case r.Head:
				lines = append(lines, st.Peach.Render(ui.Fit(r.Keys, kw))+" "+st.Accent.Render(r.Desc+"…"))
			default:
				lines = append(lines, st.Accent.Render(ui.Fit(r.Keys, kw))+" "+r.Desc)
			}
		}
	}
	room := max(3, maxH-2)
	h.Top = min(h.Top, max(0, len(lines)-room))
	if h.last != nil {
		*h.last = max(0, len(lines)-room)
	}
	lines = lines[h.Top:min(len(lines), h.Top+room)]
	return st.Popup("keys · j/k scroll · q closes", lines, w, "")
}

// ------------------------------------------------------------- palette

// Palette is Space Space: every action of the view, filtered as you
// type; Enter runs the selected one, Esc closes it.
type Palette struct {
	Items []keymap.Binding
	Query string
	Sel   int
}

// NewPalette lists the view's actions.
func NewPalette(k *keymap.Keymap, where string) Palette {
	return Palette{Items: k.Palette(where)}
}

// Matches is the actions the query matches, best first: the query
// inside the description, then at the start of its words, then its
// letters in order anywhere (in the description or the keys).
func (p Palette) Matches() []keymap.Binding {
	q := strings.ToLower(strings.TrimSpace(p.Query))
	if q == "" {
		return p.Items
	}
	var tiers [3][]keymap.Binding
	for _, b := range p.Items {
		d := strings.ToLower(b.Desc)
		switch {
		case strings.Contains(d, q):
			tiers[0] = append(tiers[0], b)
		case initials(d, q):
			tiers[1] = append(tiers[1], b)
		case subseq(d+" "+strings.ToLower(b.Keys()), q):
			tiers[2] = append(tiers[2], b)
		}
	}
	return append(append(tiers[0], tiers[1]...), tiers[2]...)
}

// initials says whether each word of q starts a word of s, in order.
func initials(s, q string) bool {
	words := strings.Fields(s)
	i := 0
	for _, w := range strings.Fields(q) {
		for i < len(words) && !strings.HasPrefix(words[i], w) {
			i++
		}
		if i == len(words) {
			return false
		}
		i++
	}
	return true
}

func subseq(s, q string) bool {
	for _, r := range q {
		if r == ' ' {
			continue
		}
		i := strings.IndexRune(s, r)
		if i < 0 {
			return false
		}
		s = s[i+utf8.RuneLen(r):]
	}
	return true
}

// Key handles a key: act is the action to run (the palette closes), or
// closed says it closed without one.
func (p *Palette) Key(key string) (act string, closed bool) {
	m := p.Matches()
	switch key {
	case "esc", "ctrl+c":
		return "", true
	case "enter":
		if p.Sel < len(m) {
			return m[p.Sel].Act, true
		}
		return "", true
	case "down", "ctrl+j", "ctrl+n", "tab":
		p.Sel = min(max(0, len(m)-1), p.Sel+1)
	case "up", "ctrl+k", "ctrl+p", "shift+tab":
		p.Sel = max(0, p.Sel-1)
	case "backspace":
		if p.Query != "" {
			_, n := utf8.DecodeLastRuneInString(p.Query)
			p.Query = p.Query[:len(p.Query)-n]
			p.Sel = 0
		}
	case "ctrl+u":
		p.Query, p.Sel = "", 0
	case "ctrl+w":
		q := strings.TrimRight(p.Query, " ")
		if i := strings.LastIndex(q, " "); i >= 0 {
			p.Query = q[:i+1]
		} else {
			p.Query = ""
		}
		p.Sel = 0
	default:
		if r := []rune(key); len(r) == 1 && r[0] >= ' ' || key == " " {
			p.Query += key
			p.Sel = 0
		}
	}
	return "", false
}

// Box draws the palette, at most maxH rows tall.
func (p Palette) Box(st ui.Styles, w, maxH int) []string {
	w = min(w-4, 84)
	m := p.Matches()
	lines := []string{st.Accent.Render("› ") + st.Text.Render(p.Query) + st.Accent.Render("▏"), ""}
	room := max(3, maxH-4)
	top := ui.Scroll(p.Sel, 0, room)
	if p.Sel >= room {
		top = p.Sel - room + 1
	}
	for i := top; i < len(m) && i < top+room; i++ {
		b := m[i]
		keys := st.Muted.Render(b.Keys())
		left := "  " + b.Desc
		if i == p.Sel {
			left = st.Accent.Render("▌ ") + st.Bold.Render(b.Desc)
		}
		lines = append(lines, ui.Row(left, keys, w-4))
	}
	if len(m) == 0 {
		lines = append(lines, st.Muted.Render("  nothing matches"))
	}
	return st.Popup("palette · enter runs · esc closes", lines, w, "")
}

// ------------------------------------------------------------- keys

// Keys is a key message as the keys the engine takes, one at a time. Keys
// typed fast (or sent by tmux) can reach Bubble Tea as one message ("gg",
// "dd"); fed whole, the engine reads that as one unknown key. A paste
// stays whole: it is text.
func Keys(k tea.KeyMsg) []string {
	if k.Type != tea.KeyRunes || k.Paste || len(k.Runes) < 2 {
		return []string{k.String()}
	}
	out := make([]string, len(k.Runes))
	for i, r := range k.Runes {
		out[i] = tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}, Alt: k.Alt}.String()
	}
	return out
}

// ------------------------------------------------------------- tmux popup

// IsPopupKey says whether key is the tmux popup's own key (--popup KEY),
// which closes the tool too, like pressing it again outside.
func IsPopupKey(key, popupKey string) bool { return popupKey != "" && key == popupKey }
