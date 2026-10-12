package presentationhttphandlerauth

import (
	"net/http"

	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

type jwksHandler struct {
	jwks domainusecasesmisc.Jwks
}

func NewJwksHandler(jwks domainusecasesmisc.Jwks) *jwksHandler {
	return &jwksHandler{jwks: jwks}
}

// JwksGet godoc
//
// @Summary Read the public JSON Web Key Set
// @Tags 01. Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Success 200 {object} presentationhttpresponse.JwksResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/jwks [get]
func (h *jwksHandler) JwksGet(c *gin.Context) {
	keys, err := h.jwks.PublicKeys(c.Request.Context())
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, presentationhttpresponse.Jwks(keys))
}
