package builtins

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode"
)

// FileInfo contains information about a file for expression evaluation.
type FileInfo struct {
	Name       string      // Base name of the file
	Mode       os.FileMode // File mode and permissions
	IsDir      bool        // True if directory
	IsLink     bool        // True if symbolic link
	LinkTarget string      // Target of symbolic link (if applicable)
}

// SearchExpr represents a search expression that can be evaluated against a file.
type SearchExpr interface {
	// Evaluate returns true if the file matches the expression.
	Evaluate(info *FileInfo) bool
	// String returns a string representation of the expression.
	String() string
}

// PatternExpr represents a glob pattern.
type PatternExpr struct {
	Pattern string
}

func (p *PatternExpr) Evaluate(info *FileInfo) bool {
	matched, err := filepath.Match(p.Pattern, info.Name)
	if err != nil {
		return false
	}
	return matched
}

func (p *PatternExpr) String() string {
	return fmt.Sprintf("%q", p.Pattern)
}

// TypePredicate represents the type of file type predicate.
type TypePredicate int

const (
	PredicateIsFile TypePredicate = iota
	PredicateIsDir
	PredicateIsLink
	PredicateIsSymlink
	PredicateIsHardlink
	PredicateIsExec
)

// TypeExpr represents a file type predicate (isFile, isDir, etc.).
type TypeExpr struct {
	Predicate TypePredicate
}

func (t *TypeExpr) Evaluate(info *FileInfo) bool {
	switch t.Predicate {
	case PredicateIsFile:
		return !info.IsDir && !info.IsLink
	case PredicateIsDir:
		return info.IsDir
	case PredicateIsLink:
		return info.IsLink
	case PredicateIsSymlink:
		return info.IsLink // Symbolic link
	case PredicateIsHardlink:
		// Hard links are regular files with link count > 1
		// We can't easily detect this without additional syscalls
		// For now, treat as regular file that is not a symlink
		return !info.IsDir && !info.IsLink
	case PredicateIsExec:
		return info.Mode&0111 != 0 && !info.IsDir
	default:
		return false
	}
}

func (t *TypeExpr) String() string {
	switch t.Predicate {
	case PredicateIsFile:
		return "isFile"
	case PredicateIsDir:
		return "isDir"
	case PredicateIsLink:
		return "isLink"
	case PredicateIsSymlink:
		return "isSymlink"
	case PredicateIsHardlink:
		return "isHardlink"
	case PredicateIsExec:
		return "isExec"
	default:
		return "unknown"
	}
}

// NotExpr represents a NOT expression.
type NotExpr struct {
	Expr SearchExpr
}

func (n *NotExpr) Evaluate(info *FileInfo) bool {
	return !n.Expr.Evaluate(info)
}

func (n *NotExpr) String() string {
	return fmt.Sprintf("NOT(%s)", n.Expr.String())
}

// AndExpr represents an AND expression.
type AndExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (a *AndExpr) Evaluate(info *FileInfo) bool {
	return a.Left.Evaluate(info) && a.Right.Evaluate(info)
}

func (a *AndExpr) String() string {
	return fmt.Sprintf("(%s AND %s)", a.Left.String(), a.Right.String())
}

// OrExpr represents an OR expression.
type OrExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (o *OrExpr) Evaluate(info *FileInfo) bool {
	return o.Left.Evaluate(info) || o.Right.Evaluate(info)
}

func (o *OrExpr) String() string {
	return fmt.Sprintf("(%s OR %s)", o.Left.String(), o.Right.String())
}

// XorExpr represents an XOR expression.
type XorExpr struct {
	Left  SearchExpr
	Right SearchExpr
}

func (x *XorExpr) Evaluate(info *FileInfo) bool {
	left := x.Left.Evaluate(info)
	right := x.Right.Evaluate(info)
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
	tokTypePredicate
	tokEOF
)

type exprToken struct {
	typ       exprTokenType
	value     string
	predicate TypePredicate // For tokTypePredicate
}

// isTypePredicate checks if a string is a type predicate and returns the predicate type.
func isTypePredicate(s string) (TypePredicate, bool) {
	lower := strings.ToLower(s)
	switch lower {
	case "isfile":
		return PredicateIsFile, true
	case "isdir":
		return PredicateIsDir, true
	case "islink":
		return PredicateIsLink, true
	case "issymlink":
		return PredicateIsSymlink, true
	case "ishardlink":
		return PredicateIsHardlink, true
	case "isexec":
		return PredicateIsExec, true
	default:
		return 0, false
	}
}

// exprLexer tokenizes the search expression arguments.
type exprLexer struct {
	args    []ExprArg
	pos     int
	tokens  []exprToken
	current int
}

func newExprLexer(args []string) *exprLexer {
	// Convert to ExprArg without quoting info for backward compatibility
	exprArgs := make([]ExprArg, len(args))
	for i, arg := range args {
		exprArgs[i] = ExprArg{Value: arg, Quoted: false}
	}
	return newExprLexerWithQuoting(exprArgs)
}

func newExprLexerWithQuoting(args []ExprArg) *exprLexer {
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

		// If the argument is quoted, it's always a pattern (never an operator or predicate)
		if arg.Quoted {
			l.tokens = append(l.tokens, exprToken{typ: tokPattern, value: arg.Value})
			continue
		}

		// Handle operators (case-insensitive) - only for non-quoted arguments
		upper := strings.ToUpper(arg.Value)
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
			// Check for type predicates
			if pred, ok := isTypePredicate(arg.Value); ok {
				l.tokens = append(l.tokens, exprToken{typ: tokTypePredicate, value: arg.Value, predicate: pred})
			} else {
				// Check for parentheses embedded in the argument
				if err := l.tokenizeWithParens(arg.Value); err != nil {
					return err
				}
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
				l.addPatternOrOperator(pattern)
			}
		}
	}
	return nil
}

// addPatternOrOperator adds a token for a pattern, operator, or type predicate.
func (l *exprLexer) addPatternOrOperator(pattern string) {
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
		// Check for type predicates
		if pred, ok := isTypePredicate(pattern); ok {
			l.tokens = append(l.tokens, exprToken{typ: tokTypePredicate, value: pattern, predicate: pred})
		} else {
			l.tokens = append(l.tokens, exprToken{typ: tokPattern, value: pattern})
		}
	}
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
//	primary  = pattern | typePredicate | "(" expr ")"
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

	case tokTypePredicate:
		p.lexer.next()
		return &TypeExpr{Predicate: tok.predicate}, nil

	case tokEOF:
		return nil, fmt.Errorf("unexpected end of expression")

	default:
		return nil, fmt.Errorf("unexpected token: %s", tok.value)
	}
}

// hasOperatorsOrPredicates checks if arguments contain operators or type predicates.
func hasOperatorsOrPredicates(args []string) bool {
	for _, arg := range args {
		upper := strings.ToUpper(arg)
		if upper == "AND" || upper == "OR" || upper == "XOR" || upper == "NOT" ||
			upper == "&&" || upper == "||" || upper == "^" || upper == "!" ||
			strings.Contains(arg, "(") || strings.Contains(arg, ")") {
			return true
		}
		// Check for type predicates
		if _, ok := isTypePredicate(arg); ok {
			return true
		}
	}
	return false
}

// hasOperatorsOrPredicatesWithQuoting checks if non-quoted arguments contain operators or type predicates.
// Quoted arguments are never considered as operators or predicates.
func hasOperatorsOrPredicatesWithQuoting(args []ExprArg) bool {
	for _, arg := range args {
		// Skip quoted arguments - they are always patterns
		if arg.Quoted {
			continue
		}
		upper := strings.ToUpper(arg.Value)
		if upper == "AND" || upper == "OR" || upper == "XOR" || upper == "NOT" ||
			upper == "&&" || upper == "||" || upper == "^" || upper == "!" ||
			strings.Contains(arg.Value, "(") || strings.Contains(arg.Value, ")") {
			return true
		}
		// Check for type predicates
		if _, ok := isTypePredicate(arg.Value); ok {
			return true
		}
	}
	return false
}

// ExprArg represents an argument with quoting information for expression parsing.
type ExprArg struct {
	Value  string
	Quoted bool
}

// ParseSearchExpression parses a list of arguments into a SearchExpr.
// If no operators are found, it creates an implicit OR of all patterns.
func ParseSearchExpression(args []string) (SearchExpr, error) {
	// Convert to ExprArg without quoting info for backward compatibility
	exprArgs := make([]ExprArg, len(args))
	for i, arg := range args {
		exprArgs[i] = ExprArg{Value: arg, Quoted: false}
	}
	return ParseSearchExpressionWithQuoting(exprArgs)
}

// ParseSearchExpressionWithQuoting parses arguments with quoting information.
// Quoted arguments are always treated as patterns, never as operators or predicates.
func ParseSearchExpressionWithQuoting(args []ExprArg) (SearchExpr, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("no patterns provided")
	}

	// Check if we have any operators or predicates (only in non-quoted args)
	if !hasOperatorsOrPredicatesWithQuoting(args) {
		// If no operators, create implicit OR of all patterns (backward compatible)
		if len(args) == 1 {
			return &PatternExpr{Pattern: args[0].Value}, nil
		}
		// Create OR chain
		var expr SearchExpr = &PatternExpr{Pattern: args[0].Value}
		for i := 1; i < len(args); i++ {
			expr = &OrExpr{
				Left:  expr,
				Right: &PatternExpr{Pattern: args[i].Value},
			}
		}
		return expr, nil
	}

	// Parse with operators
	lexer := newExprLexerWithQuoting(args)
	if err := lexer.tokenize(); err != nil {
		return nil, err
	}

	parser := newExprParser(lexer)
	return parser.Parse()
}
