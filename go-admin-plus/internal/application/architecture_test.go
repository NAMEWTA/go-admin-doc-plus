package application

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const moduleImportMarker = "/internal/modules/"

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
					importPath, err := strconv.Unquote(typed.Path.Value)
					if err != nil {
						t.Fatalf("unquote import in %s: %v", filename, err)
					}
					for _, prefix := range banned {
						if strings.HasPrefix(importPath, prefix) {
							t.Errorf("application kernel imports host dependency %q in %s", importPath, filename)
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

func TestProductionModulesDoNotImportOtherModules(t *testing.T) {
	_, source, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve architecture test path")
	}
	root := filepath.Clean(filepath.Join(filepath.Dir(source), "..", ".."))
	violations, err := productionModuleImportViolations(os.DirFS(root))
	if err != nil {
		t.Fatalf("inspect production module imports: %v", err)
	}
	if len(violations) != 0 {
		t.Fatalf("production modules must collaborate through consumer ports injected by internal/app; cross-module imports:\n%s", strings.Join(violations, "\n"))
	}
}

func TestProductionModuleImportFixtureRejectsCrossModuleDependency(t *testing.T) {
	root := t.TempDir()
	writeFixture := func(name, content string) {
		t.Helper()
		filename := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(filename), 0o755); err != nil {
			t.Fatalf("create fixture directory: %v", err)
		}
		if err := os.WriteFile(filename, []byte(content), 0o600); err != nil {
			t.Fatalf("write fixture: %v", err)
		}
	}
	writeFixture("internal/modules/orders/service.go", `package orders
import (
	"example.test/product/internal/modules/iam/session"
	"example.test/product/internal/modules/orders/transport"
)
`)
	writeFixture("internal/modules/orders/service_test.go", `package orders
import "example.test/product/internal/modules/iam/authorization"
`)

	violations, err := productionModuleImportViolations(os.DirFS(root))
	if err != nil {
		t.Fatalf("inspect fixture: %v", err)
	}
	want := []string{"internal/modules/orders/service.go: orders imports iam"}
	if fmt.Sprint(violations) != fmt.Sprint(want) {
		t.Fatalf("violations = %v, want %v", violations, want)
	}
}

func productionModuleImportViolations(root fs.FS) ([]string, error) {
	var violations []string
	seen := map[string]struct{}{}
	err := fs.WalkDir(root, "internal/modules", func(filename string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || path.Ext(filename) != ".go" || strings.HasSuffix(filename, "_test.go") {
			return nil
		}
		parts := strings.Split(filename, "/")
		if len(parts) < 4 {
			return nil
		}
		sourceModule := parts[2]
		content, err := fs.ReadFile(root, filename)
		if err != nil {
			return fmt.Errorf("read %s: %w", filename, err)
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), filename, content, parser.ImportsOnly)
		if err != nil {
			return fmt.Errorf("parse %s: %w", filename, err)
		}
		for _, specification := range parsed.Imports {
			target, err := importModule(specification)
			if err != nil {
				return fmt.Errorf("parse import in %s: %w", filename, err)
			}
			if target == "" || target == sourceModule {
				continue
			}
			violation := fmt.Sprintf("%s: %s imports %s", filename, sourceModule, target)
			if _, exists := seen[violation]; !exists {
				seen[violation] = struct{}{}
				violations = append(violations, violation)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Strings(violations)
	return violations, nil
}

func importModule(specification *ast.ImportSpec) (string, error) {
	importPath, err := strconv.Unquote(specification.Path.Value)
	if err != nil {
		return "", err
	}
	marker := strings.Index(importPath, moduleImportMarker)
	if marker < 0 {
		return "", nil
	}
	remainder := strings.TrimPrefix(importPath[marker+len(moduleImportMarker):], "/")
	if remainder == "" {
		return "", nil
	}
	return strings.SplitN(remainder, "/", 2)[0], nil
}
