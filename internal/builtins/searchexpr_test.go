package builtins

import (
	"testing"
)

func TestPatternExpr(t *testing.T) {
	tests := []struct {
		pattern  string
		filename string
		want     bool
	}{
		{"*.go", "main.go", true},
		{"*.go", "main.txt", false},
		{"test_*", "test_foo.go", true},
		{"test_*", "foo_test.go", false},
		{"*.?", "file.a", true},
		{"*.?", "file.ab", false},
		{"[abc]*", "afile", true},
		{"[abc]*", "dfile", false},
	}

	for _, tt := range tests {
		t.Run(tt.pattern+"_"+tt.filename, func(t *testing.T) {
			expr := &PatternExpr{Pattern: tt.pattern}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("PatternExpr{%q}.Evaluate(%q) = %v, want %v", tt.pattern, tt.filename, got, tt.want)
			}
		})
	}
}

func TestNotExpr(t *testing.T) {
	tests := []struct {
		pattern  string
		filename string
		want     bool
	}{
		{"*.go", "main.go", false},  // NOT *.go should NOT match .go files
		{"*.go", "main.txt", true},  // NOT *.go should match non-.go files
		{"test_*", "test_foo", false},
		{"test_*", "foo_test", true},
	}

	for _, tt := range tests {
		t.Run("NOT_"+tt.pattern+"_"+tt.filename, func(t *testing.T) {
			expr := &NotExpr{Expr: &PatternExpr{Pattern: tt.pattern}}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("NOT(%q).Evaluate(%q) = %v, want %v", tt.pattern, tt.filename, got, tt.want)
			}
		})
	}
}

func TestAndExpr(t *testing.T) {
	tests := []struct {
		name     string
		left     SearchExpr
		right    SearchExpr
		filename string
		want     bool
	}{
		{
			name:     "both match",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "left only",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     false,
		},
		{
			name:     "right only",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     false,
		},
		{
			name:     "neither match",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &AndExpr{Left: tt.left, Right: tt.right}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("AndExpr.Evaluate(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestOrExpr(t *testing.T) {
	tests := []struct {
		name     string
		left     SearchExpr
		right    SearchExpr
		filename string
		want     bool
	}{
		{
			name:     "both match",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "left only",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "right only",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "neither match",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &OrExpr{Left: tt.left, Right: tt.right}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("OrExpr.Evaluate(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestXorExpr(t *testing.T) {
	tests := []struct {
		name     string
		left     SearchExpr
		right    SearchExpr
		filename string
		want     bool
	}{
		{
			name:     "both match - XOR is false",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     false, // XOR: true XOR true = false
		},
		{
			name:     "left only - XOR is true",
			left:     &PatternExpr{Pattern: "*.go"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     true, // XOR: true XOR false = true
		},
		{
			name:     "right only - XOR is true",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "main*"},
			filename: "main.go",
			want:     true, // XOR: false XOR true = true
		},
		{
			name:     "neither match - XOR is false",
			left:     &PatternExpr{Pattern: "*.txt"},
			right:    &PatternExpr{Pattern: "test*"},
			filename: "main.go",
			want:     false, // XOR: false XOR false = false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &XorExpr{Left: tt.left, Right: tt.right}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("XorExpr.Evaluate(%q) = %v, want %v", tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_Simple(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		filename string
		want     bool
	}{
		{
			name:     "single pattern match",
			args:     []string{"*.go"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "single pattern no match",
			args:     []string{"*.go"},
			filename: "main.txt",
			want:     false,
		},
		{
			name:     "implicit OR - first matches",
			args:     []string{"*.go", "*.txt"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "implicit OR - second matches",
			args:     []string{"*.go", "*.txt"},
			filename: "data.txt",
			want:     true,
		},
		{
			name:     "implicit OR - neither matches",
			args:     []string{"*.go", "*.txt"},
			filename: "image.png",
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpression(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpression(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("ParseSearchExpression(%v).Evaluate(%q) = %v, want %v", tt.args, tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_WithOperators(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		filename string
		want     bool
	}{
		{
			name:     "explicit OR",
			args:     []string{"*.go", "OR", "*.txt"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "explicit AND - both match",
			args:     []string{"*.go", "AND", "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "explicit AND - one matches",
			args:     []string{"*.go", "AND", "test*"},
			filename: "main.go",
			want:     false,
		},
		{
			name:     "NOT pattern",
			args:     []string{"NOT", "*.go"},
			filename: "main.txt",
			want:     true,
		},
		{
			name:     "NOT pattern - negated",
			args:     []string{"NOT", "*.go"},
			filename: "main.go",
			want:     false,
		},
		{
			name:     "pattern AND NOT pattern",
			args:     []string{"*.go", "AND", "NOT", "*_test.go"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "pattern AND NOT pattern - excluded",
			args:     []string{"*.go", "AND", "NOT", "*_test.go"},
			filename: "main_test.go",
			want:     false,
		},
		{
			name:     "XOR - one matches",
			args:     []string{"*.go", "XOR", "*.txt"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "lowercase operators",
			args:     []string{"*.go", "and", "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "symbol operators &&",
			args:     []string{"*.go", "&&", "main*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "symbol operators ||",
			args:     []string{"*.go", "||", "*.txt"},
			filename: "main.txt",
			want:     true,
		},
		{
			name:     "symbol operator !",
			args:     []string{"!", "*.go"},
			filename: "main.txt",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpression(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpression(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("ParseSearchExpression(%v).Evaluate(%q) = %v, want %v", tt.args, tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_WithParentheses(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		filename string
		want     bool
	}{
		{
			name:     "grouped OR with AND",
			args:     []string{"(", "*.go", "OR", "*.md", ")", "AND", "NOT", "*_test*"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "grouped OR with AND - excluded",
			args:     []string{"(", "*.go", "OR", "*.md", ")", "AND", "NOT", "*_test*"},
			filename: "main_test.go",
			want:     false,
		},
		{
			name:     "grouped OR with AND - md file",
			args:     []string{"(", "*.go", "OR", "*.md", ")", "AND", "NOT", "*_test*"},
			filename: "README.md",
			want:     true,
		},
		{
			name:     "nested groups",
			args:     []string{"(", "(", "*.go", ")", ")"},
			filename: "main.go",
			want:     true,
		},
		{
			name:     "parentheses embedded in arg",
			args:     []string{"(*.go", "OR", "*.md)"},
			filename: "main.go",
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpression(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpression(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("ParseSearchExpression(%v).Evaluate(%q) = %v, want %v", tt.args, tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_Precedence(t *testing.T) {
	// Test operator precedence: NOT > AND > XOR > OR
	tests := []struct {
		name     string
		args     []string
		filename string
		want     bool
		desc     string
	}{
		{
			name:     "AND before OR",
			args:     []string{"*.txt", "OR", "*.go", "AND", "main*"},
			filename: "data.txt",
			want:     true,
			desc:     "*.txt OR (*.go AND main*) - txt matches first part",
		},
		{
			name:     "AND before OR - go file",
			args:     []string{"*.txt", "OR", "*.go", "AND", "main*"},
			filename: "main.go",
			want:     true,
			desc:     "*.txt OR (*.go AND main*) - go file matches second part",
		},
		{
			name:     "AND before OR - test.go doesn't match AND",
			args:     []string{"*.txt", "OR", "*.go", "AND", "main*"},
			filename: "test.go",
			want:     false,
			desc:     "*.txt OR (*.go AND main*) - test.go doesn't match main*",
		},
		{
			name:     "NOT before AND",
			args:     []string{"NOT", "*.txt", "AND", "*.go"},
			filename: "main.go",
			want:     true,
			desc:     "(NOT *.txt) AND *.go",
		},
		{
			name:     "XOR before OR",
			args:     []string{"*.a", "OR", "*.b", "XOR", "*.c"},
			filename: "file.a",
			want:     true,
			desc:     "*.a OR (*.b XOR *.c)",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpression(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpression(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.filename)
			if got != tt.want {
				t.Errorf("%s: ParseSearchExpression(%v).Evaluate(%q) = %v, want %v", tt.desc, tt.args, tt.filename, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_Errors(t *testing.T) {
	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{
			name:    "empty args",
			args:    []string{},
			wantErr: true,
		},
		{
			name:    "unclosed paren",
			args:    []string{"(", "*.go"},
			wantErr: true,
		},
		{
			name:    "extra close paren",
			args:    []string{"*.go", ")"},
			wantErr: true,
		},
		{
			name:    "dangling AND",
			args:    []string{"*.go", "AND"},
			wantErr: true,
		},
		{
			name:    "dangling NOT",
			args:    []string{"NOT"},
			wantErr: true,
		},
		{
			name:    "operator without operand",
			args:    []string{"AND", "*.go"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseSearchExpression(tt.args)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseSearchExpression(%v) error = %v, wantErr %v", tt.args, err, tt.wantErr)
			}
		})
	}
}

func TestExprString(t *testing.T) {
	// Test that String() methods work correctly
	pattern := &PatternExpr{Pattern: "*.go"}
	if s := pattern.String(); s != `"*.go"` {
		t.Errorf("PatternExpr.String() = %q, want %q", s, `"*.go"`)
	}

	not := &NotExpr{Expr: pattern}
	if s := not.String(); s != `NOT("*.go")` {
		t.Errorf("NotExpr.String() = %q, want %q", s, `NOT("*.go")`)
	}

	and := &AndExpr{Left: pattern, Right: &PatternExpr{Pattern: "main*"}}
	expected := `("*.go" AND "main*")`
	if s := and.String(); s != expected {
		t.Errorf("AndExpr.String() = %q, want %q", s, expected)
	}
}
