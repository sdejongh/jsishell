package terminal

import (
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestPromptExpanderBasic(t *testing.T) {
	p := NewPromptExpander()

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "no variables",
			format: "$ ",
			want:   "$ ",
		},
		{
			name:   "literal percent",
			format: "100%% done",
			want:   "100% done",
		},
		{
			name:   "newline",
			format: "line1%nline2",
			want:   "line1\nline2",
		},
		{
			name:   "unknown variable kept as-is",
			format: "%x",
			want:   "%x",
		},
		{
			name:   "percent at end",
			format: "test%",
			want:   "test%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Expand(tt.format)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestPromptExpanderWorkDir(t *testing.T) {
	p := NewPromptExpander()
	p.SetWorkDir("/home/testuser/projects/myapp")

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "full path %d",
			format: "%d",
			want:   "/home/testuser/projects/myapp",
		},
		{
			name:   "basename %D",
			format: "%D",
			want:   "myapp",
		},
		{
			name:   "in prompt context",
			format: "[%D] $ ",
			want:   "[myapp] $ ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Expand(tt.format)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.format, got, tt.want)
			}
		})
	}
}

func TestPromptExpanderShortPath(t *testing.T) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		t.Skip("Cannot get home directory")
	}

	p := NewPromptExpander()

	tests := []struct {
		name    string
		workDir string
		want    string
	}{
		{
			name:    "home directory",
			workDir: homeDir,
			want:    "~",
		},
		{
			name:    "subdirectory of home",
			workDir: filepath.Join(homeDir, "projects"),
			want:    "~/projects",
		},
		{
			name:    "deep subdirectory",
			workDir: filepath.Join(homeDir, "projects", "myapp", "src"),
			want:    "~/projects/myapp/src",
		},
		{
			name:    "outside home",
			workDir: "/tmp/test",
			want:    "/tmp/test",
		},
		{
			name:    "root",
			workDir: "/",
			want:    "/",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p.SetWorkDir(tt.workDir)
			got := p.Expand("%~")
			if got != tt.want {
				t.Errorf("Expand(\"%%~\") with workDir=%q = %q, want %q", tt.workDir, got, tt.want)
			}
		})
	}
}

func TestPromptExpanderUsername(t *testing.T) {
	p := NewPromptExpander()
	got := p.Expand("%u")

	// Should return something non-empty
	if got == "" {
		t.Error("Username should not be empty")
	}

	// Verify it matches current user
	if u, err := user.Current(); err == nil {
		if got != u.Username {
			t.Errorf("Username = %q, want %q", got, u.Username)
		}
	}
}

func TestPromptExpanderHostname(t *testing.T) {
	p := NewPromptExpander()

	// Short hostname
	shortHost := p.Expand("%h")
	if shortHost == "" {
		t.Error("Short hostname should not be empty")
	}
	if strings.Contains(shortHost, ".") {
		// If system hostname has dots, short version shouldn't
		fullHost := p.Expand("%H")
		if strings.Contains(fullHost, ".") && shortHost == fullHost {
			t.Error("Short hostname should not contain dots if full hostname does")
		}
	}

	// Full hostname
	fullHost := p.Expand("%H")
	if fullHost == "" {
		t.Error("Full hostname should not be empty")
	}
}

func TestPromptExpanderTime(t *testing.T) {
	p := NewPromptExpander()

	// Time HH:MM
	shortTime := p.Expand("%t")
	if len(shortTime) != 5 { // "HH:MM"
		t.Errorf("Short time format should be 5 chars, got %q", shortTime)
	}
	if shortTime[2] != ':' {
		t.Errorf("Short time should have colon at position 2, got %q", shortTime)
	}

	// Time HH:MM:SS
	longTime := p.Expand("%T")
	if len(longTime) != 8 { // "HH:MM:SS"
		t.Errorf("Long time format should be 8 chars, got %q", longTime)
	}
	if longTime[2] != ':' || longTime[5] != ':' {
		t.Errorf("Long time should have colons at positions 2 and 5, got %q", longTime)
	}

	// Verify time is approximately correct
	now := time.Now()
	expectedShort := now.Format("15:04")
	if shortTime != expectedShort {
		// Allow 1 minute difference in case of minute rollover during test
		t.Logf("Time might differ slightly: got %q, expected %q", shortTime, expectedShort)
	}
}

func TestPromptExpanderCombined(t *testing.T) {
	p := NewPromptExpander()
	p.SetWorkDir("/home/testuser/projects")

	// Test a realistic prompt format
	format := "%u@%h:%D $ "
	got := p.Expand(format)

	// Should contain the username
	if u, err := user.Current(); err == nil {
		if !strings.Contains(got, u.Username) {
			t.Errorf("Prompt should contain username %q, got %q", u.Username, got)
		}
	}

	// Should contain @
	if !strings.Contains(got, "@") {
		t.Errorf("Prompt should contain @, got %q", got)
	}

	// Should contain :
	if !strings.Contains(got, ":") {
		t.Errorf("Prompt should contain :, got %q", got)
	}

	// Should contain directory basename
	if !strings.Contains(got, "projects") {
		t.Errorf("Prompt should contain 'projects', got %q", got)
	}

	// Should end with " $ "
	if !strings.HasSuffix(got, " $ ") {
		t.Errorf("Prompt should end with ' $ ', got %q", got)
	}
}

func TestExpandPromptFunction(t *testing.T) {
	got := ExpandPrompt("%D $ ", "/home/user/myproject")
	want := "myproject $ "

	if got != want {
		t.Errorf("ExpandPrompt() = %q, want %q", got, want)
	}
}

func TestPromptExpanderMultipleVariables(t *testing.T) {
	p := NewPromptExpander()
	p.SetWorkDir("/tmp/test")

	format := "[%t] %d %% %D"
	got := p.Expand(format)

	// Should contain time in brackets
	if !strings.HasPrefix(got, "[") {
		t.Errorf("Should start with [, got %q", got)
	}

	// Should contain full path
	if !strings.Contains(got, "/tmp/test") {
		t.Errorf("Should contain /tmp/test, got %q", got)
	}

	// Should contain literal %
	if !strings.Contains(got, " % ") {
		t.Errorf("Should contain literal %%, got %q", got)
	}

	// Should end with basename
	if !strings.HasSuffix(got, "test") {
		t.Errorf("Should end with 'test', got %q", got)
	}
}

func TestPromptExpanderColors(t *testing.T) {
	p := NewPromptExpander()
	p.SetWorkDir("/home/testuser")

	tests := []struct {
		name   string
		format string
		want   string
	}{
		{
			name:   "simple color",
			format: "%{green}test%{/}",
			want:   "\033[32mtest\033[0m",
		},
		{
			name:   "reset keyword",
			format: "%{red}error%{reset}",
			want:   "\033[31merror\033[0m",
		},
		{
			name:   "bright color",
			format: "%{bright_blue}info%{/}",
			want:   "\033[94minfo\033[0m",
		},
		{
			name:   "bold style",
			format: "%{bold}important%{/}",
			want:   "\033[1mimportant\033[0m",
		},
		{
			name:   "dim style",
			format: "%{dim}faded%{/}",
			want:   "\033[2mfaded\033[0m",
		},
		{
			name:   "underline style",
			format: "%{underline}link%{/}",
			want:   "\033[4mlink\033[0m",
		},
		{
			name:   "color with variable",
			format: "%{cyan}%D%{/}",
			want:   "\033[36mtestuser\033[0m",
		},
		{
			name:   "multiple colors",
			format: "%{green}user%{/}@%{blue}host%{/}",
			want:   "\033[32muser\033[0m@\033[34mhost\033[0m",
		},
		{
			name:   "unknown color ignored",
			format: "%{unknowncolor}text%{/}",
			want:   "text\033[0m",
		},
		{
			name:   "realistic prompt",
			format: "%{green}%u%{/}@%{cyan}%h%{/}:%{blue}%D%{/}$ ",
			want:   "", // We'll check it contains color codes
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := p.Expand(tt.format)
			if tt.want != "" && got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.format, got, tt.want)
			}
			if tt.name == "realistic prompt" {
				// Just verify it contains ANSI codes and ends with "$ "
				if !strings.Contains(got, "\033[") {
					t.Errorf("Realistic prompt should contain ANSI codes, got %q", got)
				}
				if !strings.HasSuffix(got, "$ ") {
					t.Errorf("Realistic prompt should end with '$ ', got %q", got)
				}
			}
		})
	}
}

func TestPromptExpanderColorsDisabled(t *testing.T) {
	p := NewPromptExpander()
	p.SetColorsActive(false)

	format := "%{green}text%{/}"
	got := p.Expand(format)
	want := "text" // No color codes when disabled

	if got != want {
		t.Errorf("With colors disabled, Expand(%q) = %q, want %q", format, got, want)
	}
}

func TestPromptExpanderShellIndicator(t *testing.T) {
	p := NewPromptExpander()

	// Test the shell indicator variable
	got := p.Expand("%$")

	// For non-root users, should be $
	// For root, should be #
	if os.Getuid() == 0 {
		if got != "#" {
			t.Errorf("Shell indicator for root should be #, got %q", got)
		}
	} else {
		if got != "$" {
			t.Errorf("Shell indicator for user should be $, got %q", got)
		}
	}

	// Test in a prompt context
	format := "%u@%h %$ "
	got = p.Expand(format)
	if os.Getuid() == 0 {
		if !strings.HasSuffix(got, "# ") {
			t.Errorf("Prompt should end with '# ', got %q", got)
		}
	} else {
		if !strings.HasSuffix(got, "$ ") {
			t.Errorf("Prompt should end with '$ ', got %q", got)
		}
	}
}
