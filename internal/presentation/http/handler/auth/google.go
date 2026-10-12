package presentationhttphandlerauth

import (
	"net/http"
	"net/url"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// AuthGoogleStartPost godoc
//
// @Summary Start Google sign-in or registration
// @Tags 01. Auth
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Accept json
// @Param request body presentationhttprequest.AuthGoogleStartRequest true "Request"
// @Success 200 {object} presentationhttpresponse.AuthGoogleStartResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 501 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/google/start [post]
func (h *flowHandler) AuthGoogleStartPost(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	if h.google == nil {
		presentationhttputils.Error(c, domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil))
		return
	}

	var req presentationhttprequest.AuthGoogleStartRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.google.Start(c.Request.Context(), presentationhttputils.ClientIdentity(c.Request.Context()), domainusecasesauth.StartGoogleSignInRequest{State: req.State})
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("Google launch URL"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.AuthGoogleStartResponse{LaunchUrl: result.LaunchUrl})
}

// AuthGoogleLaunchGet godoc
//
// @Summary Launch Google authentication in the browser
// @Tags 01. Auth
// @Produce json
// @Param ticket query string true "ticket"
// @Success 303 {string} string "Redirect to Google"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 501 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/google/launch [get]
func (h *flowHandler) AuthGoogleLaunchGet(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	if h.google == nil {
		presentationhttputils.Error(c, domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil))
		return
	}

	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil || len(query["ticket"]) != 1 {
		presentationhttputils.Error(c, domainmodelserror.NewError("one launch ticket is required", domainmodelserror.BadArgs, nil))
		return
	}

	result, err := h.google.Launch(c.Request.Context(), query.Get("ticket"))
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("Google authorization"))
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     presentationhttputils.GoogleBindingCookieName(result.State),
		Value:    result.Binding,
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   result.CookieDuration,
	})
	c.Redirect(http.StatusSeeOther, result.AuthorizationUrl)
}

// AuthGoogleCallbackGet godoc
//
// @Summary Complete Google sign-in and redirect with a handoff code
// @Tags 01. Auth
// @Produce json
// @Param state query string true "state"
// @Param code query string false "code"
// @Param error query string false "error"
// @Success 303 {string} string "Redirect to client backend"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 501 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/auth/google/callback [get]
func (h *flowHandler) AuthGoogleCallbackGet(c *gin.Context) {
	c.Header("Cache-Control", "no-store")
	c.Header("Referrer-Policy", "no-referrer")
	if h.google == nil {
		presentationhttputils.Error(c, domainmodelserror.NewError("Google authentication is disabled", domainmodelserror.Unimplemented, nil))
		return
	}

	query, err := url.ParseQuery(c.Request.URL.RawQuery)
	if err != nil || len(query["state"]) != 1 || len(query["code"]) > 1 || len(query["error"]) > 1 || len(query.Get("state")) != 43 || len(query.Get("code")) > 4096 || len(query.Get("error")) > 256 {
		presentationhttputils.Error(c, domainmodelserror.NewError("invalid Google callback parameters", domainmodelserror.BadArgs, nil))
		return
	}

	cookieName := presentationhttputils.GoogleBindingCookieName(query.Get("state"))
	cookies := c.Request.CookiesNamed(cookieName)
	if len(cookies) != 1 || len(cookies[0].Value) != 43 {
		presentationhttputils.Error(c, domainmodelserror.NewError("Google browser binding is missing or invalid", domainmodelserror.GoogleOAuthStateInvalid, nil))
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Path:     "/",
		Secure:   true,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   -1,
	})
	result, err := h.google.SignInCallback(c.Request.Context(), domainusecasesauth.GoogleSignInCallbackRequest{
		Code:          query.Get("code"),
		State:         query.Get("state"),
		Binding:       cookies[0].Value,
		ProviderError: query.Get("error"),
	})
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("Google callback redirect"))
		return
	}

	c.Redirect(http.StatusSeeOther, result.RedirectUrl)
}
