package engine

import "fmt"

// State is the enum type for the grid state.
type State int

// GridState is the enum values
const (
	Empty = iota
	Red
	Yellow
	Green
	Magenta
	Blue
	Cyan
	Black
	Stale

	PlayerUnicode = '●'
)

// AvailablePlayers is the list of available player colors.
var AvailablePlayers = []State{Red, Yellow, Green, Magenta, Blue, Cyan, Black}

func (s State) String() string {
	switch s {
	case Red:
		return fmt.Sprintf("\x1b[1;31m%c\x1b[0m", PlayerUnicode)
	case Yellow:
		return fmt.Sprintf("\x1b[1;33m%c\x1b[0m", PlayerUnicode)
	case Green:
		return fmt.Sprintf("\x1b[1;32m%c\x1b[0m", PlayerUnicode)
	case Blue:
		return fmt.Sprintf("\x1b[1;34m%c\x1b[0m", PlayerUnicode)
	case Magenta:
		return fmt.Sprintf("\x1b[1;35m%c\x1b[0m", PlayerUnicode)
	case Cyan:
		return fmt.Sprintf("\x1b[1;36m%c\x1b[0m", PlayerUnicode)
	case Black:
		return fmt.Sprintf("\x1b[1;30m%c\x1b[0m", PlayerUnicode)
	default:
		return fmt.Sprintf("\x1b[1;37m%c\x1b[0m", PlayerUnicode)
	}
}
