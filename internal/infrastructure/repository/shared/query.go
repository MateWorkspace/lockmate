package infrastructurerepositoryshared

import (
	"context"
	"encoding/json"
	"math"

	"github.com/Masterminds/squirrel"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/MateWorkspace/mate-golang/pgxdt"
)

func PositiveIds(ids ...int64) error {
	for _, id := range ids {
		if id <= 0 {
			return domainmodelserror.NewError("ids must be positive", domainmodelserror.BadArgs, nil)
		}
	}

	return nil
}

func OptionalIds(ids ...*int64) error {
	for _, id := range ids {
		if id != nil {
			if err := PositiveIds(*id); err != nil {
				return err
			}
		}
	}

	return nil
}

func ValidateJSON(value *json.RawMessage) error {
	if value != nil && !json.Valid(*value) {
		return domainmodelserror.NewError("preferences must contain valid JSON", domainmodelserror.BadArgs, nil)
	}

	return nil
}

func Pagination(page, limit int64) (uint64, uint64, error) {
	if limit <= 0 {
		return 0, 0, nil
	}

	if page <= 1 {
		return uint64(limit), 0, nil
	}

	if page-1 > math.MaxInt64/limit {
		return 0, 0, domainmodelserror.NewError("pagination offset is out of range", domainmodelserror.BadArgs, nil)
	}

	return uint64(limit), uint64((page - 1) * limit), nil
}

func SearchPattern(search *string) (string, bool) {
	if search == nil || *search == "" {
		return "", false
	}

	return "%" + *search + "%", true
}

func QualifiedColumns(alias string, columns []string) []string {
	result := make([]string, len(columns))
	for i, column := range columns {
		result[i] = alias + "." + column
	}

	return result
}

// LockRow accepts only adapter-owned table names, never caller input.
func LockRow(ctx context.Context, dt pgxdt.Pgxdt, table string, id int64, spaceId *int64, kind domainmodelserror.Type) error {
	var found int64
	query := "SELECT id FROM " + table + " WHERE id=$1 AND deleted_at IS NULL"
	args := []any{id}
	if spaceId != nil {
		query += " AND space_id=$2"
		args = append(args, *spaceId)
	}
	err := dt.QueryRow(ctx, query+" FOR SHARE", args...).Scan(&found)
	return ReadError("referenced record not found", err, kind)
}

func LockDefault(ctx context.Context, dt pgxdt.Pgxdt, spaceId int64) error {
	_, err := dt.Exec(ctx, "SELECT pg_advisory_xact_lock(hashtextextended('lockmate/default-role/' || $1::bigint::text, 0))", spaceId)
	return MapPgxError("failed to lock default role", err)
}

func ApplySearch(query squirrel.SelectBuilder, search *string, columns ...string) squirrel.SelectBuilder {
	if pattern, ok := SearchPattern(search); ok {
		conditions := squirrel.Or{}
		for _, column := range columns {
			conditions = append(conditions, squirrel.ILike{column: pattern})
		}

		query = query.Where(conditions)
	}

	return query
}
