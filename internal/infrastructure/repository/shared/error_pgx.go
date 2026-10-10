package infrastructurerepositoryshared

import (
	"context"
	"errors"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type ConflictMatch struct {
	Constraint string
	Type       domainmodelserror.Type
}

func MapPgxError(message string, err error, conflicts ...ConflictMatch) error {
	if err == nil {
		return nil
	}
	var model *domainmodelserror.Error
	if errors.As(err, &model) {
		return err
	}
	kind := domainmodelserror.Failure
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		kind = domainmodelserror.Timeout
	case errors.Is(err, context.Canceled):
		kind = domainmodelserror.BadState
	case errors.Is(err, pgx.ErrNoRows):
		kind = domainmodelserror.NotFound
	default:
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				kind = domainmodelserror.Conflict
				for _, match := range conflicts {
					if pgErr.ConstraintName == match.Constraint {
						kind = match.Type
						break
					}
				}
			case "23503":
				kind = domainmodelserror.Conflict
			case "22001", "22P02", "23502", "23514":
				kind = domainmodelserror.BadArgs
			}
		}
	}

	return domainmodelserror.NewError(message, kind, err)
}

func QueryBuildError(message string, err error) error {
	return domainmodelserror.NewError(message, domainmodelserror.Failure, err)
}

func ReadError(message string, err error, notFound domainmodelserror.Type) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return domainmodelserror.NewError(message, notFound, err)
	}

	return MapPgxError(message, err)
}

func NotFound(message string, kind domainmodelserror.Type) error {
	return domainmodelserror.NewError(message, kind, nil)
}
