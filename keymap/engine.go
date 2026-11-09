package keymap

import (
	"strings"

	"github.com/Theyashsawarkar/vyakarana/mode"
)

// Engine turns key presses into actions. It is a value that lives in the
// tool's model; copy it with the model (its state is plain values).
//
// In NORMAL and VISUAL a digit starts a count (5j) unless the view binds
// that digit itself; a key that starts a longer sequence waits for the
// rest (g g, Space p m), for as long as it takes: there is no timeout,
// Esc cancels, Backspace goes up a level. The which-key box is the tool's to show, after a short
// delay, while Pending (popup.WhichAfter).
//
// In INSERT and COMMAND only the keys bound for that mode are the
// engine's (Esc, Enter, Ctrl+s...); everything else is the text field's.
type Engine struct {
	Map     *Keymap
	Mode    mode.Mode
	pending string
	count   int
	gen     int
}

// Result is what a key press did.
type Result struct {
	Act     string // the action to run ("" for none)
	Count   int    // the count typed before it, 1 when none
	Counted bool   // a count was typed
	Pending bool   // a count or a sequence is under way: wait for more
	Gen     int    // compare with the which-key timer's (popup.WhichMsg)
	Cancel  bool   // Esc cancelled a count or a sequence
	Back    bool   // Backspace went up a level of a sequence
	Unknown string // the keys, when they mean nothing here
	Moved   string // the keys were moved: say "<Unknown> moved to <Moved>"
	Handled bool   // false: the key belongs to the text field
}

// NewEngine starts in mode m.
func NewEngine(k *Keymap, m mode.Mode) Engine { return Engine{Map: k, Mode: m} }

// Pending is the keys typed so far, spelled ("Space p", "5", "3 g").
func (e Engine) Pending() string {
	c := ""
	if e.count > 0 {
		c = itoa(e.count)
	}
	switch {
	case c != "" && e.pending != "":
		return c + " " + e.pending
	case c != "":
		return c
	}
	return e.pending
}

// Waiting says whether a count or a sequence is under way.
func (e Engine) Waiting() bool { return e.pending != "" || e.count > 0 }

// Gen changes whenever what is pending does.
func (e Engine) Gen() int { return e.gen }

// Reset forgets a count or sequence under way.
func (e *Engine) Reset() {
	if e.pending != "" || e.count > 0 {
		e.gen++
	}
	e.pending, e.count = "", 0
}

// SetMode switches mode and forgets anything pending.
func (e *Engine) SetMode(m mode.Mode) {
	e.Reset()
	e.Mode = m
}

// Hints is what can follow the keys typed so far, for the which-key box.
func (e Engine) Hints(where string) []Hint {
	if e.pending == "" {
		return nil
	}
	return e.Map.Next(e.pending, e.Mode, where)
}

// Feed takes one key, as Bubble Tea spells it (tea.KeyMsg.String()), in
// a view ("" for a tool without views).
func (e *Engine) Feed(key, where string) Result {
	t := Token(key)
	if e.Mode.Typing() {
		if b, _ := e.Map.lookup(t, e.Mode, where); b != nil {
			return e.fire(b, t)
		}
		return Result{}
	}
	if t == "Esc" && e.Waiting() {
		e.Reset()
		return Result{Cancel: true, Handled: true, Gen: e.gen}
	}
	if t == "Backspace" && e.pending != "" {
		// up a level: "Space m" → "Space"; the last one cancels
		if i := strings.LastIndex(e.pending, " "); i >= 0 {
			e.pending = e.pending[:i]
			e.gen++
			return Result{Pending: true, Back: true, Gen: e.gen, Handled: true}
		}
		e.Reset()
		return Result{Cancel: true, Back: true, Handled: true, Gen: e.gen}
	}
	if e.pending == "" && len(t) == 1 && t[0] >= '0' && t[0] <= '9' && (t != "0" || e.count > 0) {
		// Once a count has started, every digit is part of it (72j), even
		// in a view that binds a digit itself.
		if b, longer := e.Map.lookup(t, e.Mode, where); e.count > 0 || b == nil && !longer {
			if e.count < 99999 {
				e.count = e.count*10 + int(t[0]-'0')
			}
			e.gen++
			return Result{Pending: true, Gen: e.gen, Handled: true}
		}
	}
	seq := t
	if e.pending != "" {
		seq = e.pending + " " + t
	}
	b, longer := e.Map.lookup(seq, e.Mode, where)
	if longer {
		e.pending = seq
		e.gen++
		return Result{Pending: true, Gen: e.gen, Handled: true}
	}
	if b == nil {
		spelled := e.Pending()
		if spelled != "" {
			spelled += " "
		}
		spelled += t
		e.Reset()
		return Result{Unknown: spelled, Handled: true, Gen: e.gen}
	}
	return e.fire(b, seq)
}

func (e *Engine) fire(b *Binding, seq string) Result {
	r := Result{Act: b.Act, Count: max(1, e.count), Counted: e.count > 0, Handled: true}
	if b.Moved != "" {
		r = Result{Unknown: seq, Moved: b.Moved, Handled: true}
	}
	e.Reset()
	r.Gen = e.gen
	return r
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// Message is what to tell the user about a result, if anything: a moved
// key, or a sequence that means nothing here. A lone unknown key says
// nothing (stray letters in NORMAL are harmless).
func (r Result) Message() string {
	switch {
	case r.Moved != "":
		return r.Unknown + " moved to " + r.Moved
	case strings.Contains(r.Unknown, " "):
		return r.Unknown + " isn't a key here (Space ? lists them)"
	}
	return ""
}
