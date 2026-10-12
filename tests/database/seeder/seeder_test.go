package seeder_test

import (
	"encoding/json"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strconv"
	"strings"
	"testing"

	seederdata "github.com/MateWorkspace/lockmate/database/seeder"
)

func TestEmbeddedSeedCatalog(t *testing.T) {
	if err := seederdata.Load(); err != nil {
		t.Fatal(err)
	}
	data := seederdata.SeedData
	if data.Space.Slug != "lockmate" || len(data.Roles) != 3 || len(data.Users) != 3 || len(data.Clients) != 1 {
		t.Fatal("incomplete baseline")
	}
	_, self, _, _ := runtime.Caller(0)
	source, err := parser.ParseFile(token.NewFileSet(), filepath.Join(filepath.Dir(self), "../../../internal/application/permissions/permissions.go"), nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	expected := map[string]bool{}
	ast.Inspect(source, func(node ast.Node) bool {
		spec, ok := node.(*ast.ValueSpec)
		if !ok {
			return true
		}
		for _, value := range spec.Values {
			if literal, ok := value.(*ast.BasicLit); ok && literal.Kind == token.STRING {
				slug, err := strconv.Unquote(literal.Value)
				if err != nil {
					t.Fatal(err)
				}
				expected[slug] = true
			}
		}
		return false
	})
	actual := map[string]bool{}
	for _, permission := range data.Permissions {
		if actual[permission.Slug] {
			t.Fatal("duplicate permission")
		}
		actual[permission.Slug] = true
	}
	if len(actual) != 46 || len(actual) != len(expected) {
		t.Fatal("seed and application permission counts differ")
	}
	for slug := range expected {
		if !actual[slug] {
			t.Fatalf("missing permission %s", slug)
		}
	}
	for _, role := range data.Roles {
		grants := map[string]bool{}
		for _, slug := range role.Permissions {
			if !actual[slug] || grants[slug] {
				t.Fatal("invalid role permission")
			}
			grants[slug] = true
		}
		switch role.Slug {
		case "super":
			if len(grants) != 46 || role.IsDefault {
				t.Fatal("super policy")
			}
		case "user":
			if len(grants) != 4 || !role.IsDefault || !grants["profile_security:get"] {
				t.Fatal("default policy")
			}
		case "admin":
			if len(grants) != 33 || grants["role_permission:add"] || grants["space_member_role:add"] || !grants["client_ip_allowlist:set"] || !grants["user_password:set"] {
				t.Fatal("admin policy")
			}
		default:
			t.Fatal("unexpected seeded role")
		}
	}
}

func TestSeedDefinitionsContainNoCredentials(t *testing.T) {
	if err := seederdata.Load(); err != nil {
		t.Fatal(err)
	}
	data := seederdata.SeedData
	data.Users = slices.Clone(data.Users)
	data.Clients = slices.Clone(data.Clients)
	for index := range data.Users {
		if data.Users[index].Password != "" {
			t.Fatal("embedded user credential")
		}
		data.Users[index].Password = "private-test-password"
	}
	for index := range data.Clients {
		if data.Clients[index].SecretKey != "" {
			t.Fatal("embedded client credential")
		}
		data.Clients[index].SecretKey = "private-test-client-secret"
	}
	raw, err := json.Marshal(data)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(raw), "private-test-") {
		t.Fatal("runtime seed credentials serialized")
	}

	_, self, _, _ := runtime.Caller(0)
	for name, field := range map[string]string{"user.json": "password", "client.json": "secret_key"} {
		raw, err := os.ReadFile(filepath.Join(filepath.Dir(self), "../../../database/seeder", name))
		if err != nil {
			t.Fatal(err)
		}
		var records []map[string]any
		if err := json.Unmarshal(raw, &records); err != nil {
			t.Fatal(err)
		}
		for _, record := range records {
			if _, exists := record[field]; exists {
				t.Fatal("credential field retained in seed JSON")
			}
		}
	}
}
