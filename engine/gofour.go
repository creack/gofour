package engine

import (
	"fmt"

	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrInvalidMove = errors.New("invalid move")
)

// initPlayerState is a helper to init the map for the win check.
// TODO: get rid of this.
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
		CurPlayer:        AvailablePlayers[0],
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
		return errors.Wrapf(ErrInvalidMove, "%d is an invalid move for player %d (%s)", col, f.CurPlayer, f.CurPlayer)
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
	// Set the state in the internal grid.
	f.content[j-1][col] = f.CurPlayer

	// Update Current Player.
	f.CurPlayerIdx++
	f.CurPlayerIdx %= f.nPlayers
	f.CurPlayer = f.availablePlayers[f.CurPlayerIdx]

	return f.Compute(), nil
}

// // Run starts the game with the given runtime.
// func (f *Four) Run(runtime runtime.FourRuntime) error {
// 	if err := runtime.Init(f); err != nil {
// 		return errors.Wrap(err, "error initializing the runtime")
// 	}
// 	defer func() { _ = runtime.Close() }()

// 	return runtime.Run()
// }

// compute goes point by point and tries the nWin in every directions.
func (f *Four) compute() State {
	for i := 0; i < len(f.content); i++ {
		for j := 0; j < len(f.content[i]); j++ {
			// Check Columns.
			if ret := f.checkDirection(i, j, 1, 0); ret != Empty {
				return ret
			}
			// Check lines.
			if ret := f.checkDirection(i, j, 0, 1); ret != Empty {
				return ret
			}

			// Check diagonals.
			if ret := f.checkDirection(i, j, 1, 1); ret != Empty {
				return ret
			}
			if ret := f.checkDirection(i, j, -1, 1); ret != Empty {
				return ret
			}
		}
	}
	return Empty
}

// checkDirection checks from the starting x/y if we have nWin in a row in the requested direction.
// xDir and yDir should be either 0, 1 or -1.
func (f *Four) checkDirection(x, y, xDir, yDir int) State {
	// Small map used as a helper to count.
	playerState := map[State]int{
		Empty: 0,
	}
	for _, elem := range f.availablePlayers {
		playerState[elem] = 0
	}

	// Initialize previous state as the original.
	prev := f.content[x][y]

	// Check f.nWin cells.
	for i := 0; i < f.nWin; i++ {
		// Check for boundaries, if outside, then return.
		if x+i*xDir < 0 || x+i*xDir >= len(f.content) ||
			y+i*yDir < 0 || y+i*yDir >= len(f.content[0]) {
			return Empty
		}

		// Check the cell in the requested direciton.
		cur := f.content[x+i*xDir][y+i*yDir]
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

// Compute processes the current state and checks if
// one of the player won.
func (f *Four) Compute() State {
	// Check all directions.
	if ret := f.compute(); ret != Empty {
		return ret
	}

	// Check if we are in a stale situation.
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
