package tasks

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"testing"
)

// TestDeleteNoStdinRead verifies that runDelete does not read from os.Stdin.
// The delete command should execute without interactive confirmation.
// This test parses the source file and checks for references to os.Stdin and bufio,
// which would indicate an interactive prompt.
func TestDeleteNoStdinRead(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}

	deleteFile := filepath.Join(filepath.Dir(thisFile), "delete.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, deleteFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse delete.go: %v", err)
	}

	var stdinRefs []string
	ast.Inspect(f, func(n ast.Node) bool {
		sel, ok := n.(*ast.SelectorExpr)
		if !ok {
			return true
		}
		ident, ok := sel.X.(*ast.Ident)
		if !ok {
			return true
		}
		// Flag os.Stdin references
		if ident.Name == "os" && sel.Sel.Name == "Stdin" {
			pos := fset.Position(n.Pos())
			stdinRefs = append(stdinRefs, pos.String())
		}
		return true
	})

	if len(stdinRefs) > 0 {
		t.Errorf("delete.go must not reference os.Stdin (found at %v); delete should not prompt for confirmation", stdinRefs)
	}
}

// TestDeleteNoBufioImport verifies that delete.go does not import "bufio",
// which would indicate interactive stdin reading for confirmation prompts.
func TestDeleteNoBufioImport(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("could not determine test file path")
	}

	deleteFile := filepath.Join(filepath.Dir(thisFile), "delete.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, deleteFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse delete.go: %v", err)
	}

	for _, imp := range f.Imports {
		if imp.Path.Value == `"bufio"` {
			t.Errorf("delete.go must not import \"bufio\"; delete should not prompt for confirmation")
		}
	}
}

// TestDeleteNoForceFlag verifies that the --force flag has been removed.
// Since delete no longer prompts, --force is unnecessary.
func TestDeleteNoForceFlag(t *testing.T) {
	flag := deleteCmd.Flags().Lookup("force")
	if flag != nil {
		t.Error("delete command should not have a --force flag; confirmation prompts have been removed")
	}
}
