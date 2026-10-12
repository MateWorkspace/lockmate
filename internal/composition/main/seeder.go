package compositionmain

import (
	"slices"

	seederdata "github.com/MateWorkspace/lockmate/database/seeder"
	"github.com/MateWorkspace/lockmate/internal/config"
	domainmodelsseeder "github.com/MateWorkspace/lockmate/internal/domain/models/seeder"
)

// newSeedData adds runtime credentials without changing the embedded definitions.
func newSeedData() domainmodelsseeder.Data {
	data := seederdata.SeedData
	data.Users = slices.Clone(data.Users)
	data.Clients = slices.Clone(data.Clients)

	for index := range data.Users {
		switch data.Users[index].RoleSlug {
		case "super":
			data.Users[index].Password = config.Env.SeederSuperPassword
		case "admin":
			data.Users[index].Password = config.Env.SeederAdminPassword
		case "user":
			data.Users[index].Password = config.Env.SeederUserPassword
		}
	}

	for index := range data.Clients {
		data.Clients[index].SecretKey = config.Env.SeederClientSecretKey
	}

	return data
}
