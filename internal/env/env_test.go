package env

import (
	"os"
	"sort"
	"strings"
	"testing"
)

func TestNew(t *testing.T) {
	env := New()

	// Should have at least some OS environment variables
	if len(env.vars) == 0 {
		t.Error("New() created environment with no variables")
	}

	// PATH should typically be set
	if path := env.Get("PATH"); path == "" {
		t.Log("PATH not found in environment (may be normal in isolated test)")
	}
}

func TestGetSet(t *testing.T) {
	env := New()

	// Set a variable
	env.Set("TEST_VAR", "test_value")

	// Get it back
	if got := env.Get("TEST_VAR"); got != "test_value" {
		t.Errorf("Get(TEST_VAR) = %q, want %q", got, "test_value")
	}

	// Get non-existent variable
	if got := env.Get("NONEXISTENT_VAR_12345"); got != "" {
		t.Errorf("Get(NONEXISTENT) = %q, want empty string", got)
	}

	// Overwrite variable
	env.Set("TEST_VAR", "new_value")
	if got := env.Get("TEST_VAR"); got != "new_value" {
		t.Errorf("Get(TEST_VAR) after overwrite = %q, want %q", got, "new_value")
	}
}

func TestUnset(t *testing.T) {
	env := New()

	// Set and export a variable
	env.Set("TEST_VAR", "value")
	env.Export("TEST_VAR")

	// Verify it exists
	if env.Get("TEST_VAR") != "value" {
		t.Fatal("Setup failed: variable not set")
	}
	if !env.IsExported("TEST_VAR") {
		t.Fatal("Setup failed: variable not exported")
	}

	// Unset it
	env.Unset("TEST_VAR")

	// Verify it's gone
	if env.Get("TEST_VAR") != "" {
		t.Error("Get(TEST_VAR) after Unset should be empty")
	}
	if env.IsExported("TEST_VAR") {
		t.Error("IsExported(TEST_VAR) after Unset should be false")
	}
}

func TestExport(t *testing.T) {
	env := New()

	// Set a variable (not exported by default)
	env.Set("MY_VAR", "my_value")

	// Initially not exported
	if env.IsExported("MY_VAR") {
		t.Error("IsExported(MY_VAR) should be false before Export")
	}

	// Export it
	env.Export("MY_VAR")

	// Now it should be exported
	if !env.IsExported("MY_VAR") {
		t.Error("IsExported(MY_VAR) should be true after Export")
	}

	// Export non-existent variable should be no-op
	env.Export("NONEXISTENT_VAR")
	if env.IsExported("NONEXISTENT_VAR") {
		t.Error("Export of non-existent var should not create export entry")
	}
}

func TestExpand(t *testing.T) {
	env := New()
	env.Set("HOME", "/home/user")
	env.Set("NAME", "John")
	env.Set("EMPTY", "")

	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"simple var", "$HOME", "/home/user"},
		{"braced var", "${HOME}", "/home/user"},
		{"multiple vars", "$HOME/$NAME", "/home/user/John"},
		{"mixed syntax", "${HOME}/$NAME", "/home/user/John"},
		{"nonexistent var", "$UNKNOWN", ""},
		{"empty var", "$EMPTY", ""},
		{"no vars", "plain text", "plain text"},
		{"var in text", "Hello $NAME!", "Hello John!"},
		{"adjacent vars", "$HOME$NAME", "/home/userJohn"},
		{"escaped like", "$$HOME", "$/home/user"}, // $ followed by $HOME
		{"partial match", "$", "$"},               // lone $ stays as-is
		{"braced nonexistent", "${UNKNOWN}", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := env.Expand(tt.input)
			if got != tt.want {
				t.Errorf("Expand(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestToSlice(t *testing.T) {
	env := New()

	// Clear and set up fresh environment
	env.vars = make(map[string]string)
	env.exports = make(map[string]bool)

	env.Set("VAR1", "value1")
	env.Set("VAR2", "value2")
	env.Set("VAR3", "value3")

	// Export only VAR1 and VAR2
	env.Export("VAR1")
	env.Export("VAR2")

	slice := env.ToSlice()

	// Sort for consistent comparison
	sort.Strings(slice)

	// Should only contain exported vars
	if len(slice) != 2 {
		t.Errorf("ToSlice() returned %d items, want 2", len(slice))
	}

	expected := []string{"VAR1=value1", "VAR2=value2"}
	sort.Strings(expected)

	for i, want := range expected {
		if i >= len(slice) || slice[i] != want {
			t.Errorf("ToSlice()[%d] = %q, want %q", i, slice[i], want)
		}
	}
}

func TestAll(t *testing.T) {
	env := New()

	// Clear and set up fresh environment
	env.vars = make(map[string]string)
	env.exports = make(map[string]bool)

	env.Set("A", "1")
	env.Set("B", "2")

	all := env.All()

	if len(all) != 2 {
		t.Errorf("All() returned %d items, want 2", len(all))
	}

	if all["A"] != "1" || all["B"] != "2" {
		t.Errorf("All() = %v, want map[A:1 B:2]", all)
	}

	// Modifying returned map should not affect original
	all["A"] = "modified"
	if env.Get("A") != "1" {
		t.Error("Modifying All() result should not affect environment")
	}
}

func TestClone(t *testing.T) {
	env := New()

	// Clear and set up fresh environment
	env.vars = make(map[string]string)
	env.exports = make(map[string]bool)

	env.Set("VAR1", "value1")
	env.Export("VAR1")

	clone := env.Clone()

	// Clone should have same values
	if clone.Get("VAR1") != "value1" {
		t.Error("Clone does not have VAR1")
	}
	if !clone.IsExported("VAR1") {
		t.Error("Clone VAR1 not exported")
	}

	// Modifying clone should not affect original
	clone.Set("VAR1", "modified")
	if env.Get("VAR1") != "value1" {
		t.Error("Modifying clone affected original")
	}

	// Modifying original should not affect clone
	env.Set("VAR2", "value2")
	if clone.Get("VAR2") != "" {
		t.Error("Modifying original affected clone")
	}
}

func TestConcurrentAccess(t *testing.T) {
	env := New()
	done := make(chan bool, 10)

	// Concurrent writes
	for i := 0; i < 5; i++ {
		go func(n int) {
			for j := 0; j < 100; j++ {
				env.Set("CONCURRENT", "value")
				env.Get("CONCURRENT")
				env.Expand("$CONCURRENT")
			}
			done <- true
		}(i)
	}

	// Concurrent reads
	for i := 0; i < 5; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				env.All()
				env.ToSlice()
			}
			done <- true
		}()
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestOSEnvironmentImport(t *testing.T) {
	// Set a known OS variable
	key := "JSISHELL_TEST_VAR_" + strings.ReplaceAll(t.Name(), "/", "_")
	os.Setenv(key, "test_value_from_os")
	defer os.Unsetenv(key)

	env := New()

	// Should have imported the OS variable
	if got := env.Get(key); got != "test_value_from_os" {
		t.Errorf("Get(%s) = %q, want %q", key, got, "test_value_from_os")
	}

	// OS variables should be exported by default
	if !env.IsExported(key) {
		t.Errorf("OS variable %s should be exported by default", key)
	}
}
