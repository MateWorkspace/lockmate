package domainusecasesmisc

import "context"

// AuthFlow maintains internal expiring authentication records.
type AuthFlow interface {
	PurgeExpired(ctx context.Context) error
}
