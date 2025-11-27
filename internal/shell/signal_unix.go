//go:build unix

package shell

import (
	"os"
	"os/signal"
	"syscall"
)

// platformSignals returns the signals to handle on Unix platforms.
func platformSignals() []os.Signal {
	return []os.Signal{
		os.Interrupt,    // SIGINT (Ctrl+C)
		syscall.SIGTERM, // Termination request
		syscall.SIGQUIT, // Quit (Ctrl+\)
		syscall.SIGHUP,  // Hangup (terminal closed)
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
	case syscall.SIGTERM:
		// Terminate gracefully
		s.Exit(0)
		return false
	case syscall.SIGQUIT:
		// Quit - could be used for emergency exit
		s.Exit(1)
		return false
	case syscall.SIGHUP:
		// Terminal hangup - save state and exit
		s.saveHistory()
		s.Exit(0)
		return false
	default:
		return true
	}
}

// setupPlatformSignals performs platform-specific signal setup.
func setupPlatformSignals(sigChan chan os.Signal) {
	signal.Notify(sigChan, platformSignals()...)
}
