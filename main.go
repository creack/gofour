package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"text/tabwriter"
	"time"

	"github.com/creack/gogrid"
	termbox "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

// FourRuntime is the interface to run connect four.
type FourRuntime interface {
	Init(*Four) error
	Run() error
	Close() error
}

// TerminalFour is a termcap player for connect four.
type TerminalFour struct {
	four *Four
	grid *gogrid.Grid
}

// Run .
func (f *TerminalFour) Run() error {
	// First draw.
	if err := f.grid.RedrawAll(); err != nil {
		return errors.Wrap(err, "error drawing grid")
	}
	// Start the runtime loop.
	if err := f.grid.HandleKeyboard(); err != nil {
		return errors.Wrap(err, "runtime error")
	}
	return nil
}

// Init initialize the termcap grid.
func (f *TerminalFour) Init(four *Four) error {
	f.four = four

	g, err := gogrid.NewGrid(f.four.Rows, f.four.Columns)
	if err != nil {
		return errors.Wrap(err, "error initializing termcap grid")
	}

	cursorX := 0
	end := false

	g.HeaderHeight = 2
	g.HeaderFct = func(g *gogrid.Grid) {
		if !end {
			// Display player info.
			fmt.Printf("Player %d (%s) turn, select column (Enter or Space)\n", f.four.CurPlayer, f.four.CurPlayer)
			// Set cursor to proper cell.
			g.SetCursor(cursorX, 0)
		}
	}
	leftHandler := func(*gogrid.Grid) {
		if cursorX > 0 {
			cursorX--
		}
	}
	rightHandler := func(*gogrid.Grid) {
		if cursorX < g.Width-1 {
			cursorX++
		}
	}
	toggleHandler := func(*gogrid.Grid) {
		if end {
			return
		}

		g.ClearHeader()
		ret, err := f.four.PlayerMove(cursorX)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			g.SetCursor(0, 0)
			return
		}

		// Make it fall as long as we are empty.
		j := 0
		for ; j < f.four.ColumnCount(cursorX); j++ {
			g.SetCursor(cursorX, j)
			fmt.Printf("%s", f.four.CurPlayer)
			g.SetCursor(cursorX, 0)

			time.Sleep(50 * time.Millisecond)

			g.SetCursor(cursorX, j)
			fmt.Print(" ")
		}
		g.SetCursor(cursorX, j)
		fmt.Printf("%s", f.four.CurPlayer)

		if ret != Empty {
			g.ClearHeader()
			if ret == Stale {
				fmt.Print("\n Stale, nobody wins! (ESC to exit)")
			} else {
				fmt.Printf("\n Player %d (%s) won! (ESC to exit)", ret, ret)
			}
			g.SetCursor(0, 0)
			end = true
			return
		}
	}

	g.RegisterKeyHandler(termbox.KeyArrowLeft, leftHandler)
	g.RegisterKeyHandler(termbox.KeyCtrlB, leftHandler)
	g.RegisterKeyHandler(termbox.KeyArrowRight, rightHandler)
	g.RegisterKeyHandler(termbox.KeyCtrlF, rightHandler)
	g.RegisterKeyHandler(termbox.KeySpace, toggleHandler)
	g.RegisterKeyHandler(termbox.KeyEnter, toggleHandler)
	g.RegisterKeyHandler('q', func(g *gogrid.Grid) {
		_ = g.Close()
	})
	g.RegisterKeyHandler(termbox.KeyCtrlL, func(g *gogrid.Grid) {
		_ = g.RedrawAll()
	})
	f.grid = g
	return nil
}

// Close cleans up the grid and terminal.
func (f *TerminalFour) Close() error {
	return f.grid.Close()
}

func main() {
	var (
		cols     = flag.Int("cols", 7, "number of columns")
		rows     = flag.Int("rows", 6, "number of rows")
		nPlayers = flag.Int("p", 2, fmt.Sprintf("number of players. (max: %d)", len(AvailablePlayers)))
		nWin     = flag.Int("w", 4, "number of consecutive color to win")
		mode     = flag.String("m", "term", "Game mode. Values: [term, text]")
	)
	flag.Parse()

	four, err := NewConnectFour(*cols, *rows, *nPlayers, *nWin)
	if err != nil {
		log.Fatal(err)
	}

	var game FourRuntime
	if *mode == "term" {
		game = &TerminalFour{}
	} else {
		game = &TextFour{}
	}
	if err := game.Init(four); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = game.Close() }()

	if err := game.Run(); err != nil {
		log.Fatal(err)
	}
}

// Dump displays the state of the grid on stdout.
func Dump(f *Four) {
	fmt.Println()
	w := tabwriter.NewWriter(os.Stdout, 4, 4, 4, ' ', 0)
	for i := 0; i < f.Columns; i++ {
		fmt.Fprintf(w, "\x1b[1;37m%d\x1b[0m\t", i+1)
	}
	fmt.Fprintln(w)
	for i := 0; i < f.Columns; i++ {
		fmt.Fprint(w, "\x1b[1;37m---\x1b[0m\t")
	}
	fmt.Fprintln(w)
	for i := 0; i < f.Rows; i++ {
		for j := 0; j < f.Columns; j++ {
			fmt.Fprintf(w, "%s\t", f.State(i, j))
		}
		fmt.Fprintf(w, "\n")
	}
	w.Flush()
	fmt.Println()
}

// TextFour is a basic text based client for connect four.
type TextFour struct {
	four *Four
}

// Init setups up the connect four game.
func (tf *TextFour) Init(four *Four) error {
	tf.four = four
	return nil
}

// Run starts the game loop.
func (tf *TextFour) Run() error {
	for i := 0; ; i++ {
	start:
		Dump(tf.four)
		fmt.Printf("Player %d (%s) turn, select column:\n", tf.four.CurPlayer, tf.four.CurPlayer)

		var x int
		fmt.Scanf("%d", &x)
		x-- // Back to 0 index.

		if x < 0 {
			fmt.Fprint(os.Stderr, "invalid columns number\n")
		}

		ret, err := tf.four.PlayerMove(x)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			if err := errors.Cause(err); err == ErrInvalidMove {
				goto start
			}
			return err
		}
		if ret != Empty {
			if ret == Stale {
				fmt.Print("Stale, nobody wins!\n")
			} else {
				fmt.Printf("Player %d (%s) won!\n", ret, ret)
			}
			break
		}
	}
	return nil
}

// Close is a no op.
func (tf *TextFour) Close() error {
	return nil
}
