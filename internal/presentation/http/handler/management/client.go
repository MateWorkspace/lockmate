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

// ClientOriginPost godoc
//
// @Summary Create client origin
// @Tags 10. Client Origins
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param request body presentationhttprequest.ManagementCreateClientOriginRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins [post]
func (h *handler) ClientOriginPost(c *gin.Context) {
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

	var req presentationhttprequest.ManagementCreateClientOriginRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	origin, err := presentationhttputils.RequiredString(req.Origin, "origin")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateClientOriginRequest{
		Origin:   origin,
		IsActive: req.IsActive,
	}

	id, err := h.clientUseCase.CreateOrigin(ctx, caller, spaceId, clientId, request)
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

// ClientOriginGetById godoc
//
// @Summary Read client origin by id
// @Tags 10. Client Origins
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param origin_id path int true "origin_id"
// @Success 200 {object} presentationhttpresponse.ClientOriginResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins/{origin_id} [get]
func (h *handler) ClientOriginGetById(c *gin.Context) {
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

	id, err := presentationhttputils.ParamInt64(c, "origin_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadOriginById(ctx, caller, spaceId, clientId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_origin"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientOrigin(*result))
}

// ClientOriginGetByValue godoc
//
// @Summary Read client origin by value
// @Tags 10. Client Origins
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param origin query string true "origin"
// @Success 200 {object} presentationhttpresponse.ClientOriginResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins/by-value [get]
func (h *handler) ClientOriginGetByValue(c *gin.Context) {
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

	origin, err := presentationhttputils.RequiredQueryString(c, "origin")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.clientUseCase.ReadOriginByValue(ctx, caller, spaceId, clientId, origin)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("client_origin"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientOrigin(*result))
}

// ClientOriginGetAll godoc
//
// @Summary Read client origin all
// @Tags 10. Client Origins
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Success 200 {array} presentationhttpresponse.ClientOriginResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins/all [get]
func (h *handler) ClientOriginGetAll(c *gin.Context) {
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

	result, err := h.clientUseCase.ReadOriginsByClientId(ctx, caller, spaceId, clientId)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.ClientOriginList(result))
}

// ClientOriginGetList godoc
//
// @Summary Read client origin list
// @Tags 10. Client Origins
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param is_active query boolean false "is_active"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientOriginResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins [get]
func (h *handler) ClientOriginGetList(c *gin.Context) {
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

	request := domainusecasesmanagement.FilterClientOriginRequest{
		Page:     page.Page,
		Limit:    page.Limit,
		Search:   page.Search,
		IsActive: isActive,
	}

	result, err := h.clientUseCase.ReadOriginsByFilter(ctx, caller, spaceId, clientId, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.ClientOriginResponse]{Data: presentationhttpresponse.ClientOriginList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// ClientOriginPatch godoc
//
// @Summary Update client origin
// @Tags 10. Client Origins
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param origin_id path int true "origin_id"
// @Param request body presentationhttprequest.ManagementUpdateClientOriginRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins/{origin_id} [patch]
func (h *handler) ClientOriginPatch(c *gin.Context) {
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

	id, err := presentationhttputils.ParamInt64(c, "origin_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateClientOriginRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateClientOriginRequest{
		Origin:   req.Origin,
		IsActive: req.IsActive,
	}

	if err := h.clientUseCase.UpdateOriginById(ctx, caller, spaceId, clientId, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// ClientOriginDelete godoc
//
// @Summary Delete client origin
// @Tags 10. Client Origins
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param client_id path int true "client_id"
// @Param origin_id path int true "origin_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/clients/{client_id}/origins/{origin_id} [delete]
func (h *handler) ClientOriginDelete(c *gin.Context) {
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

	id, err := presentationhttputils.ParamInt64(c, "origin_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.clientUseCase.DeleteOriginById(ctx, caller, spaceId, clientId, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
