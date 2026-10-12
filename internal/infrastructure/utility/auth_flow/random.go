package infrastructureutilityauthflow

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

type randomImpl struct{}

func NewRandomImpl() domaincontractsutility.AuthFlow {
	return &randomImpl{}
}

func (r *randomImpl) Generate() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", domainmodelserror.NewError("failed to generate authentication flow secret", domainmodelserror.Failure, err)
	}

	return base64.RawURLEncoding.EncodeToString(value), nil
}

func (r *randomImpl) Hash(value string) string {
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}
