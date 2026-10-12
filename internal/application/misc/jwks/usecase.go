package applicationmiscjwks

import (
	"context"
	"encoding/json"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

type usecase struct {
	keys domaincontractsrepository.SigningKey
}

func NewUsecaseImpl(keys domaincontractsrepository.SigningKey) domainusecasesmisc.Jwks {
	return &usecase{keys: keys}
}

func (u *usecase) PublicKeys(ctx context.Context) ([]json.RawMessage, error) {
	values, err := u.keys.ListVerifiable(ctx)
	if err != nil {
		return nil, err
	}

	result := make([]json.RawMessage, 0, len(values))
	for _, value := range values {
		if len(value.PublicJWK) == 0 {
			continue
		}

		result = append(result, value.PublicJWK)
	}

	return result, nil
}
