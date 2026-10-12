package applicationshared

import (
	"context"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

// ResolveSigningKey reads the active signing key and opens its sealed private half.
func ResolveSigningKey(ctx context.Context, keys domaincontractsrepository.SigningKey, secret domaincontractsutility.Secret) (kid string, privateKeyPKCS8 []byte, err error) {
	active, err := keys.ReadActive(ctx)
	if err != nil {
		return "", nil, err
	}

	if active == nil || active.Status != domainmodelsentity.SigningKeyStatusActive || active.Kid == "" || active.AuditDelete.At != nil {
		return "", nil, domainmodelserror.NewError("active signing key is missing", domainmodelserror.SigningKeyNotFound, nil)
	}

	privateKeyPKCS8, err = secret.Open(active.PrivateKeyCipher, active.PrivateKeyNonce, active.KeyFingerprint)
	if err != nil {
		return "", nil, err
	}

	return active.Kid, privateKeyPKCS8, nil
}

// VerifiableKeys returns the active and retiring keys usable for token validation.
func VerifiableKeys(ctx context.Context, keys domaincontractsrepository.SigningKey) ([]domainmodelsentity.SigningKey, error) {
	values, err := keys.ListVerifiable(ctx)
	if err != nil {
		return nil, IdentityError(err)
	}

	if values == nil {
		values = []domainmodelsentity.SigningKey{}
	}

	return values, nil
}
