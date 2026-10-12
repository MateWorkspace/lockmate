package seeder_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	seederdata "github.com/MateWorkspace/lockmate/database/seeder"
	applicationseeder "github.com/MateWorkspace/lockmate/internal/application/seeder"
	domainmodelsseeder "github.com/MateWorkspace/lockmate/internal/domain/models/seeder"
	infrastructureutilitykeygen "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/keygen"
	infrastructureutilityvalidator "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/validator"
	testshared "github.com/MateWorkspace/lockmate/tests/application/shared"
)

func definitions(t *testing.T) domainmodelsseeder.Data {
	t.Helper()
	if err := seederdata.Load(); err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(seederdata.SeedData)
	if err != nil {
		t.Fatal(err)
	}
	var data domainmodelsseeder.Data
	if err := json.Unmarshal(raw, &data); err != nil {
		t.Fatal(err)
	}
	for index := range data.Users {
		data.Users[index].Password = data.Users[index].RoleSlug + "-seed-password"
	}
	for index := range data.Clients {
		data.Clients[index].SecretKey = "lockmate-" + strings.Repeat("a", 64)
	}
	return data
}

func TestInvalidDefinitionsFailBeforePersistence(t *testing.T) {
	for _, mode := range []string{"space", "duplicate permission", "duplicate role", "unknown grant", "duplicate grant", "missing default", "two defaults", "unknown user role", "duplicate username", "password", "client secret", "duplicate client", "duplicate cidr", "cidr", "redirect"} {
		t.Run(mode, func(t *testing.T) {
			data := definitions(t)
			switch mode {
			case "space":
				data.Space.Slug = "Invalid"
			case "duplicate permission":
				data.Permissions = append(data.Permissions, data.Permissions[0])
			case "duplicate role":
				data.Roles = append(data.Roles, data.Roles[0])
			case "unknown grant":
				data.Roles[0].Permissions = append(data.Roles[0].Permissions, "missing:get")
			case "duplicate grant":
				data.Roles[0].Permissions = append(data.Roles[0].Permissions, data.Roles[0].Permissions[0])
			case "missing default":
				data.Roles[2].IsDefault = false
			case "two defaults":
				data.Roles[0].IsDefault = true
			case "unknown user role":
				data.Users[0].RoleSlug = "missing"
			case "duplicate username":
				data.Users = append(data.Users, data.Users[0])
			case "password":
				data.Users[0].Password = "x"
			case "client secret":
				data.Clients[0].SecretKey = "PRIVATE_SECRET"
			case "duplicate client":
				data.Clients = append(data.Clients, data.Clients[0])
			case "duplicate cidr":
				data.Clients[0].IpAllowlist = append(data.Clients[0].IpAllowlist, data.Clients[0].IpAllowlist[0])
			case "cidr":
				data.Clients[0].IpAllowlist = []string{"http://evil.example.com"}
			case "redirect":
				uri := "https://app.example.com/path?next=evil"
				data.Clients[0].AuthRedirectUri = &uri
			}
			logger := &testshared.Logger{}
			usecase := applicationseeder.NewUsecaseImpl(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, infrastructureutilitykeygen.NewGeneratorImpl(), infrastructureutilityvalidator.NewValidatorImpl(true), nil, logger, &data)
			err := usecase.Run(context.Background())
			if err == nil || len(logger.Entries) != 1 || logger.Entries[0].Level != "error" || strings.Contains(err.Error(), "PRIVATE_SECRET") {
				t.Fatalf("invalid input accepted or exposed: %v", err)
			}
		})
	}
}
