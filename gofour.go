package main

import (
	"fmt"

	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrInvalidMove = errors.New("invalid move")
)

// initPlayerState is a helper to init the map for the win check.
// TODO: get tid of this.
func initPlayerState() map[State]int {
	ret := map[State]int{
		Empty: 0,
	}
	for _, elem := range AvailablePlayers {
		ret[elem] = 0
	}
	return ret
}

// Four holds the game state.
type Four struct {
	content  [][]State
	nWin     int
	nPlayers int

	Columns      int
	Rows         int
	CurPlayerIdx int
	CurPlayer    State

	availablePlayers []State
}

// NewConnectFour instantiates a new game.
// TODO: use a single dimentional slice.
func NewConnectFour(columns, rows, nPlayers, nWin int) (*Four, error) {
	if columns < 2 || rows < 2 {
		return nil, fmt.Errorf("invalid grid size: %d/%d", columns, rows)
	}
	if nPlayers > len(AvailablePlayers) {
		return nil, fmt.Errorf("too many players. Max: %d", len(AvailablePlayers))
	}
	if nWin < 2 {
		return nil, fmt.Errorf("invalid win number: %d, minimum 2", nWin)
	}
	if columns < nWin && rows < nWin {
		return nil, fmt.Errorf("grid too small for anyone to win")
	}
	content := make([][]State, rows)
	for i := range content {
		content[i] = make([]State, columns)
	}
	return &Four{
		content:          content,
		nWin:             nWin,
		nPlayers:         nPlayers,
		Columns:          columns,
		Rows:             rows,
		availablePlayers: AvailablePlayers[:nPlayers],
	}, nil
}

// State return the state of the grid at the x/y position.
func (f *Four) State(x, y int) State {
	return f.content[x][y]
}

// ColumnCount returns the count of occupied cells in the requested column.
func (f *Four) ColumnCount(col int) int {
	for i, elem := range f.content {
		if elem[col] != Empty {
			return i
		}
	}
	return 0
}

// ValidateMove checks if the given move is valid.
func (f *Four) ValidateMove(col int) error {
	// Validate move.
	// Les than 0, too big or column already full.
	if col < 0 || col >= len(f.content[0]) || f.content[0][col] != Empty {
		return errors.Wrapf(ErrInvalidMove, "%d is an invalid move for player %d (%s)", col, f.CurPlayerIdx, f.CurPlayerIdx)
	}
	return nil
}

// PlayerMove plays a move for the given player.
func (f *Four) PlayerMove(col int) (State, error) {
	if err := f.ValidateMove(col); err != nil {
		return Empty, err
	}
	// Make it fall as long as we are empty.
	j := 0
	for ; j < len(f.content); j++ {
		if f.content[j][col] != Empty {
			break
		}
	}
	f.CurPlayer = f.availablePlayers[f.CurPlayerIdx]
	f.content[j-1][col] = f.CurPlayer

	f.CurPlayerIdx++
	f.CurPlayerIdx %= f.nPlayers

	return f.Compute(), nil
}

// computeLines checks for nWin in a row in the lines.
func (f *Four) computeLines() State {
	for _, line := range f.content {
		playerState := initPlayerState()
		prev := line[0]

		for i := 0; i < len(line); i++ {
			cur := line[i]
			if prev != cur {
				playerState[prev] = 0
				playerState[cur] = 1
			} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
				return cur
			}
			prev = cur
		}
	}
	return Empty
}

// computeColumns checks for nWin in a row in the columns.
func (f *Four) computeColumns() State {
	for i := 0; i < len(f.content[0]); i++ {
		playerState := initPlayerState()
		prev := f.content[0][i]
		for j := 0; j < len(f.content); j++ {
			cur := f.content[j][i]
			if prev != cur {
				playerState[prev] = 0
				playerState[cur] = 1
			} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
				return cur
			}
			prev = cur
		}
	}
	return Empty
}

// checkDiag1 checks from the starting x/y if we have nWin in a row.
func (f *Four) checkDiag1(x, y int) State {
	playerState := initPlayerState()
	prev := f.content[x][y]
	for i := 0; i < f.nWin; i++ {
		if x+i >= len(f.content) || y+i >= len(f.content[0]) {
			return Empty
		}
		cur := f.content[x+i][y+i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag2 checks from the starting x/y if we have nWin in a row.
func (f *Four) checkDiag2(x, y int) State {
	playerState := initPlayerState()
	prev := f.content[x][y]
	for i := 0; i < f.nWin; i++ {
		if x-i < 0 || y-i < 0 {
			return Empty
		}
		cur := f.content[x-i][y-i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag3 checks from the starting x/y if we have nWin in a row.
func (f *Four) checkDiag3(x, y int) State {
	playerState := initPlayerState()
	prev := f.content[x][y]
	for i := 0; i < f.nWin; i++ {
		if x-i < 0 || y+i >= len(f.content[0]) {
			return Empty
		}
		cur := f.content[x-i][y+i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag4 checks from the starting x/y if we have nWin in a row.
func (f *Four) checkDiag4(x, y int) State {
	playerState := initPlayerState()
	prev := f.content[x][y]
	for i := 0; i < f.nWin; i++ {
		if x+i >= len(f.content) || y-i < 0 {
			return Empty
		}
		cur := f.content[x+i][y-i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// computeDiags goes point by point and tries the nWin diagonals.
func (f *Four) computeDiags() State {
	for i := 0; i < len(f.content); i++ {
		for j := 0; j < len(f.content[i]); j++ {
			if ret := f.checkDiag1(i, j); ret != Empty {
				return ret
			}
			if ret := f.checkDiag2(i, j); ret != Empty {
				return ret
			}
			if ret := f.checkDiag3(i, j); ret != Empty {
				return ret
			}
			if ret := f.checkDiag4(i, j); ret != Empty {
				return ret
			}
		}
	}
	return Empty
}

// Compute processes the current state and checks if
// one of the player won.
func (f *Four) Compute() State {
	// nWin in a line.
	if ret := f.computeLines(); ret != Empty {
		return ret
	}
	// nWin in a column.
	if ret := f.computeColumns(); ret != Empty {
		return ret
	}

	// nWin in a diagonal.
	if ret := f.computeDiags(); ret != Empty {
		return ret
	}

	if f.checkStale() {
		return Stale
	}

	return Empty
}

// checkStale checks if the grid is complete and the game is finished.
func (f *Four) checkStale() bool {
	for _, state := range f.content[0] {
		if state == Empty {
			return false
		}
	}
	return true
}
