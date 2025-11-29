package lexer

import (
	"testing"
)

func TestTokenTypeString(t *testing.T) {
	tests := []struct {
		tt   TokenType
		want string
	}{
		{TokenWord, "WORD"},
		{TokenString, "STRING"},
		{TokenOption, "OPTION"},
		{TokenEquals, "EQUALS"},
		{TokenVariable, "VARIABLE"},
		{TokenWhitespace, "WHITESPACE"},
		{TokenNewline, "NEWLINE"},
		{TokenEOF, "EOF"},
		{TokenError, "ERROR"},
		{TokenType(999), "UNKNOWN"},
	}

	for _, tt := range tests {
		got := tt.tt.String()
		if got != tt.want {
			t.Errorf("TokenType(%d).String() = %q, want %q", tt.tt, got, tt.want)
		}
	}
}

func TestLexerSimpleWords(t *testing.T) {
	tests := []struct {
		input string
		want  []Token
	}{
		{
			"list",
			[]Token{
				{Type: TokenWord, Value: "list", Literal: "list"},
				{Type: TokenEOF},
			},
		},
		{
			"cd /home",
			[]Token{
				{Type: TokenWord, Value: "cd", Literal: "cd"},
				{Type: TokenWhitespace, Value: " "},
				{Type: TokenWord, Value: "/home", Literal: "/home"},
				{Type: TokenEOF},
			},
		},
		{
			"copy file.txt backup/",
			[]Token{
				{Type: TokenWord, Value: "copy", Literal: "copy"},
				{Type: TokenWhitespace, Value: " "},
				{Type: TokenWord, Value: "file.txt", Literal: "file.txt"},
				{Type: TokenWhitespace, Value: " "},
				{Type: TokenWord, Value: "backup/", Literal: "backup/"},
				{Type: TokenEOF},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tokens := l.Tokens()

			if len(tokens) != len(tt.want) {
				t.Fatalf("got %d tokens, want %d", len(tokens), len(tt.want))
			}

			for i, want := range tt.want {
				got := tokens[i]
				if got.Type != want.Type {
					t.Errorf("token[%d].Type = %v, want %v", i, got.Type, want.Type)
				}
				if want.Value != "" && got.Value != want.Value {
					t.Errorf("token[%d].Value = %q, want %q", i, got.Value, want.Value)
				}
				if want.Literal != "" && got.Literal != want.Literal {
					t.Errorf("token[%d].Literal = %q, want %q", i, got.Literal, want.Literal)
				}
			}
		})
	}
}

func TestLexerQuotedStrings(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantTok Token
	}{
		{
			"double quoted simple",
			`"hello world"`,
			Token{Type: TokenString, Value: `"hello world"`, Literal: "hello world"},
		},
		{
			"double quoted escapes",
			`"hello\nworld"`,
			Token{Type: TokenString, Value: `"hello\nworld"`, Literal: "hello\nworld"},
		},
		{
			"double quoted escaped quote",
			`"say \"hello\""`,
			Token{Type: TokenString, Value: `"say \"hello\""`, Literal: `say "hello"`},
		},
		{
			"single quoted simple",
			`'hello world'`,
			Token{Type: TokenString, Value: `'hello world'`, Literal: "hello world"},
		},
		{
			"single quoted with backslash",
			`'hello\nworld'`,
			Token{Type: TokenString, Value: `'hello\nworld'`, Literal: `hello\nworld`},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.wantTok.Type {
				t.Errorf("Type = %v, want %v", tok.Type, tt.wantTok.Type)
			}
			if tok.Value != tt.wantTok.Value {
				t.Errorf("Value = %q, want %q", tok.Value, tt.wantTok.Value)
			}
			if tok.Literal != tt.wantTok.Literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.wantTok.Literal)
			}
		})
	}
}

func TestLexerUnterminatedString(t *testing.T) {
	tests := []struct {
		name  string
		input string
	}{
		{"unterminated double", `"hello`},
		{"unterminated single", `'hello`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != TokenError {
				t.Errorf("Type = %v, want TokenError", tok.Type)
			}
		})
	}
}

func TestLexerOptions(t *testing.T) {
	tests := []struct {
		input string
		want  Token
	}{
		{"-v", Token{Type: TokenOption, Value: "-v", Literal: "-v"}},
		{"-a", Token{Type: TokenOption, Value: "-a", Literal: "-a"}},
		{"--verbose", Token{Type: TokenOption, Value: "--verbose", Literal: "--verbose"}},
		{"--all", Token{Type: TokenOption, Value: "--all", Literal: "--all"}},
		{"--no-color", Token{Type: TokenOption, Value: "--no-color", Literal: "--no-color"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", tok.Type, tt.want.Type)
			}
			if tok.Value != tt.want.Value {
				t.Errorf("Value = %q, want %q", tok.Value, tt.want.Value)
			}
		})
	}
}

func TestLexerDashAlone(t *testing.T) {
	l := New("-")
	tok := l.NextToken()

	if tok.Type != TokenWord {
		t.Errorf("Type = %v, want TokenWord", tok.Type)
	}
	if tok.Value != "-" {
		t.Errorf("Value = %q, want \"-\"", tok.Value)
	}
}

func TestLexerVariables(t *testing.T) {
	tests := []struct {
		input string
		want  Token
	}{
		{"$HOME", Token{Type: TokenVariable, Value: "$HOME", Literal: "HOME"}},
		{"$PATH", Token{Type: TokenVariable, Value: "$PATH", Literal: "PATH"}},
		{"${HOME}", Token{Type: TokenVariable, Value: "${HOME}", Literal: "HOME"}},
		{"${USER}", Token{Type: TokenVariable, Value: "${USER}", Literal: "USER"}},
		{"$VAR_NAME", Token{Type: TokenVariable, Value: "$VAR_NAME", Literal: "VAR_NAME"}},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Type != tt.want.Type {
				t.Errorf("Type = %v, want %v", tok.Type, tt.want.Type)
			}
			if tok.Value != tt.want.Value {
				t.Errorf("Value = %q, want %q", tok.Value, tt.want.Value)
			}
			if tok.Literal != tt.want.Literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.want.Literal)
			}
		})
	}
}

func TestLexerDollarAlone(t *testing.T) {
	l := New("$ ")
	tok := l.NextToken()

	if tok.Type != TokenWord {
		t.Errorf("Type = %v, want TokenWord", tok.Type)
	}
	if tok.Value != "$" {
		t.Errorf("Value = %q, want \"$\"", tok.Value)
	}
}

func TestLexerEquals(t *testing.T) {
	l := New("VAR=value")
	tokens := l.Tokens()

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokenWord, "VAR"},
		{TokenEquals, "="},
		{TokenWord, "value"},
		{TokenEOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d].Type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
		if exp.val != "" && tokens[i].Value != exp.val {
			t.Errorf("token[%d].Value = %q, want %q", i, tokens[i].Value, exp.val)
		}
	}
}

func TestLexerWhitespace(t *testing.T) {
	l := New("a  \t  b")
	tokens := l.Tokens()

	if len(tokens) != 4 { // a, whitespace, b, EOF
		t.Fatalf("got %d tokens, want 4", len(tokens))
	}

	if tokens[1].Type != TokenWhitespace {
		t.Errorf("token[1].Type = %v, want TokenWhitespace", tokens[1].Type)
	}
}

func TestLexerNewline(t *testing.T) {
	l := New("a\nb")
	tokens := l.Tokens()

	if len(tokens) != 4 { // a, newline, b, EOF
		t.Fatalf("got %d tokens, want 4", len(tokens))
	}

	if tokens[1].Type != TokenNewline {
		t.Errorf("token[1].Type = %v, want TokenNewline", tokens[1].Type)
	}
}

func TestLexerComplexCommand(t *testing.T) {
	input := `list --all -l "/path with spaces" $HOME`

	l := New(input)
	tokens := l.Tokens()

	expected := []struct {
		typ TokenType
		val string
	}{
		{TokenWord, "list"},
		{TokenWhitespace, " "},
		{TokenOption, "--all"},
		{TokenWhitespace, " "},
		{TokenOption, "-l"},
		{TokenWhitespace, " "},
		{TokenString, `"/path with spaces"`},
		{TokenWhitespace, " "},
		{TokenVariable, "$HOME"},
		{TokenEOF, ""},
	}

	if len(tokens) != len(expected) {
		t.Fatalf("got %d tokens, want %d", len(tokens), len(expected))
	}

	for i, exp := range expected {
		if tokens[i].Type != exp.typ {
			t.Errorf("token[%d].Type = %v, want %v", i, tokens[i].Type, exp.typ)
		}
	}
}

func TestLexerPosition(t *testing.T) {
	l := New("abc def")

	tok1 := l.NextToken()
	if tok1.Pos.Line != 1 || tok1.Pos.Column != 1 {
		t.Errorf("token1 pos = %d:%d, want 1:1", tok1.Pos.Line, tok1.Pos.Column)
	}

	l.NextToken() // whitespace

	tok2 := l.NextToken()
	if tok2.Pos.Line != 1 || tok2.Pos.Column != 5 {
		t.Errorf("token2 pos = %d:%d, want 1:5", tok2.Pos.Line, tok2.Pos.Column)
	}
}

func TestLexerMultiLine(t *testing.T) {
	l := New("abc\ndef")
	tokens := l.Tokens()

	// Find the 'def' token
	var defTok Token
	for _, tok := range tokens {
		if tok.Value == "def" {
			defTok = tok
			break
		}
	}

	if defTok.Pos.Line != 2 {
		t.Errorf("'def' token line = %d, want 2", defTok.Pos.Line)
	}
}

func TestTokenIsWord(t *testing.T) {
	tests := []struct {
		tok  Token
		want bool
	}{
		{Token{Type: TokenWord}, true},
		{Token{Type: TokenString}, true},
		{Token{Type: TokenOption}, false},
		{Token{Type: TokenVariable}, false},
	}

	for _, tt := range tests {
		if got := tt.tok.IsWord(); got != tt.want {
			t.Errorf("Token{Type: %v}.IsWord() = %v, want %v", tt.tok.Type, got, tt.want)
		}
	}
}

func TestTokenIsOption(t *testing.T) {
	tests := []struct {
		tok  Token
		want bool
	}{
		{Token{Type: TokenOption}, true},
		{Token{Type: TokenWord}, false},
	}

	for _, tt := range tests {
		if got := tt.tok.IsOption(); got != tt.want {
			t.Errorf("Token{Type: %v}.IsOption() = %v, want %v", tt.tok.Type, got, tt.want)
		}
	}
}

func TestTokenIsWhitespace(t *testing.T) {
	tests := []struct {
		tok  Token
		want bool
	}{
		{Token{Type: TokenWhitespace}, true},
		{Token{Type: TokenNewline}, true},
		{Token{Type: TokenWord}, false},
	}

	for _, tt := range tests {
		if got := tt.tok.IsWhitespace(); got != tt.want {
			t.Errorf("Token{Type: %v}.IsWhitespace() = %v, want %v", tt.tok.Type, got, tt.want)
		}
	}
}

func TestLexerBackslashInWord(t *testing.T) {
	// In unquoted words, backslash is NOT an escape character.
	// All backslashes are preserved literally (for Windows paths).
	// For filenames with spaces, use quotes: "file name"
	tests := []struct {
		input   string
		literal string
	}{
		{`path\\file`, `path\\file`},       // Double backslash preserved
		{`a\tb`, `a\tb`},                   // \t preserved (not tab)
		{`C:\Users\test`, `C:\Users\test`}, // Windows path preserved
		{`d:\pictures\`, `d:\pictures\`},   // Windows path with trailing backslash
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			l := New(tt.input)
			tok := l.NextToken()

			if tok.Literal != tt.literal {
				t.Errorf("Literal = %q, want %q", tok.Literal, tt.literal)
			}
		})
	}
}
