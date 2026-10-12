package domainusecasesmisc

import "context"

// SigningKey reconciles persisted asymmetric token-signing keys: it creates the
// active key when missing, rotates a stale active key, re-seals keys under the
// current key-encryption key, and removes keys past retention.
type SigningKey interface {
	Reconcile(ctx context.Context) error
}
