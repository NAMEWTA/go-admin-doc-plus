package application

import (
	"go/ast"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
	"testing"
)

func TestKernelDoesNotOwnHostDependencies(t *testing.T) {
	packages, err := parser.ParseDir(token.NewFileSet(), ".", nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse application package: %v", err)
	}
	banned := []string{"os/signal", "github.com/wailsapp/"}
	for _, parsedPackage := range packages {
		for filename, file := range parsedPackage.Files {
			ast.Inspect(file, func(node ast.Node) bool {
				switch typed := node.(type) {
				case *ast.ImportSpec:
					path, err := strconv.Unquote(typed.Path.Value)
					if err != nil {
						t.Fatalf("unquote import in %s: %v", filename, err)
					}
					for _, prefix := range banned {
						if strings.HasPrefix(path, prefix) {
							t.Errorf("application kernel imports host dependency %q in %s", path, filename)
						}
					}
				case *ast.CallExpr:
					selector, ok := typed.Fun.(*ast.SelectorExpr)
					if ok && strings.HasPrefix(selector.Sel.Name, "Listen") {
						t.Errorf("application kernel owns listener call %s in %s", selector.Sel.Name, filename)
					}
				}
				return true
			})
		}
	}
}
