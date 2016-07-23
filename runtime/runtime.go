package runtime

import "github.com/creack/gofour/engine"

// FourRuntime is the interface to run connect four.
type FourRuntime interface {
	Init(*engine.Four) error
	Run() error
	Close() error
}

// Runtimes holds the registered runtimes.
var Runtimes = map[string]FourRuntime{}
