package builtins

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// SearchExpr represents a search expression that can be evaluated against a filename.
type SearchExpr interface {
	// Evaluate returns true if the filename matches the expression.
	Evaluate(filename string) bool
	// String returns a string representation of the expression.
	String() string
}

// PatternExpr represents a glob pattern.
type PatternExpr struct {
	Pattern string
}

func (p *PatternExpr) Evaluate(filename string) bool {
	matched, err := filepath.Match(p.Pattern, filename)
	if err != nil {
		return false
	}
	return matched
}

func (p *PatternExpr) String() string {
	return fmt.Sprintf("%q", p.Pattern)
}

// NotExpr represents a NOT expression.
type NotExpr struct {
	Expr SearchExpr
}

func (n *NotExpr) Evaluate(filename string) bool {
	return !n.Expr.Evaluate(filename)
}

func (n *NotExpr) String() string {
	return fmt.Sprintf("NOT(%s)", n.Expr.String())
}

// AndExpr represents an AND expression.
type AndExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (a *AndExpr) Evaluate(filename string) bool {
	return a.Left.Evaluate(filename) && a.Right.Evaluate(filename)
}

func (a *AndExpr) String() string {
	return fmt.Sprintf("(%s AND %s)", a.Left.String(), a.Right.String())
}

// OrExpr represents an OR expression.
type OrExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (o *OrExpr) Evaluate(filename string) bool {
	return o.Left.Evaluate(filename) || o.Right.Evaluate(filename)
}

func (o *OrExpr) String() string {
	return fmt.Sprintf("(%s OR %s)", o.Left.String(), o.Right.String())
}

// XorExpr represents an XOR expression.
type XorExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (x *XorExpr) Evaluate(filename string) bool {
	left := x.Left.Evaluate(filename)
	right := x.Right.Evaluate(filename)
	return (left || right) && !(left && right)
}

func (x *XorExpr) String() string {
	return fmt.Sprintf("(%s XOR %s)", x.Left.String(), x.Right.String())
}

// Token types for the expression parser
type exprTokenType int

const (
	tokPattern exprTokenType = iota
	tokAnd
	tokOr
	tokXor
	tokNot
	tokLParen
	tokRParen
	tokEOF
)

type exprToken struct {
	typ   exprTokenType
	value string
}

// exprLexer tokenizes the search expression arguments.
type exprLexer struct {
	args    []string
	pos     int
	tokens  []exprToken
	current int
}

func newExprLexer(args []string) *exprLexer {
	return &exprLexer{
		args:    args,
		pos:     0,
		tokens:  nil,
		current: 0,
	}
}

// tokenize converts arguments to tokens.
func (l *exprLexer) tokenize() error {
	l.tokens = nil

	for l.pos < len(l.args) {
		arg := l.args[l.pos]
		l.pos++

		// Handle operators (case-insensitive)
		upper := strings.ToUpper(arg)
		switch upper {
		case "AND", "&&":
			l.tokens = append(l.tokens, exprToken{typ: tokAnd, value: "AND"})
		case "OR", "||":
			l.tokens = append(l.tokens, exprToken{typ: tokOr, value: "OR"})
		case "XOR", "^":
			l.tokens = append(l.tokens, exprToken{typ: tokXor, value: "XOR"})
		case "NOT", "!":
			l.tokens = append(l.tokens, exprToken{typ: tokNot, value: "NOT"})
		default:
			// Check for parentheses embedded in the argument
			if err := l.tokenizeWithParens(arg); err != nil {
				return err
			}
		}
	}

	l.tokens = append(l.tokens, exprToken{typ: tokEOF, value: ""})
	return nil
}

// tokenizeWithParens handles arguments that may contain parentheses.
func (l *exprLexer) tokenizeWithParens(arg string) error {
	// Process character by character for parentheses
	i := 0
	for i < len(arg) {
		ch := rune(arg[i])

		if ch == '(' {
			l.tokens = append(l.tokens, exprToken{typ: tokLParen, value: "("})
			i++
		} else if ch == ')' {
			l.tokens = append(l.tokens, exprToken{typ: tokRParen, value: ")"})
			i++
		} else if unicode.IsSpace(ch) {
			i++
		} else {
			// Read until next paren or end
			start := i
			for i < len(arg) && arg[i] != '(' && arg[i] != ')' {
				i++
			}
			pattern := strings.TrimSpace(arg[start:i])
			if pattern != "" {
				// Check if it's an operator
				upper := strings.ToUpper(pattern)
				switch upper {
				case "AND", "&&":
					l.tokens = append(l.tokens, exprToken{typ: tokAnd, value: "AND"})
				case "OR", "||":
					l.tokens = append(l.tokens, exprToken{typ: tokOr, value: "OR"})
				case "XOR", "^":
					l.tokens = append(l.tokens, exprToken{typ: tokXor, value: "XOR"})
				case "NOT", "!":
					l.tokens = append(l.tokens, exprToken{typ: tokNot, value: "NOT"})
				default:
					l.tokens = append(l.tokens, exprToken{typ: tokPattern, value: pattern})
				}
			}
		}
	}
	return nil
}

func (l *exprLexer) peek() exprToken {
	if l.current >= len(l.tokens) {
		return exprToken{typ: tokEOF}
	}
	return l.tokens[l.current]
}

func (l *exprLexer) next() exprToken {
	tok := l.peek()
	if l.current < len(l.tokens) {
		l.current++
	}
	return tok
}

// exprParser parses tokenized search expressions.
type exprParser struct {
	lexer *exprLexer
}

func newExprParser(lexer *exprLexer) *exprParser {
	return &exprParser{lexer: lexer}
}

// Parse parses the expression and returns the AST root.
// Grammar:
//
//	expr     = orExpr
//	orExpr   = xorExpr (("OR" | "||") xorExpr)*
//	xorExpr  = andExpr (("XOR" | "^") andExpr)*
//	andExpr  = unaryExpr (("AND" | "&&") unaryExpr)*
//	unaryExpr = ("NOT" | "!") unaryExpr | primary
//	primary  = pattern | "(" expr ")"
func (p *exprParser) Parse() (SearchExpr, error) {
	expr, err := p.parseOrExpr()
	if err != nil {
		return nil, err
	}

	// Check for unexpected tokens
	if tok := p.lexer.peek(); tok.typ != tokEOF {
		return nil, fmt.Errorf("unexpected token: %s", tok.value)
	}

	return expr, nil
}

func (p *exprParser) parseOrExpr() (SearchExpr, error) {
	left, err := p.parseXorExpr()
	if err != nil {
		return nil, err
	}

	for p.lexer.peek().typ == tokOr {
		p.lexer.next() // consume OR
		right, err := p.parseXorExpr()
		if err != nil {
			return nil, err
		}
		left = &OrExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *exprParser) parseXorExpr() (SearchExpr, error) {
	left, err := p.parseAndExpr()
	if err != nil {
		return nil, err
	}

	for p.lexer.peek().typ == tokXor {
		p.lexer.next() // consume XOR
		right, err := p.parseAndExpr()
		if err != nil {
			return nil, err
		}
		left = &XorExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *exprParser) parseAndExpr() (SearchExpr, error) {
	left, err := p.parseUnaryExpr()
	if err != nil {
		return nil, err
	}

	for p.lexer.peek().typ == tokAnd {
		p.lexer.next() // consume AND
		right, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		left = &AndExpr{Left: left, Right: right}
	}

	return left, nil
}

func (p *exprParser) parseUnaryExpr() (SearchExpr, error) {
	if p.lexer.peek().typ == tokNot {
		p.lexer.next() // consume NOT
		expr, err := p.parseUnaryExpr()
		if err != nil {
			return nil, err
		}
		return &NotExpr{Expr: expr}, nil
	}

	return p.parsePrimary()
}

func (p *exprParser) parsePrimary() (SearchExpr, error) {
	tok := p.lexer.peek()

	switch tok.typ {
	case tokLParen:
		p.lexer.next() // consume (
		expr, err := p.parseOrExpr()
		if err != nil {
			return nil, err
		}
		if p.lexer.peek().typ != tokRParen {
			return nil, fmt.Errorf("expected ')' but got: %s", p.lexer.peek().value)
		}
		p.lexer.next() // consume )
		return expr, nil

	case tokPattern:
		p.lexer.next()
		return &PatternExpr{Pattern: tok.value}, nil

	case tokEOF:
		return nil, fmt.Errorf("unexpected end of expression")

	default:
		return nil, fmt.Errorf("unexpected token: %s", tok.value)
	}
}

// ParseSearchExpression parses a list of arguments into a SearchExpr.
// If no operators are found, it creates an implicit OR of all patterns.
func ParseSearchExpression(args []string) (SearchExpr, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no patterns provided")
	}

	// Check if we have any operators
	hasOperators := false
	for _, arg := range args {
		upper := strings.ToUpper(arg)
		if upper == "AND" || upper == "OR" || upper == "XOR" || upper == "NOT" ||
			upper == "&&" || upper == "||" || upper == "^" || upper == "!" ||
			strings.Contains(arg, "(") || strings.Contains(arg, ")") {
			hasOperators = true
			break
		}
	}

	// If no operators, create implicit OR of all patterns (backward compatible)
	if !hasOperators {
		if len(args) == 1 {
			return &PatternExpr{Pattern: args[0]}, nil
		}
		// Create OR chain
		var expr SearchExpr = &PatternExpr{Pattern: args[0]}
		for i := 1; i < len(args); i++ {
			expr = &OrExpr{
				Left:  expr,
				Right: &PatternExpr{Pattern: args[i]},
			}
		}
		return expr, nil
	}

	// Parse with operators
	lexer := newExprLexer(args)
	if err := lexer.tokenize(); err != nil {
		return nil, err
	}

	parser := newExprParser(lexer)
	return parser.Parse()
}
