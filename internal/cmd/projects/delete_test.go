package projects

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"
)

// TestDeleteNoStdinRead ensures runDelete does not read from os.Stdin.
// The delete command should execute without interactive confirmation.
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
			t.Errorf("delete.go must not reference os.Stdin — delete should not prompt for confirmation")
		}
		return true
	})
}

// TestDeleteNoBufioImport ensures delete.go does not import "bufio",
// since there is no need for interactive input.
func TestDeleteNoBufioImport(t *testing.T) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "delete.go", nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse delete.go: %v", err)
	}

	for _, imp := range f.Imports {
		if imp.Path.Value == `"bufio"` {
			t.Errorf("delete.go must not import bufio — delete should not prompt for confirmation")
		}
	}
}

// TestDeleteNoForceFlag ensures the --force flag is not registered,
// since delete should execute unconditionally.
func TestDeleteNoForceFlag(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag != nil {
		t.Errorf("delete command should not have a --force flag — delete should execute without confirmation")
	}
}
