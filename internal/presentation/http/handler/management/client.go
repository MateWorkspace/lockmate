package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// ClientPost godoc
//
// @Summary Create client
// @Tags 09. Clients
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param request body presentationhttprequest.ManagementCreateClientRequest true "Request"
// @Success 201 {object} presentationhttpresponse.ClientSecretResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients [post]
func (h *handler) ClientPost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementCreateClientRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	slug, err := presentationhttputils.RequiredString(req.Slug, "slug")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	name, err := presentationhttputils.RequiredString(req.Name, "name")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateClientRequest{
		AuthRedirectUri: req.AuthRedirectUri,
		Slug:            slug,
		Name:            name,
		Description:     req.Description,
		IsActive:        req.IsActive,
		Preferences:     req.Preferences,
	}

	result, err := h.clientUseCase.Create(ctx, caller, spaceId, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_secret"))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusCreated, presentationhttpresponse.ClientSecret(*result))
}

// ClientGetById godoc
//
// @Summary Read client by id
// @Tags 09. Clients
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Success 200 {object} presentationhttpresponse.ClientResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id} [get]
func (h *handler) ClientGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadById(ctx, caller, spaceId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.Client(*result))
}

// ClientGetBySlug godoc
//
// @Summary Read client by slug
// @Tags 09. Clients
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param slug path string true "slug"
// @Success 200 {object} presentationhttpresponse.ClientResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/by-slug/{slug} [get]
func (h *handler) ClientGetBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	value, err := presentationhttputils.RequiredString(c.Param("slug"), "slug")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadBySlug(ctx, caller, spaceId, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.Client(*result))
}

// ClientGetList godoc
//
// @Summary Read client list
// @Tags 09. Clients
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param space_id path int true "space_id"
// @Param is_active query boolean false "is_active"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients [get]
func (h *handler) ClientGetList(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	isActive, err := presentationhttputils.QueryBool(c, "is_active")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.FilterClientRequest{
		Page:     page.Page,
		Limit:    page.Limit,
		Search:   page.Search,
		IsActive: isActive,
	}

	result, err := h.clientUseCase.ReadByFilter(ctx, caller, spaceId, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientResponse]{Data: presentationhttpresponse.ClientList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// ClientPatch godoc
//
// @Summary Update client
// @Tags 09. Clients
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param request body presentationhttprequest.ManagementUpdateClientRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id} [patch]
func (h *handler) ClientPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateClientRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateClientRequest{
		AuthRedirectUri:        req.AuthRedirectUri,
		SetAuthRedirectUriNull: req.SetAuthRedirectUriNull,
		Name:                   req.Name,
		Description:            req.Description,
		IsActive:               req.IsActive,
		Preferences:            req.Preferences,
	}

	if err := h.clientUseCase.UpdateById(ctx, caller, spaceId, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ClientSecretPatch godoc
//
// @Summary Update client secret
// @Tags 09. Clients
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Success 200 {object} presentationhttpresponse.ClientSecretResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/secret [patch]
func (h *handler) ClientSecretPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.RotateSecret(ctx, caller, spaceId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_secret"))
		return
	}

	c.Header("Cache-Control", "no-store")
	c.JSON(http.StatusOK, presentationhttpresponse.ClientSecret(*result))
}

// ClientDelete godoc
//
// @Summary Delete client
// @Tags 09. Clients
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id} [delete]
func (h *handler) ClientDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.clientUseCase.DeleteById(ctx, caller, spaceId, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ClientIpAllowlistPost godoc
//
// @Summary Create client IP allowlist
// @Tags 10. Client IP Allowlist
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param request body presentationhttprequest.ManagementCreateClientIpAllowlistRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist [post]
func (h *handler) ClientIpAllowlistPost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementCreateClientIpAllowlistRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	cidr, err := presentationhttputils.RequiredString(req.Cidr, "cidr")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateClientIpAllowlistRequest{
		Cidr:     cidr,
		IsActive: req.IsActive,
	}

	id, err := h.clientUseCase.CreateIpAllowlist(ctx, caller, spaceId, clientId, request)
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

// ClientIpAllowlistGetById godoc
//
// @Summary Read client IP allowlist by id
// @Tags 10. Client IP Allowlist
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param ip_allowlist_id path int true "ip_allowlist_id"
// @Success 200 {object} presentationhttpresponse.ClientIpAllowlistResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist/{ip_allowlist_id} [get]
func (h *handler) ClientIpAllowlistGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "ip_allowlist_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadIpAllowlistById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_ip_allowlist"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientIpAllowlist(*result))
}

// ClientIpAllowlistGetByCidr godoc
//
// @Summary Read client IP allowlist by value
// @Tags 10. Client IP Allowlist
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param cidr query string true "cidr"
// @Success 200 {object} presentationhttpresponse.ClientIpAllowlistResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist/by-cidr [get]
func (h *handler) ClientIpAllowlistGetByCidr(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	cidr, err := presentationhttputils.RequiredQueryString(c, "cidr")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadIpAllowlistByCidr(ctx, caller, spaceId, clientId, cidr)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_ip_allowlist"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientIpAllowlist(*result))
}

// ClientIpAllowlistGetAll godoc
//
// @Summary Read client IP allowlist all
// @Tags 10. Client IP Allowlist
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Success 200 {array} presentationhttpresponse.ClientIpAllowlistResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist/all [get]
func (h *handler) ClientIpAllowlistGetAll(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadIpAllowlistsByClientId(ctx, caller, spaceId, clientId)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientIpAllowlistList(result))
}

// ClientIpAllowlistGetList godoc
//
// @Summary Read client IP allowlist list
// @Tags 10. Client IP Allowlist
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param is_active query boolean false "is_active"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientIpAllowlistResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist [get]
func (h *handler) ClientIpAllowlistGetList(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	page, err := presentationhttputils.PageArgs(c)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	isActive, err := presentationhttputils.QueryBool(c, "is_active")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.FilterClientIpAllowlistRequest{
		Page:     page.Page,
		Limit:    page.Limit,
		Search:   page.Search,
		IsActive: isActive,
	}

	result, err := h.clientUseCase.ReadIpAllowlistsByFilter(ctx, caller, spaceId, clientId, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientIpAllowlistResponse]{Data: presentationhttpresponse.ClientIpAllowlistList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// ClientIpAllowlistPatch godoc
//
// @Summary Update client IP allowlist
// @Tags 10. Client IP Allowlist
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param ip_allowlist_id path int true "ip_allowlist_id"
// @Param request body presentationhttprequest.ManagementUpdateClientIpAllowlistRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist/{ip_allowlist_id} [patch]
func (h *handler) ClientIpAllowlistPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "ip_allowlist_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateClientIpAllowlistRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateClientIpAllowlistRequest{
		Cidr:     req.Cidr,
		IsActive: req.IsActive,
	}

	if err := h.clientUseCase.UpdateIpAllowlistById(ctx, caller, spaceId, clientId, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ClientIpAllowlistDelete godoc
//
// @Summary Delete client IP allowlist
// @Tags 10. Client IP Allowlist
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param ip_allowlist_id path int true "ip_allowlist_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/ip-allowlist/{ip_allowlist_id} [delete]
func (h *handler) ClientIpAllowlistDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	clientId, err := presentationhttputils.ParamInt64(c, "client_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "ip_allowlist_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.clientUseCase.DeleteIpAllowlistById(ctx, caller, spaceId, clientId, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
