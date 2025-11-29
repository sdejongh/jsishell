package builtins

import (
	"os"
	"testing"
)

// Helper to create FileInfo for testing
func testFileInfo(name string) *FileInfo {
	return &FileInfo{
		Name:   name,
		Mode:   0644, // Regular file
		IsDir:  false,
		IsLink: false,
	}
}

func testDirInfo(name string) *FileInfo {
	return &FileInfo{
		Name:   name,
		Mode:   os.ModeDir | 0755,
		IsDir:  true,
		IsLink: false,
	}
}

func testExecInfo(name string) *FileInfo {
	return &FileInfo{
		Name:   name,
		Mode:   0755, // Executable
		IsDir:  false,
		IsLink: false,
	}
}

func testLinkInfo(name string, target string) *FileInfo {
	return &FileInfo{
		Name:       name,
		Mode:       os.ModeSymlink | 0777,
		IsDir:      false,
		IsLink:     true,
		LinkTarget: target,
	}
}

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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
		{"*.go", "main.go", false}, // NOT *.go should NOT match .go files
		{"*.go", "main.txt", true}, // NOT *.go should match non-.go files
		{"test_*", "test_foo", false},
		{"test_*", "foo_test", true},
	}

	for _, tt := range tests {
		t.Run("NOT_"+tt.pattern+"_"+tt.filename, func(t *testing.T) {
			expr := &NotExpr{Expr: &PatternExpr{Pattern: tt.pattern}}
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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
			got := expr.Evaluate(testFileInfo(tt.filename))
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

func TestTypeExpr(t *testing.T) {
	tests := []struct {
		name      string
		predicate TypePredicate
		info      *FileInfo
		want      bool
	}{
		// isFile tests
		{
			name:      "isFile - regular file",
			predicate: PredicateIsFile,
			info:      testFileInfo("test.txt"),
			want:      true,
		},
		{
			name:      "isFile - directory",
			predicate: PredicateIsFile,
			info:      testDirInfo("testdir"),
			want:      false,
		},
		{
			name:      "isFile - symlink",
			predicate: PredicateIsFile,
			info:      testLinkInfo("link", "target"),
			want:      false,
		},
		// isDir tests
		{
			name:      "isDir - directory",
			predicate: PredicateIsDir,
			info:      testDirInfo("testdir"),
			want:      true,
		},
		{
			name:      "isDir - regular file",
			predicate: PredicateIsDir,
			info:      testFileInfo("test.txt"),
			want:      false,
		},
		// isLink tests
		{
			name:      "isLink - symlink",
			predicate: PredicateIsLink,
			info:      testLinkInfo("link", "target"),
			want:      true,
		},
		{
			name:      "isLink - regular file",
			predicate: PredicateIsLink,
			info:      testFileInfo("test.txt"),
			want:      false,
		},
		// isSymlink tests (alias for isLink)
		{
			name:      "isSymlink - symlink",
			predicate: PredicateIsSymlink,
			info:      testLinkInfo("link", "target"),
			want:      true,
		},
		// isExec tests
		{
			name:      "isExec - executable file",
			predicate: PredicateIsExec,
			info:      testExecInfo("script.sh"),
			want:      true,
		},
		{
			name:      "isExec - regular file",
			predicate: PredicateIsExec,
			info:      testFileInfo("test.txt"),
			want:      false,
		},
		{
			name:      "isExec - directory (not exec)",
			predicate: PredicateIsExec,
			info:      testDirInfo("testdir"),
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr := &TypeExpr{Predicate: tt.predicate}
			got := expr.Evaluate(tt.info)
			if got != tt.want {
				t.Errorf("TypeExpr{%v}.Evaluate(%v) = %v, want %v", tt.predicate, tt.info.Name, got, tt.want)
			}
		})
	}
}

func TestTypeExprString(t *testing.T) {
	tests := []struct {
		predicate TypePredicate
		want      string
	}{
		{PredicateIsFile, "isFile"},
		{PredicateIsDir, "isDir"},
		{PredicateIsLink, "isLink"},
		{PredicateIsSymlink, "isSymlink"},
		{PredicateIsHardlink, "isHardlink"},
		{PredicateIsExec, "isExec"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			expr := &TypeExpr{Predicate: tt.predicate}
			if got := expr.String(); got != tt.want {
				t.Errorf("TypeExpr{%v}.String() = %q, want %q", tt.predicate, got, tt.want)
			}
		})
	}
}

func TestParseSearchExpression_WithTypePredicates(t *testing.T) {
	tests := []struct {
		name string
		args []string
		info *FileInfo
		want bool
	}{
		{
			name: "isFile alone",
			args: []string{"isFile"},
			info: testFileInfo("test.txt"),
			want: true,
		},
		{
			name: "isFile alone - directory",
			args: []string{"isFile"},
			info: testDirInfo("testdir"),
			want: false,
		},
		{
			name: "isDir alone",
			args: []string{"isDir"},
			info: testDirInfo("testdir"),
			want: true,
		},
		{
			name: "pattern AND isFile - match",
			args: []string{"*.go", "AND", "isFile"},
			info: testFileInfo("main.go"),
			want: true,
		},
		{
			name: "pattern AND isFile - directory matches pattern",
			args: []string{"*.go", "AND", "isFile"},
			info: testDirInfo("vendor.go"),
			want: false,
		},
		{
			name: "pattern AND isDir",
			args: []string{"*config*", "AND", "isDir"},
			info: testDirInfo("config"),
			want: true,
		},
		{
			name: "pattern AND isDir - file",
			args: []string{"*config*", "AND", "isDir"},
			info: testFileInfo("config.yaml"),
			want: false,
		},
		{
			name: "isExec - executable",
			args: []string{"isExec"},
			info: testExecInfo("script.sh"),
			want: true,
		},
		{
			name: "isExec AND pattern",
			args: []string{"isExec", "AND", "*.sh"},
			info: testExecInfo("run.sh"),
			want: true,
		},
		{
			name: "NOT isDir",
			args: []string{"NOT", "isDir"},
			info: testFileInfo("test.txt"),
			want: true,
		},
		{
			name: "NOT isDir - directory",
			args: []string{"NOT", "isDir"},
			info: testDirInfo("testdir"),
			want: false,
		},
		{
			name: "isFile OR isDir",
			args: []string{"isFile", "OR", "isDir"},
			info: testFileInfo("test.txt"),
			want: true,
		},
		{
			name: "isFile OR isDir - directory",
			args: []string{"isFile", "OR", "isDir"},
			info: testDirInfo("testdir"),
			want: true,
		},
		{
			name: "isFile OR isDir - symlink",
			args: []string{"isFile", "OR", "isDir"},
			info: testLinkInfo("link", "target"),
			want: false,
		},
		{
			name: "complex: (*.go OR *.md) AND isFile",
			args: []string{"(", "*.go", "OR", "*.md", ")", "AND", "isFile"},
			info: testFileInfo("main.go"),
			want: true,
		},
		{
			name: "complex: (*.go OR *.md) AND isFile - directory",
			args: []string{"(", "*.go", "OR", "*.md", ")", "AND", "isFile"},
			info: testDirInfo("test.go"),
			want: false,
		},
		{
			name: "case insensitive predicate - ISFILE",
			args: []string{"ISFILE"},
			info: testFileInfo("test.txt"),
			want: true,
		},
		{
			name: "case insensitive predicate - IsDir",
			args: []string{"IsDir"},
			info: testDirInfo("testdir"),
			want: true,
		},
		{
			name: "isLink",
			args: []string{"isLink"},
			info: testLinkInfo("link", "target"),
			want: true,
		},
		{
			name: "isSymlink",
			args: []string{"isSymlink"},
			info: testLinkInfo("link", "target"),
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpression(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpression(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.info)
			if got != tt.want {
				t.Errorf("ParseSearchExpression(%v).Evaluate(%v) = %v, want %v", tt.args, tt.info.Name, got, tt.want)
			}
		})
	}
}

func TestIsTypePredicate(t *testing.T) {
	tests := []struct {
		input string
		want  TypePredicate
		ok    bool
	}{
		{"isFile", PredicateIsFile, true},
		{"ISFILE", PredicateIsFile, true},
		{"IsFile", PredicateIsFile, true},
		{"isDir", PredicateIsDir, true},
		{"ISDIR", PredicateIsDir, true},
		{"isLink", PredicateIsLink, true},
		{"isSymlink", PredicateIsSymlink, true},
		{"isHardlink", PredicateIsHardlink, true},
		{"isExec", PredicateIsExec, true},
		{"notapredicate", 0, false},
		{"*.go", 0, false},
		{"AND", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			pred, ok := isTypePredicate(tt.input)
			if ok != tt.ok {
				t.Errorf("isTypePredicate(%q) ok = %v, want %v", tt.input, ok, tt.ok)
			}
			if ok && pred != tt.want {
				t.Errorf("isTypePredicate(%q) = %v, want %v", tt.input, pred, tt.want)
			}
		})
	}
}

func TestParseSearchExpressionWithQuoting(t *testing.T) {
	tests := []struct {
		name string
		args []ExprArg
		info *FileInfo
		want bool
		desc string
	}{
		{
			name: "quoted AND is treated as pattern",
			args: []ExprArg{
				{Value: "AND", Quoted: true},
			},
			info: testFileInfo("AND"),
			want: true,
			desc: "A file named 'AND' should match when AND is quoted",
		},
		{
			name: "quoted AND does not match other files",
			args: []ExprArg{
				{Value: "AND", Quoted: true},
			},
			info: testFileInfo("test.txt"),
			want: false,
			desc: "A file named 'test.txt' should not match quoted 'AND'",
		},
		{
			name: "quoted isFile is treated as pattern",
			args: []ExprArg{
				{Value: "isFile", Quoted: true},
			},
			info: testFileInfo("isFile"),
			want: true,
			desc: "A file named 'isFile' should match when isFile is quoted",
		},
		{
			name: "quoted isFile does not act as predicate",
			args: []ExprArg{
				{Value: "isFile", Quoted: true},
			},
			info: testFileInfo("test.txt"),
			want: false,
			desc: "Quoted 'isFile' should not act as a type predicate",
		},
		{
			name: "quoted isDir is treated as pattern",
			args: []ExprArg{
				{Value: "isDir", Quoted: true},
			},
			info: testDirInfo("isDir"),
			want: true,
			desc: "A directory named 'isDir' should match when isDir is quoted",
		},
		{
			name: "unquoted isDir acts as predicate",
			args: []ExprArg{
				{Value: "isDir", Quoted: false},
			},
			info: testDirInfo("anydir"),
			want: true,
			desc: "Unquoted 'isDir' should act as type predicate matching any directory",
		},
		{
			name: "mixed quoted and unquoted",
			args: []ExprArg{
				{Value: "*.go", Quoted: true},
				{Value: "AND", Quoted: false},
				{Value: "isFile", Quoted: false},
			},
			info: testFileInfo("main.go"),
			want: true,
			desc: "Pattern *.go AND isFile should match main.go",
		},
		{
			name: "mixed quoted and unquoted - directory",
			args: []ExprArg{
				{Value: "*.go", Quoted: true},
				{Value: "AND", Quoted: false},
				{Value: "isFile", Quoted: false},
			},
			info: testDirInfo("vendor.go"),
			want: false,
			desc: "Pattern *.go AND isFile should not match a directory named vendor.go",
		},
		{
			name: "quoted OR is pattern, not operator",
			args: []ExprArg{
				{Value: "test", Quoted: false},
				{Value: "OR", Quoted: true}, // This is a pattern, not operator
			},
			info: testFileInfo("OR"),
			want: true,
			desc: "Quoted 'OR' should be treated as pattern, creating implicit OR with 'test'",
		},
		{
			name: "quoted NOT is pattern",
			args: []ExprArg{
				{Value: "NOT", Quoted: true},
			},
			info: testFileInfo("NOT"),
			want: true,
			desc: "A file named 'NOT' should match when NOT is quoted",
		},
		{
			name: "all quoted - no operators",
			args: []ExprArg{
				{Value: "AND", Quoted: true},
				{Value: "OR", Quoted: true},
				{Value: "NOT", Quoted: true},
			},
			info: testFileInfo("AND"),
			want: true,
			desc: "All quoted operators should be patterns with implicit OR",
		},
		{
			name: "all quoted - file named OR",
			args: []ExprArg{
				{Value: "AND", Quoted: true},
				{Value: "OR", Quoted: true},
				{Value: "NOT", Quoted: true},
			},
			info: testFileInfo("OR"),
			want: true,
			desc: "File named 'OR' should match implicit OR of quoted patterns",
		},
		{
			name: "quoted XOR is pattern",
			args: []ExprArg{
				{Value: "XOR", Quoted: true},
			},
			info: testFileInfo("XOR"),
			want: true,
			desc: "A file named 'XOR' should match when XOR is quoted",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ParseSearchExpressionWithQuoting(tt.args)
			if err != nil {
				t.Fatalf("ParseSearchExpressionWithQuoting(%v) error = %v", tt.args, err)
			}
			got := expr.Evaluate(tt.info)
			if got != tt.want {
				t.Errorf("%s: got %v, want %v", tt.desc, got, tt.want)
			}
		})
	}
}
