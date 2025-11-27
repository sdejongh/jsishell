package terminal

import (
	"os"
	"testing"
)

// T075: Tests for color support detection
func TestColorSupport(t *testing.T) {
	tests := []struct {
		name        string
		setup       func()
		cleanup     func()
		wantSupport bool
	}{
		{
			name: "NO_COLOR set disables colors",
			setup: func() {
				os.Setenv("NO_COLOR", "1")
			},
			cleanup: func() {
				os.Unsetenv("NO_COLOR")
			},
			wantSupport: false,
		},
		{
			name: "TERM=dumb disables colors",
			setup: func() {
				os.Setenv("TERM", "dumb")
			},
			cleanup: func() {
				os.Unsetenv("TERM")
			},
			wantSupport: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			defer func() {
				if tt.cleanup != nil {
					tt.cleanup()
				}
			}()

			cs := NewColorScheme(nil)
			// Force enabled to check environment override
			cs.enabled = true

			if cs.IsSupported() != tt.wantSupport {
				t.Errorf("IsSupported() = %v, want %v", cs.IsSupported(), tt.wantSupport)
			}
		})
	}
}

func TestColorSchemeEnabled(t *testing.T) {
	cs := NewColorScheme(nil)

	// Default should be enabled
	if !cs.enabled {
		t.Error("ColorScheme should be enabled by default")
	}

	cs.SetEnabled(false)
	if cs.enabled {
		t.Error("ColorScheme should be disabled after SetEnabled(false)")
	}

	cs.SetEnabled(true)
	if !cs.enabled {
		t.Error("ColorScheme should be enabled after SetEnabled(true)")
	}
}

// T076: Tests for ANSI color codes
func TestColorCodes(t *testing.T) {
	tests := []struct {
		color string
		want  string
	}{
		{"black", "\033[30m"},
		{"red", "\033[31m"},
		{"green", "\033[32m"},
		{"yellow", "\033[33m"},
		{"blue", "\033[34m"},
		{"magenta", "\033[35m"},
		{"cyan", "\033[36m"},
		{"white", "\033[37m"},
		{"bright_black", "\033[90m"},
		{"bright_red", "\033[91m"},
		{"bright_green", "\033[92m"},
		{"bright_yellow", "\033[93m"},
		{"bright_blue", "\033[94m"},
		{"bright_magenta", "\033[95m"},
		{"bright_cyan", "\033[96m"},
		{"bright_white", "\033[97m"},
	}

	for _, tt := range tests {
		t.Run(tt.color, func(t *testing.T) {
			code := ColorCode(tt.color)
			if code != tt.want {
				t.Errorf("ColorCode(%q) = %q, want %q", tt.color, code, tt.want)
			}
		})
	}
}

func TestColorCodeInvalid(t *testing.T) {
	code := ColorCode("invalid_color")
	if code != "" {
		t.Errorf("ColorCode(invalid) = %q, want empty string", code)
	}
}

func TestResetCode(t *testing.T) {
	if ResetCode != "\033[0m" {
		t.Errorf("ResetCode = %q, want %q", ResetCode, "\033[0m")
	}
}

func TestColorize(t *testing.T) {
	cs := NewColorScheme(nil)
	cs.SetEnabled(true)
	// Clear environment to ensure colors work
	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "xterm-256color")
	defer os.Unsetenv("TERM")

	tests := []struct {
		name  string
		text  string
		color string
		want  string
	}{
		{
			name:  "colorize with red",
			text:  "error",
			color: "red",
			want:  "\033[31merror\033[0m",
		},
		{
			name:  "colorize with blue",
			text:  "directory",
			color: "blue",
			want:  "\033[34mdirectory\033[0m",
		},
		{
			name:  "colorize with bright_green",
			text:  "success",
			color: "bright_green",
			want:  "\033[92msuccess\033[0m",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cs.Colorize(tt.text, tt.color)
			if got != tt.want {
				t.Errorf("Colorize(%q, %q) = %q, want %q", tt.text, tt.color, got, tt.want)
			}
		})
	}
}

func TestColorizeDisabled(t *testing.T) {
	cs := NewColorScheme(nil)
	cs.SetEnabled(false)

	got := cs.Colorize("text", "red")
	if got != "text" {
		t.Errorf("Colorize with disabled colors = %q, want %q", got, "text")
	}
}

func TestColorizeWithNO_COLOR(t *testing.T) {
	os.Setenv("NO_COLOR", "1")
	defer os.Unsetenv("NO_COLOR")

	cs := NewColorScheme(nil)
	cs.SetEnabled(true)

	got := cs.Colorize("text", "red")
	if got != "text" {
		t.Errorf("Colorize with NO_COLOR = %q, want %q", got, "text")
	}
}

func TestColorizeInvalidColor(t *testing.T) {
	cs := NewColorScheme(nil)
	cs.SetEnabled(true)
	os.Unsetenv("NO_COLOR")

	got := cs.Colorize("text", "invalid_color")
	if got != "text" {
		t.Errorf("Colorize with invalid color = %q, want %q", got, "text")
	}
}

func TestStripColors(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "strip single color",
			input: "\033[31mred text\033[0m",
			want:  "red text",
		},
		{
			name:  "strip multiple colors",
			input: "\033[31mred\033[0m and \033[34mblue\033[0m",
			want:  "red and blue",
		},
		{
			name:  "no colors to strip",
			input: "plain text",
			want:  "plain text",
		},
		{
			name:  "strip bright colors",
			input: "\033[92mbright green\033[0m",
			want:  "bright green",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := StripColors(tt.input)
			if got != tt.want {
				t.Errorf("StripColors(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestColorSchemeColors(t *testing.T) {
	cs := NewColorScheme(nil)
	os.Unsetenv("NO_COLOR")
	os.Setenv("TERM", "xterm")
	defer os.Unsetenv("TERM")

	// Test convenience methods
	text := "test"

	if cs.Directory(text) == text {
		t.Error("Directory() should colorize when enabled")
	}

	if cs.Error(text) == text {
		t.Error("Error() should colorize when enabled")
	}

	if cs.Success(text) == text {
		t.Error("Success() should colorize when enabled")
	}
}
