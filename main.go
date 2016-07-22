package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
)

// GridState is the enum type for the grid state.
type GridState int

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
)

// AvailablePlayers is the list of available player colors.
var AvailablePlayers = []GridState{Red, Yellow, Green, Magenta, Blue, Cyan, Black}

func (s GridState) String() string {
	switch s {
	case Red:
		return fmt.Sprintf("\x1b[1;31m%c\x1b[0m", 'x')
	case Yellow:
		return fmt.Sprintf("\x1b[1;33m%c\x1b[0m", 'x')
	case Green:
		return fmt.Sprintf("\x1b[1;32m%c\x1b[0m", 'x')
	case Blue:
		return fmt.Sprintf("\x1b[1;34m%c\x1b[0m", 'x')
	case Magenta:
		return fmt.Sprintf("\x1b[1;35m%c\x1b[0m", 'x')
	case Cyan:
		return fmt.Sprintf("\x1b[1;36m%c\x1b[0m", 'x')
	case Black:
		return fmt.Sprintf("\x1b[1;30m%c\x1b[0m", 'x')
	default:
		return fmt.Sprintf("\x1b[1;37m%c\x1b[0m", 'x')
	}
}

// initPlayerState is a helper to init the map for the win check.
// TODO: get tid of this.
func initPlayerState() map[GridState]int {
	ret := map[GridState]int{
		Empty: 0,
	}
	for _, elem := range AvailablePlayers {
		ret[elem] = 0
	}
	return ret
}

// Grid holds the game state.
type Grid struct {
	content  [][]GridState
	nWin     int
	nPlayers int
}

// NewGrid instantiates a new Grid.
// TODO: use a single dimentional slice.
func NewGrid(columns, rows, nPlayers, nWin int) (*Grid, error) {
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
	content := make([][]GridState, columns)
	for i := range content {
		content[i] = make([]GridState, rows)
	}
	return &Grid{
		content:  content,
		nWin:     nWin,
		nPlayers: nPlayers,
	}, nil
}

// Dump displays the state of the grid on stdout.
func (g *Grid) Dump() {
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	for i := range g.content[0] {
		fmt.Fprintf(w, "\x1b[1;37m%d\x1b[0m\t", i+1)
	}
	fmt.Fprintln(w)
	for range g.content[0] {
		fmt.Fprint(w, "\x1b[1;37m---\x1b[0m\t")
	}
	fmt.Fprintln(w)
	for _, lines := range g.content {
		for _, state := range lines {
			fmt.Fprintf(w, "%s\t", state)
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
	fmt.Println()
}

// computeLines checks for nWin in a row in the lines.
func (g *Grid) computeLines() GridState {
	for _, line := range g.content {
		playerState := initPlayerState()
		prev := line[0]

		for i := 0; i < len(line); i++ {
			cur := line[i]
			if prev != cur {
				playerState[prev] = 0
				playerState[cur] = 1
			} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
				return cur
			}
			prev = cur
		}
	}
	return Empty
}

// computeColumns checks for nWin in a row in the columns.
func (g *Grid) computeColumns() GridState {
	for i := 0; i < len(g.content[0]); i++ {
		playerState := initPlayerState()
		prev := g.content[0][i]
		for j := 0; j < len(g.content); j++ {
			cur := g.content[j][i]
			if prev != cur {
				playerState[prev] = 0
				playerState[cur] = 1
			} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
				return cur
			}
			prev = cur
		}
	}
	return Empty
}

// checkDiag1 checks from the starting x/y if we have nWin in a row.
func (g *Grid) checkDiag1(x, y int) GridState {
	playerState := initPlayerState()
	prev := g.content[x][y]
	for i := 0; i < g.nWin; i++ {
		if x+i >= len(g.content) || y+i >= len(g.content[0]) {
			return Empty
		}
		cur := g.content[x+i][y+i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag2 checks from the starting x/y if we have nWin in a row.
func (g *Grid) checkDiag2(x, y int) GridState {
	playerState := initPlayerState()
	prev := g.content[x][y]
	for i := 0; i < g.nWin; i++ {
		if x-i < 0 || y-i < 0 {
			return Empty
		}
		cur := g.content[x-i][y-i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag3 checks from the starting x/y if we have nWin in a row.
func (g *Grid) checkDiag3(x, y int) GridState {
	playerState := initPlayerState()
	prev := g.content[x][y]
	for i := 0; i < g.nWin; i++ {
		if x-i < 0 || y+i >= len(g.content[0]) {
			return Empty
		}
		cur := g.content[x-i][y+i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// checkDiag4 checks from the starting x/y if we have nWin in a row.
func (g *Grid) checkDiag4(x, y int) GridState {
	playerState := initPlayerState()
	prev := g.content[x][y]
	for i := 0; i < g.nWin; i++ {
		if x+i >= len(g.content) || y-i < 0 {
			return Empty
		}
		cur := g.content[x+i][y-i]
		if prev != cur {
			playerState[prev] = 0
			playerState[cur] = 1
		} else if playerState[cur]++; cur != Empty && playerState[cur] >= g.nWin {
			return cur
		}
		prev = cur
	}
	return Empty
}

// computeDiags goes point by point and tries the nWin diagonals.
func (g *Grid) computeDiags() GridState {
	for i := 0; i < len(g.content); i++ {
		for j := 0; j < len(g.content[i]); j++ {
			if ret := g.checkDiag1(i, j); ret != Empty {
				return ret
			}
			if ret := g.checkDiag2(i, j); ret != Empty {
				return ret
			}
			if ret := g.checkDiag3(i, j); ret != Empty {
				return ret
			}
			if ret := g.checkDiag4(i, j); ret != Empty {
				return ret
			}
		}
	}
	return Empty
}

// Compute processes the current state and checks if
// one of the player won.
func (g *Grid) Compute() GridState {
	// nWin in a line.
	if ret := g.computeLines(); ret != Empty {
		return ret
	}
	// nWin in a column.
	if ret := g.computeColumns(); ret != Empty {
		return ret
	}

	// nWin in a diagonal.
	if ret := g.computeDiags(); ret != Empty {
		return ret
	}

	return Empty
}

// Stale checks if the grid is complete and the game is finished.
func (g *Grid) Stale() bool {
	for _, state := range g.content[0] {
		if state == Empty {
			return false
		}
	}
	return true
}

// Run starts the game.
func (g *Grid) Run() {
	playerState := AvailablePlayers[:g.nPlayers]
	for i := 0; ; i++ {
	start:
		g.Dump()
		fmt.Printf("Player %d (%s) turn, select column:\n", playerState[i%g.nPlayers], playerState[i%g.nPlayers])
		var x int
		fmt.Scanf("%d", &x)

		// Validate user input.
		// Les than 0, too big or column already full.
		if x <= 0 || x >= len(g.content[0])+1 || g.content[0][x-1] != 0 {
			fmt.Fprint(os.Stderr, "Invalid input\n")
			goto start
		}

		x-- // Back to 0 index.

		// Make it fall as long as we are empty.
		j := 0
		for ; j < len(g.content); j++ {
			if g.content[j][x] != Empty {
				break
			}
		}
		g.content[j-1][x] = playerState[i%g.nPlayers]
		if ret := g.Compute(); ret != Empty {
			g.Dump()
			fmt.Printf("Player %d (%s) won!\n", ret, ret)
			return
		}
		if g.Stale() {
			g.Dump()
			fmt.Print("Stale, nobody wins!\n")
			return
		}
	}
}

func main() {
	var (
		cols     = flag.Int("cols", 7, "number of columns")
		rows     = flag.Int("rows", 6, "number of rows")
		nPlayers = flag.Int("p", 2, fmt.Sprintf("number of players. (max: %d)", len(AvailablePlayers)))
		nWin     = flag.Int("w", 4, "number of consecutive color to win")
	)
	flag.Parse()
	g, err := NewGrid(*cols, *rows, *nPlayers, *nWin)
	if err != nil {
		log.Fatal(err)
	}
	g.Run()
}
