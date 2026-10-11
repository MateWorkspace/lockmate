package presentationhttphandlerauth

import (
	"net/http"

	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// AuthSignInPost godoc
//
// @Summary Sign in using classic credentials
// @Tags 01. Auth
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.AuthSignInRequest true "Request"
// @Success 200 {object} presentationhttpresponse.SessionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/sign-in [post]
func (h *handler) AuthSignInPost(c *gin.Context) {
	ctx := c.Request.Context()
	client := presentationhttputils.ClientIdentity(ctx)

	var req presentationhttprequest.AuthSignInRequest
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

	request := domainusecasesauth.SignInRequest{
		Username: username,
		Password: password,
	}

	result, err := h.sessionUseCase.SignIn(ctx, client, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("session"))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, presentationhttpresponse.Session(*result))
}

// AuthRegisterPost godoc
//
// @Summary Register a user and issue session tokens
// @Tags 01. Auth
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.AuthRegisterRequest true "Request"
// @Success 201 {object} presentationhttpresponse.SessionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/register [post]
func (h *handler) AuthRegisterPost(c *gin.Context) {
	ctx := c.Request.Context()
	client := presentationhttputils.ClientIdentity(ctx)

	var req presentationhttprequest.AuthRegisterRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	name, err := presentationhttputils.RequiredString(req.Name, "name")
	if err != nil {
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

	request := domainusecasesauth.RegisterRequest{
		Name:     name,
		Username: username,
		Password: password,
		Bio:      req.Bio,
		Email:    req.Email,
		Phone:    req.Phone,
	}

	result, err := h.sessionUseCase.Register(ctx, client, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("session"))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusCreated, presentationhttpresponse.Session(*result))
}

// AuthRefreshPost godoc
//
// @Summary Refresh session tokens
// @Tags 01. Auth
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.AuthRefreshRequest true "Request"
// @Success 200 {object} presentationhttpresponse.SessionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/refresh [post]
func (h *handler) AuthRefreshPost(c *gin.Context) {
	ctx := c.Request.Context()
	client := presentationhttputils.ClientIdentity(ctx)

	var req presentationhttprequest.AuthRefreshRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	refreshToken, err := presentationhttputils.RequiredString(req.RefreshToken, "refresh_token")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesauth.RefreshRequest{
		RefreshToken: refreshToken,
	}

	result, err := h.sessionUseCase.Refresh(ctx, client, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("session"))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, presentationhttpresponse.Session(*result))
}
