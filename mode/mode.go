// Package mode is the modes every Vayu tool shares: NORMAL, where keys are
// commands and typing changes nothing; INSERT, where typing goes into the
// field; VISUAL, where you mark many rows; COMMAND, the ":" line.
package mode

// Mode is where the keys go.
type Mode int

const (
	Normal Mode = iota
	Insert
	Visual
	Command
)

// String is the mode's name as the pill in the footer shows it.
func (m Mode) String() string {
	switch m {
	case Insert:
		return "INSERT"
	case Visual:
		return "VISUAL"
	case Command:
		return "COMMAND"
	}
	return "NORMAL"
}

// Lower is the mode's name for the keybinding search's scope
// ("pattan ports:normal").
func (m Mode) Lower() string {
	switch m {
	case Insert:
		return "insert"
	case Visual:
		return "visual"
	case Command:
		return "command"
	}
	return "normal"
}

// Typing says whether keys go into a text field in this mode.
func (m Mode) Typing() bool { return m == Insert || m == Command }
