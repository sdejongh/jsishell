// Package terminal provides terminal handling and prompt expansion.
package terminal

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

// PromptExpander expands prompt variables to their values.
type PromptExpander struct {
	workDir string // Current working directory
	homeDir string // User's home directory
}

// NewPromptExpander creates a new PromptExpander.
func NewPromptExpander() *PromptExpander {
	homeDir, _ := os.UserHomeDir()
	workDir, _ := os.Getwd()

	return &PromptExpander{
		workDir: workDir,
		homeDir: homeDir,
	}
}

// SetWorkDir updates the current working directory.
func (p *PromptExpander) SetWorkDir(dir string) {
	p.workDir = dir
}

// Expand expands all prompt variables in the given format string.
//
// Supported variables:
//   - %d  - Current working directory (full path)
//   - %D  - Current working directory (basename only)
//   - %~  - Current working directory with ~ for home
//   - %u  - Username
//   - %h  - Hostname (short, without domain)
//   - %H  - Hostname (full, with domain)
//   - %t  - Time (HH:MM)
//   - %T  - Time with seconds (HH:MM:SS)
//   - %n  - Newline
//   - %%  - Literal %
func (p *PromptExpander) Expand(format string) string {
	var result strings.Builder
	result.Grow(len(format) * 2) // Pre-allocate some space

	i := 0
	for i < len(format) {
		if format[i] == '%' && i+1 < len(format) {
			switch format[i+1] {
			case 'd': // Full working directory
				result.WriteString(p.workDir)
				i += 2
			case 'D': // Basename of working directory
				result.WriteString(filepath.Base(p.workDir))
				i += 2
			case '~': // Working directory with ~ for home
				result.WriteString(p.shortPath())
				i += 2
			case 'u': // Username
				result.WriteString(p.username())
				i += 2
			case 'h': // Short hostname
				result.WriteString(p.shortHostname())
				i += 2
			case 'H': // Full hostname
				result.WriteString(p.fullHostname())
				i += 2
			case 't': // Time HH:MM
				result.WriteString(time.Now().Format("15:04"))
				i += 2
			case 'T': // Time HH:MM:SS
				result.WriteString(time.Now().Format("15:04:05"))
				i += 2
			case 'n': // Newline
				result.WriteByte('\n')
				i += 2
			case '%': // Literal %
				result.WriteByte('%')
				i += 2
			default:
				// Unknown variable, keep as-is
				result.WriteByte(format[i])
				i++
			}
		} else {
			result.WriteByte(format[i])
			i++
		}
	}

	return result.String()
}

// shortPath returns the working directory with ~ substituted for home.
func (p *PromptExpander) shortPath() string {
	if p.homeDir != "" && strings.HasPrefix(p.workDir, p.homeDir) {
		return "~" + p.workDir[len(p.homeDir):]
	}
	return p.workDir
}

// username returns the current username.
func (p *PromptExpander) username() string {
	if u, err := user.Current(); err == nil {
		return u.Username
	}
	// Fallback to environment variable
	if name := os.Getenv("USER"); name != "" {
		return name
	}
	return "user"
}

// shortHostname returns the hostname without domain.
func (p *PromptExpander) shortHostname() string {
	hostname := p.fullHostname()
	if idx := strings.Index(hostname, "."); idx != -1 {
		return hostname[:idx]
	}
	return hostname
}

// fullHostname returns the full hostname.
func (p *PromptExpander) fullHostname() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return "localhost"
}

// ExpandPrompt is a convenience function that expands a prompt string.
// It creates a temporary expander with the given working directory.
func ExpandPrompt(format, workDir string) string {
	p := NewPromptExpander()
	p.SetWorkDir(workDir)
	return p.Expand(format)
}
