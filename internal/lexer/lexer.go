package lexer

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

// Lexer tokenizes shell input.
type Lexer struct {
	input   string
	pos     int  // Current position in input (byte index)
	readPos int  // Next position to read
	ch      rune // Current character
	line    int  // Current line number (1-indexed)
	col     int  // Current column number (1-indexed)
}

// New creates a new Lexer for the given input.
func New(input string) *Lexer {
	l := &Lexer{
		input: input,
		line:  1,
		col:   0,
	}
	l.readChar()
	return l
}

// readChar advances to the next character.
func (l *Lexer) readChar() {
	if l.readPos >= len(l.input) {
		l.ch = 0 // EOF
	} else {
		l.ch, _ = utf8.DecodeRuneInString(l.input[l.readPos:])
	}
	l.pos = l.readPos
	l.readPos += utf8.RuneLen(l.ch)
	if l.ch == 0 {
		l.readPos = len(l.input) + 1 // Prevent further reading
	}

	// Track position
	if l.ch == '\n' {
		l.line++
		l.col = 0
	} else {
		l.col++
	}
}

// peekChar returns the next character without advancing.
func (l *Lexer) peekChar() rune {
	if l.readPos >= len(l.input) {
		return 0
	}
	ch, _ := utf8.DecodeRuneInString(l.input[l.readPos:])
	return ch
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	startPos := Position{Line: l.line, Column: l.col, Offset: l.pos}

	switch {
	case l.ch == 0:
		return Token{Type: TokenEOF, Pos: startPos}

	case l.ch == '\n':
		l.readChar()
		return Token{Type: TokenNewline, Value: "\n", Literal: "\n", Pos: startPos}

	case unicode.IsSpace(l.ch):
		return l.readWhitespace(startPos)

	case l.ch == '"':
		return l.readDoubleQuotedString(startPos)

	case l.ch == '\'':
		return l.readSingleQuotedString(startPos)

	case l.ch == '$':
		return l.readVariable(startPos)

	case l.ch == '=':
		l.readChar()
		return Token{Type: TokenEquals, Value: "=", Literal: "=", Pos: startPos}

	case l.ch == '-':
		return l.readOption(startPos)

	default:
		return l.readWord(startPos)
	}
}

// Tokens returns all tokens from the input.
func (l *Lexer) Tokens() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF || tok.Type == TokenError {
			break
		}
	}
	return tokens
}

// readWhitespace reads whitespace characters (not newlines).
func (l *Lexer) readWhitespace(startPos Position) Token {
	start := l.pos
	for unicode.IsSpace(l.ch) && l.ch != '\n' {
		l.readChar()
	}
	value := l.input[start:l.pos]
	return Token{Type: TokenWhitespace, Value: value, Literal: value, Pos: startPos}
}

// readDoubleQuotedString reads a double-quoted string.
// Supports escape sequences: \\, \", \n, \t, \$
func (l *Lexer) readDoubleQuotedString(startPos Position) Token {
	l.readChar() // Skip opening quote
	var literal strings.Builder
	start := l.pos

	for l.ch != '"' && l.ch != 0 {
		if l.ch == '\\' {
			next := l.peekChar()
			switch next {
			case '\\':
				literal.WriteRune('\\')
				l.readChar()
				l.readChar()
			case '"':
				literal.WriteRune('"')
				l.readChar()
				l.readChar()
			case 'n':
				literal.WriteRune('\n')
				l.readChar()
				l.readChar()
			case 't':
				literal.WriteRune('\t')
				l.readChar()
				l.readChar()
			case '$':
				literal.WriteRune('$')
				l.readChar()
				l.readChar()
			default:
				// Unknown escape, keep backslash
				literal.WriteRune(l.ch)
				l.readChar()
			}
		} else {
			literal.WriteRune(l.ch)
			l.readChar()
		}
	}

	if l.ch == 0 {
		// Unterminated string
		return Token{
			Type:    TokenError,
			Value:   l.input[start-1 : l.pos],
			Literal: "unterminated string",
			Pos:     startPos,
		}
	}

	value := l.input[start-1 : l.pos+1] // Include quotes in value
	l.readChar()                        // Skip closing quote

	return Token{Type: TokenString, Value: value, Literal: literal.String(), Pos: startPos}
}

// readSingleQuotedString reads a single-quoted string.
// No escape sequences are processed (raw string).
func (l *Lexer) readSingleQuotedString(startPos Position) Token {
	l.readChar() // Skip opening quote
	start := l.pos

	for l.ch != '\'' && l.ch != 0 {
		l.readChar()
	}

	if l.ch == 0 {
		// Unterminated string
		return Token{
			Type:    TokenError,
			Value:   l.input[start-1 : l.pos],
			Literal: "unterminated string",
			Pos:     startPos,
		}
	}

	literal := l.input[start:l.pos]
	value := l.input[start-1 : l.pos+1] // Include quotes in value
	l.readChar()                        // Skip closing quote

	return Token{Type: TokenString, Value: value, Literal: literal, Pos: startPos}
}

// readVariable reads a variable reference ($VAR or ${VAR}).
func (l *Lexer) readVariable(startPos Position) Token {
	start := l.pos
	l.readChar() // Skip $

	if l.ch == '{' {
		// ${VAR} form
		l.readChar() // Skip {
		varStart := l.pos

		for isIdentChar(l.ch) {
			l.readChar()
		}

		if l.ch != '}' {
			return Token{
				Type:    TokenError,
				Value:   l.input[start:l.pos],
				Literal: "unterminated variable",
				Pos:     startPos,
			}
		}

		varName := l.input[varStart:l.pos]
		value := l.input[start : l.pos+1]
		l.readChar() // Skip }

		return Token{Type: TokenVariable, Value: value, Literal: varName, Pos: startPos}
	}

	// $VAR form
	for isIdentChar(l.ch) {
		l.readChar()
	}

	value := l.input[start:l.pos]
	varName := value[1:] // Remove $

	if varName == "" {
		// Just $ alone - treat as word
		return Token{Type: TokenWord, Value: "$", Literal: "$", Pos: startPos}
	}

	return Token{Type: TokenVariable, Value: value, Literal: varName, Pos: startPos}
}

// readOption reads an option (--flag or -abc for combined short options).
func (l *Lexer) readOption(startPos Position) Token {
	start := l.pos
	l.readChar() // Skip first -

	if l.ch == '-' {
		// Long option --flag
		l.readChar() // Skip second -
		for isOptionChar(l.ch) {
			l.readChar()
		}
	} else if isLetter(l.ch) || unicode.IsDigit(l.ch) {
		// Short option(s) -f or -abc or -123
		// Read all alphanumeric characters as part of the option
		for isLetter(l.ch) || unicode.IsDigit(l.ch) {
			l.readChar()
		}
	} else {
		// Just - alone, treat as word
		return Token{Type: TokenWord, Value: "-", Literal: "-", Pos: startPos}
	}

	value := l.input[start:l.pos]
	return Token{Type: TokenOption, Value: value, Literal: value, Pos: startPos}
}

// readWord reads a word (command name or argument).
func (l *Lexer) readWord(startPos Position) Token {
	start := l.pos

	for l.ch != 0 && !isWordTerminator(l.ch) {
		// All characters including backslash are kept literally.
		// Backslash is NOT an escape character in unquoted words.
		// This ensures Windows paths like C:\Users\name work correctly.
		// For filenames with spaces, use quotes: "file name" or 'file name'
		l.readChar()
	}

	value := l.input[start:l.pos]
	return Token{Type: TokenWord, Value: value, Literal: value, Pos: startPos}
}

// isIdentChar returns true if ch is a valid identifier character.
func isIdentChar(ch rune) bool {
	return isLetter(ch) || unicode.IsDigit(ch) || ch == '_'
}

// isLetter returns true if ch is a letter.
func isLetter(ch rune) bool {
	return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')
}

// isOptionChar returns true if ch is valid in an option name.
func isOptionChar(ch rune) bool {
	return isIdentChar(ch) || ch == '-'
}

// isWordTerminator returns true if ch terminates a word.
func isWordTerminator(ch rune) bool {
	return unicode.IsSpace(ch) || ch == '"' || ch == '\'' || ch == '$' || ch == '='
}
