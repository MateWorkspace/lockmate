package presentationhttphandlerauth

import (
	"net/http"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// AuthExchangePost godoc
//
// @Summary Exchange a single-use authentication handoff for a session
// @Tags 01. Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Accept json
// @Param request body presentationhttprequest.AuthExchangeRequest true "Request"
// @Success 200 {object} presentationhttpresponse.SessionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 501 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/exchange [post]
func (h *flowHandler) AuthExchangePost(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	if h.handoff == nil {
		presentationhttputils.Error(c, domainmodelserror.NewError("authentication handoff is unavailable", domainmodelserror.Unimplemented, nil))
		return
	}

	var req presentationhttprequest.AuthExchangeRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.handoff.Exchange(c.Request.Context(), presentationhttputils.ClientIdentity(c.Request.Context()), domainusecasesauth.ExchangeAuthHandoffRequest{Code: req.Code})
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("session"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.Session(*result))
}
