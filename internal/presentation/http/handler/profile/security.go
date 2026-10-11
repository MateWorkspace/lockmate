package presentationhttphandlerprofile

import (
	"net/http"

	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// ProfileAuthMethodsGet godoc
//
// @Summary Read profile auth methods
// @Tags 02. Profile
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Success 200 {object} presentationhttpresponse.AuthMethodsResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/profile/auth-methods [get]
func (h *handler) ProfileAuthMethodsGet(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	result, err := h.securityUseCase.GetAuthMethods(ctx, caller)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("auth_methods"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.AuthMethods(*result))
}

// ProfileClassicAuthPatch godoc
//
// @Summary Update profile classic auth
// @Tags 02. Profile
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.ProfileUpdateClassicAuthRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/profile/classic-auth [patch]
func (h *handler) ProfileClassicAuthPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)

	var req presentationhttprequest.ProfileUpdateClassicAuthRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesprofile.UpdateClassicAuthRequest{
		Username:     req.Username,
		Email:        req.Email,
		EmailSetNull: req.EmailSetNull,
		Phone:        req.Phone,
		PhoneSetNull: req.PhoneSetNull,
	}

	if err := h.securityUseCase.UpdateClassicAuth(ctx, caller, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ProfilePasswordPatch godoc
//
// @Summary Update profile password
// @Tags 02. Profile
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.ProfileChangePasswordRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/profile/password [patch]
func (h *handler) ProfilePasswordPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)

	var req presentationhttprequest.ProfileChangePasswordRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	currentPassword, err := presentationhttputils.RequiredString(req.CurrentPassword, "current_password")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	newPassword, err := presentationhttputils.RequiredString(req.NewPassword, "new_password")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesprofile.ChangePasswordRequest{
		CurrentPassword: currentPassword,
		NewPassword:     newPassword,
	}

	if err := h.securityUseCase.ChangePassword(ctx, caller, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
