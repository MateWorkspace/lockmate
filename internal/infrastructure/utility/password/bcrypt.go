package infrastructureutilitypassword

import (
	"errors"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"golang.org/x/crypto/bcrypt"
)

type bcryptImpl struct {
	cost int
}

func NewBcryptImpl(cost int) domaincontractsutility.Password {
	if cost < bcrypt.MinCost || cost > bcrypt.MaxCost {
		cost = bcrypt.DefaultCost
	}

	return &bcryptImpl{
		cost: cost,
	}
}

func (b *bcryptImpl) Hash(password string) (hashed string, err error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), b.cost)
	if err != nil {
		return "", domainmodelserror.NewError("failed to hash password", domainmodelserror.PasswordHashFailed, err)
	}
	return string(hash), nil
}

func (b *bcryptImpl) Compare(storedHash string, password string) (err error) {
	if len(password) > 72 {
		return domainmodelserror.NewError("password exceeds bcrypt's byte limit", domainmodelserror.PasswordCompareFailed, bcrypt.ErrPasswordTooLong)
	}

	if err := bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password)); err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return domainmodelserror.NewError("password does not match", domainmodelserror.PasswordMismatch, err)
		}
		return domainmodelserror.NewError("failed to compare password", domainmodelserror.PasswordCompareFailed, err)
	}

	return nil
}
