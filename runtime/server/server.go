package server

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/creack/gofour/engine"
	"github.com/creack/gofour/runtime"
	"github.com/creack/gofour/runtime/text"
	"github.com/pkg/errors"
)

func init() {
	runtime.Runtimes["server"] = &Runtime{}
}

// Runtime is a HTTP server for Connect Four.
type Runtime struct {
	four *engine.Four
}

// Init setup the connect four game.
func (r *Runtime) Init(four *engine.Four) error {
	r.four = four
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		_ = req.ParseForm()
		colStr := req.Form.Get("col")
		col, err := strconv.Atoi(colStr)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			text.Dump(w, r.four)
			fmt.Fprintf(w, "invalid column: %s\n", colStr)
			return
		}
		col-- // Back to 0 index.

		ret, err := r.four.PlayerMove(col)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			text.Dump(w, r.four)
			fmt.Fprintf(w, "Error executing move: %s\n", err)
			return
		}

		text.Dump(w, r.four)
		if ret != engine.Empty {
			_ = r.Close()
			if ret == engine.Stale {
				fmt.Fprintf(w, "Sale, nobody wins!\n")
			} else {
				fmt.Fprintf(w, "Player %d (%s) won!\n", ret, ret)
			}
		} else {
			fmt.Fprintf(w, "Player's %d (%s) turn!\n", r.four.CurPlayer, r.four.CurPlayer)
		}
	})
	return nil
}

// Run is the main loop.
// TODO: use flags to config the listen address.
func (r *Runtime) Run() error {
	return http.ListenAndServe("0.0.0.0:8080", nil)
}

// Close terminates the game.
func (r *Runtime) Close() error {
	f, err := r.four.Reset()
	if err != nil {
		return errors.Wrap(err, "error restarting the game")
	}
	r.four = f
	return nil
}
