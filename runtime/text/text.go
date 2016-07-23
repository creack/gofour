package text

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"syscall"
	"text/tabwriter"

	"github.com/creack/gofour/engine"
	"github.com/creack/gofour/runtime"
	"github.com/pkg/errors"
)

func init() {
	runtime.Runtimes["text"] = &Runtime{}
}

// Dump displays the state of the grid on the given writer.
func Dump(w io.Writer, f *engine.Four) {
	fmt.Fprintln(w)

	tabW := tabwriter.NewWriter(w, 4, 4, 4, ' ', 0)
	for i := 0; i < f.Columns; i++ {
		fmt.Fprintf(tabW, "\x1b[1;37m%d\x1b[0m\t", i+1)
	}
	fmt.Fprintln(tabW)
	for i := 0; i < f.Columns; i++ {
		fmt.Fprint(tabW, "\x1b[1;37m---\x1b[0m\t")
	}
	fmt.Fprintln(tabW)
	for i := 0; i < f.Rows; i++ {
		for j := 0; j < f.Columns; j++ {
			fmt.Fprintf(tabW, "%s\t", f.State(i, j))
		}
		fmt.Fprintf(tabW, "\n")
	}
	_ = tabW.Flush()
	fmt.Fprintln(w)
}

// Runtime is a basic text based client for connect four.
type Runtime struct {
	four     *engine.Four
	stopChan chan struct{}
	r        *io.PipeReader
	w        *io.PipeWriter
}

// Init setup the connect four game.
func (r *Runtime) Init(four *engine.Four) error {
	r.four = four
	r.stopChan = make(chan struct{})

	// Setup the pipe to allow to interrupt scanf.
	r.r, r.w = io.Pipe()
	go func() {
		_, _ = io.Copy(r.w, os.Stdin)
	}()

	// Watch for signals so we are not stuck in the game forever.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-ch

		signal.Stop(ch)
		close(ch)
		_ = r.Close()
	}()

	return nil
}

// Run starts the game loop.
func (r *Runtime) Run() error {
	for i := 0; ; i++ {
	start:
		select {
		case <-r.stopChan:
			return nil
		default:
		}
		Dump(os.Stdout, r.four)
		fmt.Printf("Player %d (%s) turn, select column:\n", r.four.CurPlayer, r.four.CurPlayer)

		var x int
		if _, err := fmt.Fscanf(r.r, "%d", &x); err != nil {
			if err == io.EOF {
				break
			}
			// Ignore other errors.
		}
		x-- // Back to 0 index.

		if x < 0 {
			fmt.Fprint(os.Stderr, "invalid columns number\n")
			goto start
		}

		ret, err := r.four.PlayerMove(x)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			if err := errors.Cause(err); err == engine.ErrInvalidMove {
				goto start
			}
			return err
		}
		if ret != engine.Empty {
			Dump(os.Stdout, r.four)
			if ret == engine.Stale {
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
func (r *Runtime) Close() error {
	select {
	case <-r.stopChan:
	default:
		_ = r.w.CloseWithError(io.EOF)
		close(r.stopChan)
	}
	return nil
}
