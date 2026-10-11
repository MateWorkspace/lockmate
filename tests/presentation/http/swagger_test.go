package presentationhttp_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"testing"

	swaggerdocs "github.com/MateWorkspace/lockmate/docs/swagger"
)

func swaggerDocument(t *testing.T) map[string]any {
	t.Helper()
	var document map[string]any
	if err := json.Unmarshal([]byte(swaggerdocs.SwaggerInfo.ReadDoc()), &document); err != nil {
		t.Fatal(err)
	}
	return document
}
func swaggerPath(path string) string {
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if strings.HasPrefix(part, ":") {
			parts[i] = "{" + part[1:] + "}"
		}
	}
	return strings.Join(parts, "/")
}
func TestSwaggerMatchesRoutesAndSecurity(t *testing.T) {
	fixture := newHTTPFixture()
	document := swaggerDocument(t)
	if document["basePath"] != "/api" {
		t.Fatal("wrong base path")
	}
	paths := document["paths"].(map[string]any)
	count := 0
	for _, route := range fixture.router.Routes() {
		if strings.HasPrefix(route.Path, "/api/docs") {
			continue
		}
		path := swaggerPath(strings.TrimPrefix(route.Path, "/api"))
		methods, ok := paths[path].(map[string]any)
		if !ok {
			t.Fatalf("undocumented route %s", route.Path)
		}
		operation, ok := methods[strings.ToLower(route.Method)].(map[string]any)
		if !ok {
			t.Fatalf("undocumented method %s %s", route.Method, route.Path)
		}
		count++
		status := 200
		if route.Method == "POST" && !strings.HasSuffix(path, "/sign-in") && !strings.HasSuffix(path, "/refresh") {
			status = 201
		}
		if route.Method == "PATCH" && !strings.HasSuffix(path, "/secret") || route.Method == "DELETE" {
			status = 204
		}
		responses := operation["responses"].(map[string]any)
		if _, ok := responses[strconv.Itoa(status)]; !ok {
			t.Fatalf("wrong success status in Swagger: %s %s", route.Method, path)
		}
		if path == "/version" {
			if _, exists := operation["security"]; exists {
				t.Fatal("public version requires credentials")
			}
			continue
		}
		security := operation["security"].([]any)
		if len(security) != 1 {
			t.Fatal("security schemes are alternatives")
		}
		requirements := security[0].(map[string]any)
		expected := 3
		if !strings.Contains(path, "/auth/") {
			expected = 4
		}
		if len(requirements) != expected {
			t.Fatalf("wrong security requirements: %v", requirements)
		}
		for _, key := range []string{"ClientSlug", "ClientSecret", "ClientExecId"} {
			if _, ok := requirements[key]; !ok {
				t.Fatalf("missing %s", key)
			}
		}
	}
	docsCount := 0
	for _, methods := range paths {
		docsCount += len(methods.(map[string]any))
	}
	if count != docsCount || count != len(operationCases)+1 {
		t.Fatalf("route/doc coverage: routes=%d docs=%d", count, docsCount)
	}
	raw := swaggerdocs.SwaggerInfo.ReadDoc()
	for _, field := range []string{"password_hash", "secret_key_hash"} {
		if strings.Contains(raw, `"`+field+`"`) {
			t.Fatalf("credential field %s in wire schema", field)
		}
	}
}
func TestUsecaseOperationCoverage(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	domain := filepath.Join(filepath.Dir(file), "../../../internal/domain/usecases")
	covered := map[string]bool{}
	for _, item := range operationCases {
		covered[item.operation] = true
	}
	for _, group := range []string{"auth", "profile", "management"} {
		paths, err := filepath.Glob(filepath.Join(domain, group, "*.go"))
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
			if err != nil {
				t.Fatal(err)
			}
			ast.Inspect(parsed, func(node ast.Node) bool {
				spec, ok := node.(*ast.TypeSpec)
				if !ok {
					return true
				}
				iface, ok := spec.Type.(*ast.InterfaceType)
				if !ok {
					return true
				}
				if group == "auth" && spec.Name.Name != "Session" {
					return false
				}
				for _, method := range iface.Methods.List {
					name := method.Names[0].Name
					if name == "AuthenticateAccessToken" {
						continue
					}
					key := group + "." + spec.Name.Name + "." + name
					if !covered[key] {
						t.Errorf("usecase operation has no HTTP coverage: %s", key)
					}
				}
				return false
			})
		}
	}
}

func TestPresentationDependsOnUsecases(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	directory := filepath.Join(filepath.Dir(file), "../../../internal/presentation/http")
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".go") {
			return nil
		}
		parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
		if err != nil {
			return err
		}
		for _, dependency := range parsed.Imports {
			name := dependency.Path.Value
			if strings.Contains(name, "/internal/infrastructure/") || strings.Contains(name, "/internal/application/") || strings.Contains(name, "/internal/domain/contracts/") {
				t.Errorf("presentation dependency %s in %s", name, path)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
