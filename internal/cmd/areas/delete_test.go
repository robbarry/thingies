package areas

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestDeleteNoStdinRead verifies that the areas delete command does not read
// from os.Stdin (i.e., no interactive confirmation prompt). This test parses
// the source file using go/ast and fails if it finds references to os.Stdin
// or the bufio package.
func TestDeleteNoStdinRead(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "delete.go", nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse delete.go: %v", err)
	}

	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		if ident.Name == "os" && sel.Sel.Name == "Stdin" {
			t.Errorf("delete.go must not reference os.Stdin (found at %s)", fset.Position(n.Pos()))
		}
		return true
	})

	for _, imp := range f.Imports {
		if imp.Path.Value == `"bufio"` {
			t.Errorf("delete.go must not import bufio (found at %s)", fset.Position(imp.Pos()))
		}
	}
}

// TestDeleteNoForceFlag verifies that the --force flag is removed or not
// registered on the delete command, since delete should execute without
// prompting.
func TestDeleteNoForceFlag(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag != nil {
		t.Errorf("delete command should not have a --force flag; confirmation prompts should be removed entirely")
	}
}
