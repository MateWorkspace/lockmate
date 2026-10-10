package domaincontractsutility

import domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"

type Token interface {
	GenerateAccess(claims *domainmodelsentity.AccessClaims) (string, error)
	// ValidateAccess checks signature, purpose, timestamps, and expected space.
	ValidateAccess(spaceId int64, token string) (*domainmodelsentity.AccessClaims, error)
	GenerateRefresh(claims *domainmodelsentity.RefreshClaims) (string, error)
	// ValidateRefresh checks the token; application code must recheck live identity
	// and membership before issuing tokens with refreshed grants.
	ValidateRefresh(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error)
}
