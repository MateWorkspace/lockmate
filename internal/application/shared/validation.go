package applicationshared

import (
	"encoding/json"
	"math"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

func ValidateIds(ids ...int64) error {
	for _, id := range ids {
		if id <= 0 {
			return domainmodelserror.NewError("invalid identifier", domainmodelserror.BadArgs, nil)
		}
	}

	return nil
}

func ValidateOptionalId(id *int64) error {
	if id != nil {
		return ValidateIds(*id)
	}

	return nil
}

func ValidatePagination(page, limit int64) error {
	if page > 1 && limit > 0 && page-1 > math.MaxInt64/limit {
		return domainmodelserror.NewError("pagination overflow", domainmodelserror.BadArgs, nil)
	}

	return nil
}

func ValidateJSON(value *json.RawMessage) error {
	if value != nil && !json.Valid(*value) {
		return domainmodelserror.NewError("invalid preferences", domainmodelserror.BadArgs, nil)
	}

	return nil
}

func ValidateClear(value *string, clear bool) error {
	if value != nil && clear {
		return domainmodelserror.NewError("value and clear flag cannot both be supplied", domainmodelserror.BadArgs, nil)
	}

	return nil
}
