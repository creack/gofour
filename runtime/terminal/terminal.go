package terminal

import (
	"fmt"
	"os"
	"time"

	"github.com/creack/gofour/engine"
	"github.com/creack/gofour/runtime"
	"github.com/creack/gogrid"
	termbox "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

func init() {
	runtime.Runtimes["terminal"] = &Runtime{}
}

// Runtime is a termcap player for connect four.
type Runtime struct {
	four *engine.Four
	grid *gogrid.Grid

	cursorX int  // Current cursor.
	end     bool // Flag for end of game.
}

// Run starts the runtime.
func (tf *Runtime) Run() error {
	// First draw.
	if err := tf.grid.RedrawAll(); err != nil {
		return errors.Wrap(err, "error drawing grid")
	}
	// Start the runtime loop.
	if err := tf.grid.HandleKeyboard(); err != nil {
		return errors.Wrap(err, "runtime error")
	}
	return nil
}

// HeaderHandler displays info in the header section of the grid.
func (tf *Runtime) HeaderHandler(g *gogrid.Grid) {
	if !tf.end {
		// Display player info.
		fmt.Printf("Player %d (%s) turn, select column (Enter or Space)\n", tf.four.CurPlayer, tf.four.CurPlayer)
		// Set cursor to proper cell.
		g.SetCursor(tf.cursorX, 0)
	}
}

func (tf *Runtime) leftKeyHandler(*gogrid.Grid) {
	if tf.cursorX > 0 {
		tf.cursorX--
	}
}

func (tf *Runtime) rightKeyHandler(g *gogrid.Grid) {
	if tf.cursorX < g.Width-1 {
		tf.cursorX++
	}
}

func (tf *Runtime) toggleHandler(g *gogrid.Grid) {
	if tf.end {
		return
	}

	g.ClearHeader()
	curPlayer := tf.four.CurPlayer
	ret, err := tf.four.PlayerMove(tf.four.CurPlayer, tf.cursorX)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		g.SetCursor(0, 0)
		return
	}

	// Make it fall as long as we are empty.
	j := 0
	for ; j < tf.four.ColumnCount(tf.cursorX); j++ {
		g.SetCursor(tf.cursorX, j)
		fmt.Printf("%s", curPlayer)
		g.SetCursor(tf.cursorX, 0)

		time.Sleep(50 * time.Millisecond)

		g.SetCursor(tf.cursorX, j)
		fmt.Print(" ")
	}
	g.SetCursor(tf.cursorX, j)
	fmt.Printf("%s", curPlayer)

	if ret != engine.Empty {
		g.ClearHeader()
		if ret == engine.Stale {
			fmt.Print("\n Stale, nobody wins! (ESC to exit)")
		} else {
			fmt.Printf("\n Player %d (%s) won! (ESC to exit)", ret, ret)
		}
		g.SetCursor(0, 0)
		tf.end = true
		return
	}
}

// Init initialize the termcap grid.
func (tf *Runtime) Init(four *engine.Four) error {
	tf.four = four

	// Initialize new termbox grid.
	g, err := gogrid.NewGrid(tf.four.Rows, tf.four.Columns)
	if err != nil {
		return errors.Wrap(err, "error initializing termcap grid")
	}
	tf.grid = g

	// Setup header.
	g.HeaderHeight = 2
	g.HeaderFct = tf.HeaderHandler

	// Register the key handlers.
	g.RegisterKeyHandler(termbox.KeyArrowLeft, tf.leftKeyHandler)
	g.RegisterKeyHandler(termbox.KeyCtrlB, tf.leftKeyHandler)
	g.RegisterKeyHandler(termbox.KeyArrowRight, tf.rightKeyHandler)
	g.RegisterKeyHandler(termbox.KeyCtrlF, tf.rightKeyHandler)
	g.RegisterKeyHandler(termbox.KeySpace, tf.toggleHandler)
	g.RegisterKeyHandler(termbox.KeyEnter, tf.toggleHandler)
	g.RegisterKeyHandler('q', func(g *gogrid.Grid) { _ = g.Close() })
	g.RegisterKeyHandler(termbox.KeyCtrlL, func(g *gogrid.Grid) { _ = g.RedrawAll() })

	return nil
}

// Close cleans up the grid and terminal.
func (tf *Runtime) Close() error {
	return tf.grid.Close()
}
