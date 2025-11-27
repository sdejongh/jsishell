package parser

import (
	"fmt"
	"strings"

	"github.com/sdejongh/jsishell/internal/env"
	"github.com/sdejongh/jsishell/internal/errors"
	"github.com/sdejongh/jsishell/internal/lexer"
)

// Parser parses tokens into a Command.
type Parser struct {
	tokens []lexer.Token
	pos    int
	env    *env.Environment
}

// New creates a new Parser for the given tokens.
func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
	}
}

// NewWithEnv creates a new Parser with environment for variable expansion.
func NewWithEnv(tokens []lexer.Token, e *env.Environment) *Parser {
	return &Parser{
		tokens: tokens,
		pos:    0,
		env:    e,
	}
}

// Parse parses the tokens into a Command.
func (p *Parser) Parse() (*Command, error) {
	cmd := NewCommand()

	// Skip leading whitespace
	p.skipWhitespace()

	// Check for empty input
	if p.current().Type == lexer.TokenEOF {
		return nil, nil // Empty command
	}

	// First non-whitespace token should be the command name
	nameTok := p.current()
	if nameTok.Type == lexer.TokenError {
		return nil, fmt.Errorf("%w: %s", errors.ErrInvalidSyntax, nameTok.Literal)
	}

	cmd.Name = p.expandValue(nameTok)
	p.advance()

	// Parse remaining tokens as arguments, options, and flags
	for {
		p.skipWhitespace()
		tok := p.current()

		switch tok.Type {
		case lexer.TokenEOF, lexer.TokenNewline:
			return cmd, nil

		case lexer.TokenError:
			return nil, fmt.Errorf("%w: %s", errors.ErrInvalidSyntax, tok.Literal)

		case lexer.TokenOption:
			if err := p.parseOption(cmd, tok); err != nil {
				return nil, err
			}

		case lexer.TokenWord, lexer.TokenString, lexer.TokenVariable:
			cmd.Args = append(cmd.Args, p.expandValue(tok))
			p.advance()

		case lexer.TokenEquals:
			// Unexpected equals - treat as argument
			cmd.Args = append(cmd.Args, tok.Literal)
			p.advance()

		default:
			p.advance() // Skip unknown tokens
		}
	}
}

// parseOption parses an option token and its potential value.
func (p *Parser) parseOption(cmd *Command, tok lexer.Token) error {
	optName := tok.Value

	// Check if this is --key=value format
	if idx := strings.Index(optName, "="); idx != -1 {
		key := optName[:idx]
		value := optName[idx+1:]
		cmd.Options[key] = value
		p.advance()
		return nil
	}

	p.advance()

	// Check if this is a long option (--something)
	if strings.HasPrefix(optName, "--") {
		return p.parseLongOption(cmd, optName)
	}

	// Short option(s): -a or -abc (combined)
	// Expand combined short options into individual flags
	// e.g., -al becomes -a and -l
	if len(optName) > 2 {
		// Combined short options: -abc -> -a, -b, -c
		for _, ch := range optName[1:] {
			flag := "-" + string(ch)
			cmd.Flags[flag] = true
		}
	} else {
		// Single short option: -a
		cmd.Flags[optName] = true
	}

	return nil
}

// parseLongOption parses a long option (--something).
func (p *Parser) parseLongOption(cmd *Command, optName string) error {
	// Check for following = or value
	p.skipWhitespace()
	next := p.current()

	if next.Type == lexer.TokenEquals {
		// --key = value or --key =value
		p.advance()
		p.skipWhitespace()
		valueTok := p.current()

		if valueTok.Type == lexer.TokenWord || valueTok.Type == lexer.TokenString || valueTok.Type == lexer.TokenVariable {
			cmd.Options[optName] = p.expandValue(valueTok)
			p.advance()
		} else {
			// --key= with no value
			cmd.Options[optName] = ""
		}
		return nil
	}

	// Treat as a flag
	cmd.Flags[optName] = true
	return nil
}

// expandValue expands variables in a token value.
func (p *Parser) expandValue(tok lexer.Token) string {
	if tok.Type == lexer.TokenVariable {
		if p.env != nil {
			return p.env.Get(tok.Literal)
		}
		return "" // No environment, return empty
	}
	return tok.Literal
}

// current returns the current token.
func (p *Parser) current() lexer.Token {
	if p.pos >= len(p.tokens) {
		return lexer.Token{Type: lexer.TokenEOF}
	}
	return p.tokens[p.pos]
}

// advance moves to the next token.
func (p *Parser) advance() {
	p.pos++
}

// skipWhitespace skips whitespace tokens.
func (p *Parser) skipWhitespace() {
	for p.current().IsWhitespace() {
		p.advance()
	}
}

// ParseInput is a convenience function that lexes and parses input.
func ParseInput(input string) (*Command, error) {
	l := lexer.New(input)
	tokens := l.Tokens()
	parser := New(tokens)
	cmd, err := parser.Parse()
	if cmd != nil {
		cmd.RawInput = input
	}
	return cmd, err
}

// ParseInputWithEnv parses input with environment for variable expansion.
func ParseInputWithEnv(input string, e *env.Environment) (*Command, error) {
	l := lexer.New(input)
	tokens := l.Tokens()
	parser := NewWithEnv(tokens, e)
	cmd, err := parser.Parse()
	if cmd != nil {
		cmd.RawInput = input
	}
	return cmd, err
}
