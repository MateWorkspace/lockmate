package presentationhttp_test

import (
	"errors"
	"fmt"
	"net/http/httptest"
	"strings"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

func TestErrorClassificationAndSanitization(t *testing.T) {
	for _, item := range []struct {
		name           string
		classification error
		status         int
		public         bool
	}{
		{"validation", domainmodelserror.UserPasswordTooShort, 400, true},
		{"password mismatch", domainmodelserror.PasswordMismatch, 400, false},
		{"entity missing", domainmodelserror.ClientIpAllowlistNotFound, 404, true},
		{"conflict", domainmodelserror.ClientIpAllowlistConflict, 409, true},
		{"state", domainmodelserror.BadState, 412, true},
		{"expired", domainmodelserror.TokenExpired, 401, false},
		{"permission", domainmodelserror.Forbidden, 403, true},
		{"timeout", domainmodelserror.Timeout, 504, false},
		{"environment", domainmodelserror.EnvironmentInvalid, 500, false},
		{"hash failure", domainmodelserror.PasswordHashFailed, 500, false},
		{"trace failure", domainmodelserror.TraceEventIdFailed, 500, false},
		{"failure", domainmodelserror.Failure, 500, false},
		{"unknown", errors.New("unclassified"), 500, false},
	} {
		t.Run(item.name, func(t *testing.T) {
			err := fmt.Errorf("wrapper: %w", domainmodelserror.NewError("safe domain message", item.classification, errors.New("private source must stay hidden")))
			router := gin.New()
			router.GET("/error", func(c *gin.Context) { presentationhttputils.Error(c, err) })
			response := perform(router, httptest.NewRequest("GET", "/error", nil))
			if response.Code != item.status || strings.Contains(response.Body.String(), "private source") || strings.Contains(response.Body.String(), "wrapper") {
				t.Fatalf("bad response: %d %s", response.Code, response.Body.String())
			}
			if strings.Contains(response.Body.String(), "safe domain message") != item.public {
				t.Fatal("wrong message policy")
			}
		})
	}
	router := gin.New()
	router.GET("/native", func(c *gin.Context) { presentationhttputils.Error(c, errors.New("private native failure")) })
	if response := perform(router, httptest.NewRequest("GET", "/native", nil)); response.Code != 500 || strings.Contains(response.Body.String(), "private native") {
		t.Fatal("native diagnostic leaked")
	}
}
