package presentationhttputils

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"net/http"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	"github.com/gin-gonic/gin"
)

// Bind accepts one JSON object. It does not bind query or path fields into body DTOs.
func Bind(c *gin.Context, value any) error {
	mediaType, _, err := mime.ParseMediaType(c.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return domainmodelserror.NewError("request body must use application/json", domainmodelserror.BadArgs, nil)
	}
	// A bounded read avoids allocating arbitrarily large JSON bodies.
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, 1<<20)
	decoder := json.NewDecoder(c.Request.Body)
	var raw json.RawMessage
	if err := decoder.Decode(&raw); err != nil {
		return domainmodelserror.NewError("request body is invalid", domainmodelserror.BadArgs, nil)
	}
	if trimmed := bytes.TrimSpace(raw); len(trimmed) == 0 || trimmed[0] != '{' {
		return domainmodelserror.NewError("request body must be a JSON object", domainmodelserror.BadArgs, nil)
	}
	if err := json.Unmarshal(raw, value); err != nil {
		return domainmodelserror.NewError("request body is invalid", domainmodelserror.BadArgs, nil)
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		return domainmodelserror.NewError("request body must contain one JSON object", domainmodelserror.BadArgs, nil)
	}
	return nil
}
