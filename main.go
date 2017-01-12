package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/creack/gofour/engine"
	"github.com/creack/gofour/runtime"

	// Load runtimes.
	_ "github.com/creack/gofour/runtime/server"
	_ "github.com/creack/gofour/runtime/terminal"
	_ "github.com/creack/gofour/runtime/text"
)

func main() {
	var (
		cols     = flag.Int("cols", engine.DefaultCols, "number of columns")
		rows     = flag.Int("rows", engine.DefaultRows, "number of rows")
		nPlayers = flag.Int("p", engine.DefaultNPlayers, fmt.Sprintf("number of players. (max: %d)", len(engine.AvailablePlayers)))
		nWin     = flag.Int("w", engine.DefaultNWin, "number of consecutive color to win")
		mode     = flag.String("mode", engine.DefaultMode, "Game mode. Values: [terminal, text]")
	)
	flag.Parse()

	run, exists := runtime.Runtimes[*mode]
	if !exists {
		log.Fatalf("%s is not a valid runtime.", *mode)
	}

	four, err := engine.NewConnectFour(*cols, *rows, *nPlayers, *nWin)
	if err != nil {
		log.Fatal(err)
	}

	if err := run.Init(four); err != nil {
		log.Fatal(err)
	}
	defer func() { _ = run.Close() }() // Best effort.

	if err := run.Run(); err != nil {
		log.Fatal(err)
	}
}
