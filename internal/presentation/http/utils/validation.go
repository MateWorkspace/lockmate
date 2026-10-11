package presentationhttputils

import (
	"strconv"
	"strings"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequiredString(value string, field string) (string, error) {
	if strings.TrimSpace(value) == "" {
		return "", domainmodelserror.NewError(field+" is required", domainmodelserror.BadArgs, nil)
	}
	return value, nil
}

func RequiredUUID(value, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(value)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, domainmodelserror.NewError(field+" must be a nonzero UUID", domainmodelserror.BadArgs, nil)
	}
	return id, nil
}

func RequiredInt64(value, field string) (int64, error) {
	id, err := strconv.ParseInt(strings.TrimSpace(value), 10, 64)
	if err != nil {
		return 0, domainmodelserror.NewError(field+" must be a valid integer", domainmodelserror.BadArgs, nil)
	}
	return id, nil
}

func ParamInt64(c *gin.Context, field string) (int64, error) {
	return RequiredInt64(c.Param(field), field)
}

func RequiredQueryString(c *gin.Context, field string) (string, error) {
	return RequiredString(c.Query(field), field)
}

func QueryString(c *gin.Context, field string) *string {
	value := strings.TrimSpace(c.Query(field))
	if value == "" {
		return nil
	}
	return &value
}

func QueryInt64(c *gin.Context, field string) (*int64, error) {
	value := strings.TrimSpace(c.Query(field))
	if value == "" {
		return nil, nil
	}
	parsed, err := RequiredInt64(value, field)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func QueryBool(c *gin.Context, field string) (*bool, error) {
	value := strings.TrimSpace(c.Query(field))
	if value == "" {
		return nil, nil
	}
	parsed, err := strconv.ParseBool(value)
	if err != nil {
		return nil, domainmodelserror.NewError(field+" must be a valid boolean", domainmodelserror.BadArgs, nil)
	}
	return &parsed, nil
}
