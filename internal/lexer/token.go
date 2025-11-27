// Package lexer provides tokenization for shell input.
package lexer

// TokenType identifies the type of token.
type TokenType int

const (
	TokenWord       TokenType = iota // Command or argument word
	TokenString                      // Quoted string ("..." or '...')
	TokenOption                      // Option (--flag or -f)
	TokenEquals                      // Assignment operator (=)
	TokenVariable                    // Variable reference ($VAR or ${VAR})
	TokenWhitespace                  // Whitespace (space or tab)
	TokenNewline                     // Newline character
	TokenEOF                         // End of input
	TokenError                       // Lexer error
)

// String returns the string representation of a TokenType.
func (t TokenType) String() string {
	switch t {
	case TokenWord:
		return "WORD"
	case TokenString:
		return "STRING"
	case TokenOption:
		return "OPTION"
	case TokenEquals:
		return "EQUALS"
	case TokenVariable:
		return "VARIABLE"
	case TokenWhitespace:
		return "WHITESPACE"
	case TokenNewline:
		return "NEWLINE"
	case TokenEOF:
		return "EOF"
	case TokenError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Position represents a location in the source input.
type Position struct {
	Line   int // Line number (1-indexed)
	Column int // Column number (1-indexed)
	Offset int // Byte offset from start
}

// Token represents a lexical token from the input.
type Token struct {
	Type    TokenType // Type of the token
	Value   string    // Raw value from input
	Literal string    // Processed value (escapes applied, quotes removed)
	Pos     Position  // Position in source
}

// String returns a debug representation of the token.
func (t Token) String() string {
	return t.Type.String() + "(" + t.Value + ")"
}

// IsWord returns true if the token is a word or string (usable as argument).
func (t Token) IsWord() bool {
	return t.Type == TokenWord || t.Type == TokenString
}

// IsOption returns true if the token is an option flag.
func (t Token) IsOption() bool {
	return t.Type == TokenOption
}

// IsWhitespace returns true if the token is whitespace or newline.
func (t Token) IsWhitespace() bool {
	return t.Type == TokenWhitespace || t.Type == TokenNewline
}
