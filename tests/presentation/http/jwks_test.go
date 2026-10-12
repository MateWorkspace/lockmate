package presentationhttp_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	presentationhttphandlerauth "github.com/MateWorkspace/lockmate/internal/presentation/http/handler/auth"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
)

type jwksStub struct {
	keys []json.RawMessage
	err  error
}

func (s jwksStub) PublicKeys(context.Context) ([]json.RawMessage, error) { return s.keys, s.err }

func jwksRouter(stub jwksStub) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/v1/jwks", presentationhttphandlerauth.NewJwksHandler(stub).JwksGet)
	return router
}

func TestJwksHandlerReturnsKeys(t *testing.T) {
	router := jwksRouter(jwksStub{keys: []json.RawMessage{json.RawMessage(`{"kty":"EC","kid":"a"}`)}})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/jwks", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("jwks: %d %s", recorder.Code, recorder.Body.String())
	}

	var body presentationhttpresponse.JwksResponse
	if err := json.Unmarshal(recorder.Body.Bytes(), &body); err != nil {
		t.Fatal(err)
	}
	if len(body.Keys) != 1 {
		t.Fatalf("unexpected key set: %s", recorder.Body.String())
	}
}

func TestJwksHandlerSanitizesFailure(t *testing.T) {
	router := jwksRouter(jwksStub{err: domainmodelserror.NewError("private-source", domainmodelserror.Failure, nil)})
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/api/v1/jwks", nil)
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("jwks failure: %d %s", recorder.Code, recorder.Body.String())
	}
}
