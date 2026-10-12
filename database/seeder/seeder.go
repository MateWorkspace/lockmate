package seeder

import (
	"bytes"
	"embed"
	"encoding/json"
	"io"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelsseeder "github.com/MateWorkspace/lockmate/internal/domain/models/seeder"
)

//go:embed space.json permission.json role.json user.json client.json
var files embed.FS

var SeedData domainmodelsseeder.Data

// Load publishes seed definitions only after every embedded document is decoded.
func Load() error {
	var data domainmodelsseeder.Data
	if err := decode("space.json", &data.Space); err != nil {
		return err
	}
	if err := decode("permission.json", &data.Permissions); err != nil {
		return err
	}
	if err := decode("role.json", &data.Roles); err != nil {
		return err
	}
	if err := decode("user.json", &data.Users); err != nil {
		return err
	}
	if err := decode("client.json", &data.Clients); err != nil {
		return err
	}

	SeedData = data
	return nil
}

func decode(name string, target any) error {
	raw, err := files.ReadFile(name)
	if err != nil {
		return domainmodelserror.NewError("failed to read embedded seed definition "+name, domainmodelserror.Failure, err)
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return domainmodelserror.NewError("invalid embedded seed definition "+name, domainmodelserror.BadArgs, nil)
	}
	if err := decoder.Decode(new(any)); err != io.EOF {
		return domainmodelserror.NewError("seed definition must contain one JSON value: "+name, domainmodelserror.BadArgs, nil)
	}
	return nil
}
