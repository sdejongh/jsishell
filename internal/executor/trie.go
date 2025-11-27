package executor

// Trie is a prefix tree for efficient command name lookup and abbreviation resolution.
// It provides O(k) lookup time where k is the length of the search key,
// compared to O(n*k) for naive prefix matching.
type Trie struct {
	root *trieNode
}

// trieNode represents a node in the trie.
type trieNode struct {
	children map[rune]*trieNode
	isEnd    bool   // True if this node marks the end of a command name
	command  string // The full command name if isEnd is true
}

// NewTrie creates a new empty Trie.
func NewTrie() *Trie {
	return &Trie{
		root: newTrieNode(),
	}
}

// newTrieNode creates a new trie node.
func newTrieNode() *trieNode {
	return &trieNode{
		children: make(map[rune]*trieNode),
	}
}

// Insert adds a command name to the trie.
func (t *Trie) Insert(command string) {
	node := t.root
	for _, ch := range command {
		if _, exists := node.children[ch]; !exists {
			node.children[ch] = newTrieNode()
		}
		node = node.children[ch]
	}
	node.isEnd = true
	node.command = command
}

// Search finds all command names that start with the given prefix.
// Returns the matching command names in no particular order.
func (t *Trie) Search(prefix string) []string {
	node := t.root

	// Navigate to the node representing the prefix
	for _, ch := range prefix {
		if child, exists := node.children[ch]; exists {
			node = child
		} else {
			return nil // No matches
		}
	}

	// Collect all commands from this node onwards
	var results []string
	t.collectCommands(node, &results)
	return results
}

// collectCommands recursively collects all command names from a node.
func (t *Trie) collectCommands(node *trieNode, results *[]string) {
	if node.isEnd {
		*results = append(*results, node.command)
	}
	for _, child := range node.children {
		t.collectCommands(child, results)
	}
}

// ExactMatch checks if the exact command name exists.
func (t *Trie) ExactMatch(command string) bool {
	node := t.root
	for _, ch := range command {
		if child, exists := node.children[ch]; exists {
			node = child
		} else {
			return false
		}
	}
	return node.isEnd
}

// MatchResult represents the result of a prefix match operation.
type MatchResult struct {
	Exact   bool     // True if prefix is an exact match for a command
	Matches []string // All commands that start with the prefix
}

// Match performs a comprehensive prefix match and returns detailed results.
func (t *Trie) Match(prefix string) MatchResult {
	node := t.root

	// Navigate to the node representing the prefix
	for _, ch := range prefix {
		if child, exists := node.children[ch]; exists {
			node = child
		} else {
			return MatchResult{Exact: false, Matches: nil}
		}
	}

	// Collect all commands from this node onwards
	var matches []string
	t.collectCommands(node, &matches)

	return MatchResult{
		Exact:   node.isEnd,
		Matches: matches,
	}
}

// Count returns the total number of commands in the trie.
func (t *Trie) Count() int {
	return t.countNodes(t.root)
}

// countNodes recursively counts command end nodes.
func (t *Trie) countNodes(node *trieNode) int {
	count := 0
	if node.isEnd {
		count = 1
	}
	for _, child := range node.children {
		count += t.countNodes(child)
	}
	return count
}

// Clear removes all commands from the trie.
func (t *Trie) Clear() {
	t.root = newTrieNode()
}

// All returns all command names in the trie.
func (t *Trie) All() []string {
	var results []string
	t.collectCommands(t.root, &results)
	return results
}
