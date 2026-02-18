package tags

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

// parseDeleteFile parses delete.go in the same directory as this test file.
func parseDeleteFile(t *testing.T) *ast.File {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("cannot determine test file path")
	}
	deleteFile := filepath.Join(filepath.Dir(thisFile), "delete.go")

	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, deleteFile, nil, parser.AllErrors)
	if err != nil {
		t.Fatalf("failed to parse delete.go: %v", err)
	}
	return f
}

func TestDeleteDoesNotUseStdin(t *testing.T) {
	f := parseDeleteFile(t)

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
			t.Error("delete.go must not reference os.Stdin — delete should execute without interactive confirmation")
		}
		return true
	})
}

func TestDeleteDoesNotUseBufio(t *testing.T) {
	f := parseDeleteFile(t)

	for _, imp := range f.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		if path == "bufio" {
			t.Error("delete.go must not import bufio — delete should execute without interactive confirmation")
		}
	}
}

func TestDeleteForceFlag(t *testing.T) {
	// The --force flag should not exist or should be a deprecated no-op.
	// Check that init() does not register a "force" flag.
	cmd := deleteCmd
	flag := cmd.Flags().Lookup("force")
	if flag != nil {
		t.Error("delete command should not have a --force flag — delete should always execute without confirmation")
	}
}
