//go:build windows

package shell

import (
	"os"
	"os/signal"
)

// platformSignals returns the signals to handle on Windows platforms.
// Windows has limited signal support compared to Unix.
func platformSignals() []os.Signal {
	return []os.Signal{
		os.Interrupt, // Ctrl+C
	}
}

// handlePlatformSignal handles platform-specific signals.
// Returns true if the signal was handled and the shell should continue,
// false if the shell should exit.
func (s *Shell) handlePlatformSignal(sig os.Signal) bool {
	switch sig {
	case os.Interrupt:
		// Ctrl+C - interrupt current operation, continue shell
		return true
	default:
		return true
	}
}

// setupPlatformSignals performs platform-specific signal setup.
func setupPlatformSignals(sigChan chan os.Signal) {
	signal.Notify(sigChan, platformSignals()...)
}
