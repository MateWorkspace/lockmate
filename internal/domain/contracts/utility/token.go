package domaincontractsutility

import domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"

type Token interface {
	// GenerateAccess signs access claims with the supplied ES256 private key.
	GenerateAccess(claims *domainmodelsentity.AccessClaims, kid string, privateKeyPKCS8 []byte) (string, error)
	// ValidateAccess checks signature, purpose, timestamps, and expected space
	// against the supplied verifiable keys, then returns the claims.
	ValidateAccess(keys []domainmodelsentity.SigningKey, token string) (*domainmodelsentity.AccessClaims, error)
	GenerateRefresh(claims *domainmodelsentity.RefreshClaims) (string, error)
	// ValidateRefresh checks the token; application code must recheck live identity
	// and membership before issuing tokens with refreshed grants.
	ValidateRefresh(spaceId int64, token string) (*domainmodelsentity.RefreshClaims, error)
}
