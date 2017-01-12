package engine

import (
	"sync"

	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrInvalidMove = errors.New("invalid move")
)

// Common defaults.
const (
	DefaultCols     = 7
	DefaultRows     = 6
	DefaultNPlayers = 2
	DefaultNWin     = 4
	DefaultMode     = "terminal"
)

// Four holds the game state.
type Four struct {
	sync.RWMutex // Lock to protect the Players map.

	Content  [][]State `json:"content"`
	NWin     int       `json:"nwin"`
	NPlayers int       `json:"nplayers"`

	Columns      int   `json:"columns"`
	Rows         int   `json:"rows"`
	CurPlayerIdx int   `json:"cur_player_idx"`
	CurPlayer    State `json:"cur_player"`

	AvailablePlayers []State          `json:"available_players"`
	Players          map[State]string `json:"players"` // Used for the server mode.

	ActivityChan chan State `json:"-"`          // Channel populated with latest game state.
	GridState    State      `json:"grid_state"` // If not "Empty", then the game is finished.
}

// NewConnectFour instantiates a new game.
// TODO: Use a single dimension slice.
func NewConnectFour(columns, rows, nPlayers, nWin int) (*Four, error) {
	if columns < 2 || rows < 2 {
		return nil, errors.Errorf("invalid grid size: %d/%d", columns, rows)
	}
	if nPlayers > len(AvailablePlayers) {
		return nil, errors.Errorf("too many players. Max: %d", len(AvailablePlayers))
	}
	if nWin < 2 {
		return nil, errors.Errorf("invalid win number: %d, minimum 2", nWin)
	}
	if columns < nWin && rows < nWin {
		return nil, errors.New("grid too small for anyone to win")
	}
	content := make([][]State, rows)
	for i := range content {
		content[i] = make([]State, columns)
	}
	return &Four{
		Content:          content,
		NWin:             nWin,
		NPlayers:         nPlayers,
		Columns:          columns,
		Rows:             rows,
		AvailablePlayers: AvailablePlayers[:nPlayers],
		CurPlayer:        AvailablePlayers[0],
		Players:          map[State]string{},
		GridState:        Empty,
		ActivityChan:     make(chan State, 1e3), // Arbitrary large size.
	}, nil
}

// Reset restarts the game.
func (f *Four) Reset() (*Four, error) {
	return NewConnectFour(f.Columns, f.Rows, f.NPlayers, f.NWin)
}

// State return the state of the grid at the x/y position.
func (f *Four) State(x, y int) State {
	return f.Content[x][y]
}

// ColumnCount returns the count of occupied cells in the requested column.
func (f *Four) ColumnCount(col int) int {
	for i, elem := range f.Content {
		if elem[col] != Empty {
			return i
		}
	}
	return 0
}

// ValidateMove checks if the given move is valid.
func (f *Four) ValidateMove(player State, col int) error {
	// Validate move.
	// Check if expected player.
	if player != f.CurPlayer {
		return errors.New("invalid move, not player's turn")
	}
	// Les than 0, too big or column already full.
	if col < 0 || col >= len(f.Content[0]) || f.Content[0][col] != Empty {
		return errors.Wrapf(ErrInvalidMove, "%d is an invalid move for player %d (%s)", col, f.CurPlayer, f.CurPlayer)
	}
	return nil
}

// PlayerMove plays a move for the given player.
func (f *Four) PlayerMove(player State, col int) (State, error) {
	if err := f.ValidateMove(player, col); err != nil {
		return Empty, err
	}
	// Make it fall as long as we are empty.
	j := 0
	for ; j < len(f.Content); j++ {
		if f.Content[j][col] != Empty {
			break
		}
	}
	// Set the state in the internal grid.
	f.Content[j-1][col] = f.CurPlayer

	// Update Current Player.
	f.CurPlayerIdx++
	f.CurPlayerIdx %= f.NPlayers
	f.CurPlayer = f.AvailablePlayers[f.CurPlayerIdx]

	return f.Compute(), nil
}

// compute goes point by point and tries the nWin in every directions.
func (f *Four) compute() State {
	for i := 0; i < len(f.Content); i++ {
		for j := 0; j < len(f.Content[i]); j++ {
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
	for _, elem := range f.AvailablePlayers {
		playerState[elem] = 0
	}

	// Initialize previous state as the original.
	prev := f.Content[x][y]

	// Check f.nWin cells.
	for i := 0; i < f.NWin; i++ {
		// Check for boundaries, if outside, then return.
		if x+i*xDir < 0 || x+i*xDir >= len(f.Content) ||
			y+i*yDir < 0 || y+i*yDir >= len(f.Content[0]) {
			return Empty
		}

		// Check the cell in the requested direciton.
		cur := f.Content[x+i*xDir][y+i*yDir]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= f.NWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// notify sends the given state to the activityChan.
// if state is not Empty, close the channel.
func (f *Four) notify(s State) {
	f.Lock()
	defer f.Unlock()

	if f.ActivityChan == nil {
		return
	}
	select {
	case f.ActivityChan <- s:
	default:
	}
	if s != Empty {
		close(f.ActivityChan)
		f.ActivityChan = nil
	}
}

// Compute processes the current state and checks if
// one of the player won.
func (f *Four) Compute() State {
	// Check all directions.
	if ret := f.compute(); ret != Empty {
		f.GridState = ret
		f.notify(ret)
		return ret
	}

	// Check if we are in a stale situation.
	if f.checkStale() {
		f.GridState = Stale
		f.notify(Stale)
		return Stale
	}
	f.notify(Empty)
	return Empty
}

// checkStale checks if the grid is complete and the game is finished.
func (f *Four) checkStale() bool {
	for _, state := range f.Content[0] {
		if state == Empty {
			return false
		}
	}
	return true
}
