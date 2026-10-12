package presentationgrpcmapper

import (
	"encoding/json"

	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
)

type jwkFields struct {
	Kty string `json:"kty"`
	Crv string `json:"crv"`
	Alg string `json:"alg"`
	Use string `json:"use"`
	Kid string `json:"kid"`
	X   string `json:"x"`
	Y   string `json:"y"`
}

// Jwks converts raw public JWK documents into typed keys.
func Jwks(values []json.RawMessage) []*lockmatev1.Jwk {
	result := make([]*lockmatev1.Jwk, 0, len(values))
	for _, raw := range values {
		var fields jwkFields
		if json.Unmarshal(raw, &fields) != nil {
			continue
		}

		result = append(result, &lockmatev1.Jwk{
			Kty: fields.Kty,
			Crv: fields.Crv,
			Alg: fields.Alg,
			Use: fields.Use,
			Kid: fields.Kid,
			X:   fields.X,
			Y:   fields.Y,
		})
	}

	return result
}
