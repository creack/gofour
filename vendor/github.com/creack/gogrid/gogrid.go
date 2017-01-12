package gogrid

import (
	"fmt"
	"strings"

	termbox "github.com/nsf/termbox-go"
	"github.com/pkg/errors"
)

// Common errors.
var (
	ErrTerminalTooSmall = errors.New("terminal too small")
	ErrKeyInUse         = errors.New("key already registered")
)

// KeyHandler associates a key with a callback.
type KeyHandler struct {
	Key termbox.Key
	Fct func(*Grid)
}

// Grid holds the data for the terminal grid.
//
// Sample output for a 2 rows 3 columns grid:
//  ┌─┬─┬─┐
//  │ │ │ │
//  ├─┼─┼─┤
//  │ │ │ │
//  └─┴─┴─┘
type Grid struct {
	// Dimensions.
	Height int
	Width  int

	// Offset at the top of the screen.
	HeaderHeight int
	HeaderFct    func(*Grid)

	// Internal controls.
	StopChan    <-chan struct{}
	stopChan    chan struct{}
	keyHandlers []*KeyHandler
}

// NewGrid initialize the terminal and the grid.
// Should be closed by the caller.
func NewGrid(h, w int) (*Grid, error) {
	if err := termbox.Init(); err != nil {
		return nil, errors.Wrap(err, "error initializing the terminal")
	}

	ch := make(chan struct{})
	return &Grid{
		Height:   h,
		Width:    w,
		stopChan: ch,
		StopChan: ch,
	}, nil
}

// RedrawAll redraws the whole grid.
func (g *Grid) RedrawAll() error {
	const defCol = termbox.ColorDefault

	termbox.SetInputMode(termbox.InputEsc)

	// Clear terminal.
	if err := termbox.Clear(defCol, defCol); err != nil {
		return errors.Wrap(err, "error clearing the terminal")
	}

	// Get the current terminal dimensions.
	w, h := termbox.Size()

	// Make sure we have enough space.
	if g.Width*2+1 > w || g.Height*2+1 > h-g.HeaderHeight {
		return errors.Wrap(ErrTerminalTooSmall, "not enough space to draw the grid")
	}

	// Draw first line.
	termbox.SetCell(0, 0+g.HeaderHeight, '┌', defCol, defCol)
	j := 1
	for ; j < g.Width*2; j++ {
		if j%2 == 1 {
			termbox.SetCell(j, 0+g.HeaderHeight, '─', defCol, defCol)
		} else {
			termbox.SetCell(j, 0+g.HeaderHeight, '┬', defCol, defCol)
		}
	}
	termbox.SetCell(j, 0+g.HeaderHeight, '┐', defCol, defCol)
	for j = 0; j <= g.Width*2; j += 2 {
		termbox.SetCell(j, 1+g.HeaderHeight, '│', defCol, defCol)
	}

	// Draw grid.
	i := g.HeaderHeight
	for ; i < g.Height*2; i += 2 {
		// Left border.
		termbox.SetCell(0, i+g.HeaderHeight, '├', defCol, defCol)
		termbox.SetCell(0, i+1+g.HeaderHeight, '│', defCol, defCol)

		// Inner grid.
		j := 1
		for ; j < g.Width*2; j++ {
			if j%2 == 1 {
				termbox.SetCell(j, i+g.HeaderHeight, '─', defCol, defCol)
			} else {
				termbox.SetCell(j, i+g.HeaderHeight, '┼', defCol, defCol)
			}
		}
		for j = 2; j < g.Width*2; j += 2 {
			termbox.SetCell(j, i+1+g.HeaderHeight, '│', defCol, defCol)
		}

		// Right border.
		termbox.SetCell(j, i+g.HeaderHeight, '┤', defCol, defCol)
		termbox.SetCell(j, i+1+g.HeaderHeight, '│', defCol, defCol)
	}

	// Draw last line.
	termbox.SetCell(0, i+g.HeaderHeight, '└', defCol, defCol)
	for j = 1; j < g.Width*2; j++ {
		if j%2 == 1 {
			termbox.SetCell(j, i+g.HeaderHeight, '─', defCol, defCol)
		} else {
			termbox.SetCell(j, i+g.HeaderHeight, '┴', defCol, defCol)
		}
	}
	termbox.SetCell(j, i+g.HeaderHeight, '┘', defCol, defCol)

	// Flush.
	_ = termbox.Flush()
	return nil
}

// SetCursor sets the cursor at the proper place and flushes the terminal.
// x and y are Cells, not absolute.
func (g *Grid) SetCursor(x, y int) {
	termbox.SetCursor(1+x*2, g.HeaderHeight+1+2*y)
	_ = termbox.Flush()
}

// RegisterKeyHandler adds a handler for the given key.
// Returns a pointer to the handler, needed to unregister.
func (g *Grid) RegisterKeyHandler(key termbox.Key, fct func(*Grid)) *KeyHandler {
	hdlr := &KeyHandler{Key: key, Fct: fct}
	g.keyHandlers = append(g.keyHandlers, hdlr)
	return hdlr
}

// UnregisterKeyHandler removes the given handler from the list.
// hdlr is returned by (*gogrid.Grid).RegisterKeyHandler.
func (g *Grid) UnregisterKeyHandler(hdlr *KeyHandler) {
	for i, elem := range g.keyHandlers {
		if elem == hdlr {
			g.keyHandlers = append(g.keyHandlers[:i], g.keyHandlers[i+1:]...)
		}
	}
}

// HandleKeyboard is the runtime loop monitoring keyboard activity.
func (g *Grid) HandleKeyboard() error {
mainloop:
	for {
		select {
		case <-g.stopChan:
			break mainloop
		default:
		}

		termbox.SetCursor(0, 1)
		_ = termbox.Flush()

		g.HeaderFct(g)

		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc, termbox.KeyCtrlC, termbox.KeyCtrlBackslash:
				break mainloop
			}
			for _, hdlr := range g.keyHandlers {
				if ev.Key == hdlr.Key || (ev.Key == 0x00 && ev.Ch == rune(hdlr.Key)) {
					hdlr.Fct(g)
				}
			}
		case termbox.EventError:
			return errors.Wrap(ev.Err, "error with the terminal")
		}
	}
	return nil
}

// Close terminates the runtime loop.
func (g *Grid) Close() error {
	// Signal the channel if not already closed.
	select {
	case <-g.stopChan:
	default:
		// Close termbox.
		termbox.Close()
		close(g.stopChan)
	}
	return nil
}

// SetCursorOrigin sets the cursor back to 0,0 to display the header.
func (g *Grid) SetCursorOrigin() {
	termbox.SetCursor(0, 0)
	_ = termbox.Flush()
}

// ClearHeader zero out the header and reset the cursor to 0,0.
func (g *Grid) ClearHeader() {
	g.SetCursorOrigin()
	x, _ := termbox.Size()
	for i := 0; i < g.HeaderHeight; i++ {
		fmt.Print(strings.Repeat(" ", x))
	}
	g.SetCursorOrigin()
}
