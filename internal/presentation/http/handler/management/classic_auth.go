package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// ClassicAuthPost godoc
//
// @Summary Create classic auth
// @Tags 07. Classic Auth
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.ManagementCreateClassicAuthRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths [post]
func (h *handler) ClassicAuthPost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)

	var req presentationhttprequest.ManagementCreateClassicAuthRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	username, err := presentationhttputils.RequiredString(req.Username, "username")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	password, err := presentationhttputils.RequiredString(req.Password, "password")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateClassicAuthRequest{
		UserId:          req.UserId,
		Username:        username,
		Email:           req.Email,
		Phone:           req.Phone,
		Password:        password,
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
		Preferences:     req.Preferences,
	}

	id, err := h.classicauthUseCase.Create(ctx, caller, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if id <= 0 {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("created identifier"))
		return
	}

	c.JSON(http.StatusCreated, presentationhttpresponse.IdResponse{Id: id})
}

// ClassicAuthGetById godoc
//
// @Summary Read classic auth by id
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param classic_auth_id path int true "classic_auth_id"
// @Success 200 {object} presentationhttpresponse.ClassicAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/{classic_auth_id} [get]
func (h *handler) ClassicAuthGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "classic_auth_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.classicauthUseCase.ReadById(ctx, caller, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("classic_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClassicAuth(*result))
}

// ClassicAuthGetByUserId godoc
//
// @Summary Read classic auth by user id
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param user_id path int true "user_id"
// @Success 200 {object} presentationhttpresponse.ClassicAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/by-user/{user_id} [get]
func (h *handler) ClassicAuthGetByUserId(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.ParamInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.classicauthUseCase.ReadByUserId(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("classic_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClassicAuth(*result))
}

// ClassicAuthGetByUsername godoc
//
// @Summary Read classic auth by username
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param username path string true "username"
// @Success 200 {object} presentationhttpresponse.ClassicAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/by-username/{username} [get]
func (h *handler) ClassicAuthGetByUsername(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.RequiredString(c.Param("username"), "username")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.classicauthUseCase.ReadByUsername(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("classic_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClassicAuth(*result))
}

// ClassicAuthGetByEmail godoc
//
// @Summary Read classic auth by email
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param email query string true "email"
// @Success 200 {object} presentationhttpresponse.ClassicAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/by-email [get]
func (h *handler) ClassicAuthGetByEmail(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.RequiredQueryString(c, "email")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.classicauthUseCase.ReadByEmail(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("classic_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClassicAuth(*result))
}

// ClassicAuthGetByPhone godoc
//
// @Summary Read classic auth by phone
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param phone query string true "phone"
// @Success 200 {object} presentationhttpresponse.ClassicAuthResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/by-phone [get]
func (h *handler) ClassicAuthGetByPhone(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.RequiredQueryString(c, "phone")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.classicauthUseCase.ReadByPhone(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("classic_auth"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClassicAuth(*result))
}

// ClassicAuthGetList godoc
//
// @Summary Read classic auth list
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param user_id query int false "user_id"
// @Param is_email_verified query boolean false "is_email_verified"
// @Param is_phone_verified query boolean false "is_phone_verified"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClassicAuthResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths [get]
func (h *handler) ClassicAuthGetList(c *gin.Context) {
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

	isPhoneVerified, err := presentationhttputils.QueryBool(c, "is_phone_verified")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.FilterClassicAuthRequest{
		Page:            page.Page,
		Limit:           page.Limit,
		Search:          page.Search,
		UserId:          userId,
		IsEmailVerified: isEmailVerified,
		IsPhoneVerified: isPhoneVerified,
	}

	result, err := h.classicauthUseCase.ReadByFilter(ctx, caller, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClassicAuthResponse]{Data: presentationhttpresponse.ClassicAuthList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// ClassicAuthPatch godoc
//
// @Summary Update classic auth
// @Tags 07. Classic Auth
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param classic_auth_id path int true "classic_auth_id"
// @Param request body presentationhttprequest.ManagementUpdateClassicAuthRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/{classic_auth_id} [patch]
func (h *handler) ClassicAuthPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "classic_auth_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateClassicAuthRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateClassicAuthRequest{
		Username:        req.Username,
		Email:           req.Email,
		EmailSetNull:    req.EmailSetNull,
		Phone:           req.Phone,
		PhoneSetNull:    req.PhoneSetNull,
		IsEmailVerified: req.IsEmailVerified,
		IsPhoneVerified: req.IsPhoneVerified,
		Preferences:     req.Preferences,
	}

	if err := h.classicauthUseCase.UpdateById(ctx, caller, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// UserPasswordPatch godoc
//
// @Summary Update user password
// @Tags 06. Users
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param user_id path int true "user_id"
// @Param request body presentationhttprequest.ManagementResetUserPasswordRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/users/{user_id}/password [patch]
func (h *handler) UserPasswordPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	userId, err := presentationhttputils.ParamInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementResetUserPasswordRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	password, err := presentationhttputils.RequiredString(req.Password, "password")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.ResetUserPasswordRequest{
		Password: password,
	}

	if err := h.classicauthUseCase.ResetPassword(ctx, caller, userId, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ClassicAuthDelete godoc
//
// @Summary Delete classic auth
// @Tags 07. Classic Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param classic_auth_id path int true "classic_auth_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/classic-auths/{classic_auth_id} [delete]
func (h *handler) ClassicAuthDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "classic_auth_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.classicauthUseCase.DeleteById(ctx, caller, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
