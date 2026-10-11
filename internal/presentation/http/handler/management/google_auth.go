package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// GoogleAuthGetById godoc
//
// @Summary Read google auth by id
// @Tags 08. Google Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param google_auth_id path int true "google_auth_id"
// @Success 200 {object} presentationhttpresponse.GoogleAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/google-auths/{google_auth_id} [get]
func (h *handler) GoogleAuthGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "google_auth_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.googleauthUseCase.ReadById(ctx, caller, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("google_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.GoogleAuth(*result))
}

// GoogleAuthGetByUserId godoc
//
// @Summary Read google auth by user id
// @Tags 08. Google Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param user_id path int true "user_id"
// @Success 200 {object} presentationhttpresponse.GoogleAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/google-auths/by-user/{user_id} [get]
func (h *handler) GoogleAuthGetByUserId(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.ParamInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.googleauthUseCase.ReadByUserId(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("google_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.GoogleAuth(*result))
}

// GoogleAuthGetBySubject godoc
//
// @Summary Read google auth by subject
// @Tags 08. Google Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param subject query string true "subject"
// @Success 200 {object} presentationhttpresponse.GoogleAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/google-auths/by-subject [get]
func (h *handler) GoogleAuthGetBySubject(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.RequiredQueryString(c, "subject")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.googleauthUseCase.ReadBySubject(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("google_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.GoogleAuth(*result))
}

// GoogleAuthGetList godoc
//
// @Summary Read google auth list
// @Tags 08. Google Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param user_id query int false "user_id"
// @Param is_email_verified query boolean false "is_email_verified"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.GoogleAuthResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/google-auths [get]
func (h *handler) GoogleAuthGetList(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	userId, err := presentationhttputils.QueryInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	isEmailVerified, err := presentationhttputils.QueryBool(c, "is_email_verified")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.FilterGoogleAuthRequest{
		Page:            page.Page,
		Limit:           page.Limit,
		Search:          page.Search,
		UserId:          userId,
		IsEmailVerified: isEmailVerified,
	}

	result, err := h.googleauthUseCase.ReadByFilter(ctx, caller, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.GoogleAuthResponse]{Data: presentationhttpresponse.GoogleAuthList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}
