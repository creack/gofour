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

// Dump displays the state of the grid on stdout.
func Dump(f *engine.Four) {
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
	_ = w.Flush()
	fmt.Println()
}

// Runtime is a basic text based client for connect four.
type Runtime struct {
	four     *engine.Four
	stopChan chan struct{}
	r        *io.PipeReader
	w        *io.PipeWriter
}

// Init setups up the connect four game.
func (tf *Runtime) Init(four *engine.Four) error {
	tf.four = four
	tf.stopChan = make(chan struct{})

	// Setup the pipe to allow to interrupt scanf.
	tf.r, tf.w = io.Pipe()
	go func() {
		_, _ = io.Copy(tf.w, os.Stdin)
	}()

	// Watch for signals so we are not stuck in the game forever.
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	go func() {
		<-ch

		signal.Stop(ch)
		close(ch)
		_ = tf.Close()
	}()

	return nil
}

// Run starts the game loop.
func (tf *Runtime) Run() error {
	for i := 0; ; i++ {
	start:
		select {
		case <-tf.stopChan:
			return nil
		default:
		}
		Dump(tf.four)
		fmt.Printf("Player %d (%s) turn, select column:\n", tf.four.CurPlayer, tf.four.CurPlayer)

		var x int
		if _, err := fmt.Fscanf(tf.r, "%d", &x); err != nil {
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

		ret, err := tf.four.PlayerMove(x)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err)
			if err := errors.Cause(err); err == engine.ErrInvalidMove {
				goto start
			}
			return err
		}
		if ret != engine.Empty {
			Dump(tf.four)
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
func (tf *Runtime) Close() error {
	select {
	case <-tf.stopChan:
	default:
		_ = tf.w.CloseWithError(io.EOF)
		close(tf.stopChan)
	}
	return nil
}
