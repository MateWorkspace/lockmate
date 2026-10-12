package applicationmiscsigningkey

import (
	"context"
	"errors"
	"time"

	domaincontractsrepository "github.com/MateWorkspace/lockmate/internal/domain/contracts/repository"
	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainmodelslogger "github.com/MateWorkspace/lockmate/internal/domain/models/logger"
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
)

const tagPath = "misc/signing_key"

type Config struct {
	RotationInterval  time.Duration
	RetentionInterval time.Duration
}

type usecase struct {
	keys       domaincontractsrepository.SigningKey
	secret     domaincontractsutility.Secret
	transactor domaincontractsutility.Transactor
	logger     domaincontractsutility.Logger
	config     Config
}

func NewUsecaseImpl(
	keys domaincontractsrepository.SigningKey,
	secret domaincontractsutility.Secret,
	transactor domaincontractsutility.Transactor,
	logger domaincontractsutility.Logger,
	config Config,
) domainusecasesmisc.SigningKey {
	return &usecase{
		keys:       keys,
		secret:     secret,
		transactor: transactor,
		logger:     logger,
		config:     config,
	}
}

func (u *usecase) Reconcile(ctx context.Context) error {
	const tag = tagPath + "/Reconcile"

	if err := u.resealStale(ctx, tag); err != nil {
		u.logger.Error(ctx, tag, "failed to re-seal signing keys under the current encryption key", domainmodelslogger.Meta{"err": err})

		return err
	}

	if err := u.ensureActive(ctx, tag); err != nil {
		u.logger.Error(ctx, tag, "failed to reconcile the active signing key", domainmodelslogger.Meta{"err": err})

		return err
	}

	if err := u.retireExpired(ctx, tag); err != nil {
		u.logger.Error(ctx, tag, "failed to retire expired signing keys", domainmodelslogger.Meta{"err": err})

		return err
	}

	return nil
}

func (u *usecase) resealStale(ctx context.Context, tag string) error {
	stale, err := u.keys.ListStale(ctx, u.secret.Fingerprint())
	if err != nil {
		return err
	}

	for _, value := range stale {
		plaintext, err := u.secret.Open(value.PrivateKeyCipher, value.PrivateKeyNonce, value.KeyFingerprint)
		if err != nil {
			return err
		}

		cipher, nonce, err := u.secret.Seal(plaintext)
		if err != nil {
			return err
		}

		if err := u.keys.UpdateSealed(ctx, value.Id, u.secret.Fingerprint(), cipher, nonce); err != nil {
			return err
		}

		u.logger.Info(ctx, tag, "re-sealed signing key under the current encryption key", domainmodelslogger.Meta{"kid": value.Kid, "signing_key_id": value.Id})
	}

	return nil
}

func (u *usecase) ensureActive(ctx context.Context, tag string) error {
	active, err := u.keys.ReadActive(ctx)
	if err != nil {
		if !errors.Is(err, domainmodelserror.SigningKeyNotFound) {
			return err
		}

		return u.createActive(ctx, tag)
	}

	if time.Since(active.NotBefore) < u.config.RotationInterval {
		return nil
	}

	return u.rotateActive(ctx, tag, active)
}

func (u *usecase) createActive(ctx context.Context, tag string) error {
	for attempt := 0; attempt < 2; attempt++ {
		value, err := u.buildKey()
		if err != nil {
			return err
		}

		var id int64
		err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
			id, err = u.keys.Create(txCtx, value)

			return err
		})
		if err == nil {
			u.logger.Info(ctx, tag, "created signing key", domainmodelslogger.Meta{"kid": value.Kid, "signing_key_id": id})

			return nil
		}

		if !errors.Is(err, domainmodelserror.SigningKeyConflict) {
			return err
		}

		// Another instance may have created the active key first; treat that as success.
		if existing, readErr := u.keys.ReadActive(ctx); readErr == nil && existing != nil {
			return nil
		}

		if attempt == 1 {
			return err
		}

		u.logger.Warn(ctx, tag, "retrying signing key creation after conflict", domainmodelslogger.Meta{"err": err, "next_attempt": attempt + 2})
	}

	return nil
}

func (u *usecase) rotateActive(ctx context.Context, tag string, active *domainmodelsentity.SigningKey) error {
	value, err := u.buildKey()
	if err != nil {
		return err
	}

	var id int64
	err = u.transactor.WithTx(ctx, func(txCtx context.Context) error {
		if _, err := u.keys.DemoteActive(txCtx, time.Now().UTC(), nil); err != nil {
			return err
		}

		id, err = u.keys.Create(txCtx, value)

		return err
	})
	if errors.Is(err, domainmodelserror.SigningKeyNotFound) {
		// Another instance rotated the active key first.
		return nil
	}
	if err != nil {
		return err
	}

	u.logger.Info(ctx, tag, "rotated signing key", domainmodelslogger.Meta{"kid": value.Kid, "signing_key_id": id, "previous_kid": active.Kid})

	return nil
}

func (u *usecase) retireExpired(ctx context.Context, tag string) error {
	cutoff := time.Now().UTC().Add(-u.config.RetentionInterval)
	count, err := u.keys.DeleteRetiringBefore(ctx, cutoff, nil)
	if err != nil {
		return err
	}

	if count > 0 {
		u.logger.Info(ctx, tag, "retired expired signing keys", domainmodelslogger.Meta{"count": count})
	}

	return nil
}

func (u *usecase) buildKey() (domainmodelsentity.SigningKey, error) {
	privateKey, publicJWK, kid, algorithm, err := u.secret.GenerateKeypair()
	if err != nil {
		return domainmodelsentity.SigningKey{}, err
	}

	cipher, nonce, err := u.secret.Seal(privateKey)
	if err != nil {
		return domainmodelsentity.SigningKey{}, err
	}

	return domainmodelsentity.SigningKey{
		Kid:              kid,
		Algorithm:        algorithm,
		PrivateKeyCipher: cipher,
		PrivateKeyNonce:  nonce,
		KeyFingerprint:   u.secret.Fingerprint(),
		PublicJWK:        publicJWK,
		Status:           domainmodelsentity.SigningKeyStatusActive,
		NotBefore:        time.Now().UTC(),
	}, nil
}
