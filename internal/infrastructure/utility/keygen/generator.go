package infrastructureutilitykeygen

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"strings"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

const (
	keyPrefix    = "lockmate-"
	randomBytes  = 32
	keyLength    = len(keyPrefix) + randomBytes*2
	redactedSize = 4
)

type generatorImpl struct{}

func NewGeneratorImpl() domaincontractsutility.Keygen {
	return &generatorImpl{}
}

func (g *generatorImpl) Generate() (raw string, hash string, redacted string, err error) {
	buf := make([]byte, randomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", "", "", domainmodelserror.NewError("failed to generate key", domainmodelserror.KeygenFailed, err)
	}
	raw = keyPrefix + hex.EncodeToString(buf)
	hash = g.Hash(raw)
	redacted = raw[:redactedSize] + strings.Repeat("*", len(raw)-2*redactedSize) + raw[len(raw)-redactedSize:]
	return raw, hash, redacted, nil
}

func (g *generatorImpl) Hash(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

func (g *generatorImpl) Validate(raw string) error {
	if len(raw) != keyLength || !strings.HasPrefix(raw, keyPrefix) {
		return domainmodelserror.NewError("key has an invalid format", domainmodelserror.KeyInvalid, nil)
	}
	for _, c := range raw[len(keyPrefix):] {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return domainmodelserror.NewError("key has an invalid format", domainmodelserror.KeyInvalid, nil)
		}
	}
	return nil
}

func (g *generatorImpl) Compare(storedHash string, raw string) error {
	if err := g.Validate(raw); err != nil {
		return err
	}
	if subtle.ConstantTimeCompare([]byte(storedHash), []byte(g.Hash(raw))) != 1 {
		return domainmodelserror.NewError("key does not match", domainmodelserror.KeyMismatch, nil)
	}
	return nil
}
