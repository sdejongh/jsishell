package executor

import (
	"sort"
	"testing"
)

func TestTrieInsertAndExactMatch(t *testing.T) {
	trie := NewTrie()

	commands := []string{"list", "copy", "move", "cd", "clear"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	// Test exact matches
	for _, cmd := range commands {
		if !trie.ExactMatch(cmd) {
			t.Errorf("ExactMatch(%q) = false, want true", cmd)
		}
	}

	// Test non-existent commands
	nonExistent := []string{"l", "co", "xyz", "listing"}
	for _, cmd := range nonExistent {
		if trie.ExactMatch(cmd) {
			t.Errorf("ExactMatch(%q) = true, want false", cmd)
		}
	}
}

func TestTrieSearch(t *testing.T) {
	trie := NewTrie()

	commands := []string{"list", "copy", "clear", "cd", "compare", "count"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	tests := []struct {
		prefix string
		want   []string
	}{
		{"l", []string{"list"}},
		{"c", []string{"copy", "clear", "cd", "compare", "count"}},
		{"co", []string{"copy", "compare", "count"}},
		{"cop", []string{"copy"}},
		{"cl", []string{"clear"}},
		{"cd", []string{"cd"}},
		{"x", nil},
		{"listing", nil},
		{"", []string{"list", "copy", "clear", "cd", "compare", "count"}},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := trie.Search(tt.prefix)

			// Sort both for comparison
			sort.Strings(got)
			sort.Strings(tt.want)

			if len(got) != len(tt.want) {
				t.Errorf("Search(%q) = %v, want %v", tt.prefix, got, tt.want)
				return
			}

			for i := range got {
				if got[i] != tt.want[i] {
					t.Errorf("Search(%q) = %v, want %v", tt.prefix, got, tt.want)
					return
				}
			}
		})
	}
}

func TestTrieMatch(t *testing.T) {
	trie := NewTrie()

	commands := []string{"list", "copy", "clear"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	tests := []struct {
		prefix    string
		wantExact bool
		wantLen   int
	}{
		{"list", true, 1},  // Exact match
		{"l", false, 1},    // Prefix, one match
		{"c", false, 2},    // Prefix, multiple matches
		{"x", false, 0},    // No match
		{"clear", true, 1}, // Exact match
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			result := trie.Match(tt.prefix)

			if result.Exact != tt.wantExact {
				t.Errorf("Match(%q).Exact = %v, want %v", tt.prefix, result.Exact, tt.wantExact)
			}

			if len(result.Matches) != tt.wantLen {
				t.Errorf("Match(%q).Matches len = %d, want %d (matches: %v)",
					tt.prefix, len(result.Matches), tt.wantLen, result.Matches)
			}
		})
	}
}

func TestTrieCount(t *testing.T) {
	trie := NewTrie()

	if trie.Count() != 0 {
		t.Errorf("Count() on empty trie = %d, want 0", trie.Count())
	}

	commands := []string{"list", "copy", "clear", "cd"}
	for i, cmd := range commands {
		trie.Insert(cmd)
		if trie.Count() != i+1 {
			t.Errorf("Count() after inserting %d commands = %d, want %d", i+1, trie.Count(), i+1)
		}
	}
}

func TestTrieClear(t *testing.T) {
	trie := NewTrie()

	trie.Insert("list")
	trie.Insert("copy")

	if trie.Count() != 2 {
		t.Errorf("Count() = %d, want 2", trie.Count())
	}

	trie.Clear()

	if trie.Count() != 0 {
		t.Errorf("Count() after Clear() = %d, want 0", trie.Count())
	}

	if trie.ExactMatch("list") {
		t.Error("ExactMatch('list') should be false after Clear()")
	}
}

func TestTrieAll(t *testing.T) {
	trie := NewTrie()

	commands := []string{"list", "copy", "clear", "cd"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	all := trie.All()
	sort.Strings(all)
	sort.Strings(commands)

	if len(all) != len(commands) {
		t.Errorf("All() = %v, want %v", all, commands)
		return
	}

	for i := range all {
		if all[i] != commands[i] {
			t.Errorf("All() = %v, want %v", all, commands)
			return
		}
	}
}

func TestTrieDuplicateInsert(t *testing.T) {
	trie := NewTrie()

	trie.Insert("list")
	trie.Insert("list") // Duplicate

	if trie.Count() != 1 {
		t.Errorf("Count() after duplicate insert = %d, want 1", trie.Count())
	}
}

func TestTrieUnicodeSupport(t *testing.T) {
	trie := NewTrie()

	commands := []string{"läs", "déplacer", "копировать"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	// Test exact matches for unicode commands
	for _, cmd := range commands {
		if !trie.ExactMatch(cmd) {
			t.Errorf("ExactMatch(%q) = false, want true", cmd)
		}
	}

	// Test prefix search with unicode
	matches := trie.Search("lä")
	if len(matches) != 1 || matches[0] != "läs" {
		t.Errorf("Search('lä') = %v, want ['läs']", matches)
	}
}

func TestTrieEmptyString(t *testing.T) {
	trie := NewTrie()

	trie.Insert("list")
	trie.Insert("") // Empty string

	// Empty string should match everything
	matches := trie.Search("")
	if len(matches) != 2 { // "list" and ""
		t.Errorf("Search('') should return all commands, got %v", matches)
	}
}

// Benchmark tests
func BenchmarkTrieInsert(b *testing.B) {
	commands := []string{"list", "copy", "move", "cd", "clear", "echo", "env", "exit", "help", "pwd"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		trie := NewTrie()
		for _, cmd := range commands {
			trie.Insert(cmd)
		}
	}
}

func BenchmarkTrieSearch(b *testing.B) {
	trie := NewTrie()
	commands := []string{"list", "copy", "move", "cd", "clear", "echo", "env", "exit", "help", "pwd"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	prefixes := []string{"l", "c", "e", "h", "p", "m"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, p := range prefixes {
			trie.Search(p)
		}
	}
}

func BenchmarkTrieExactMatch(b *testing.B) {
	trie := NewTrie()
	commands := []string{"list", "copy", "move", "cd", "clear", "echo", "env", "exit", "help", "pwd"}
	for _, cmd := range commands {
		trie.Insert(cmd)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, cmd := range commands {
			trie.ExactMatch(cmd)
		}
	}
}
