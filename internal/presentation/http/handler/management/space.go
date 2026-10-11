package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// SpacePost godoc
//
// @Summary Create space
// @Tags 03. Spaces
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param request body presentationhttprequest.ManagementCreateSpaceRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces [post]
func (h *handler) SpacePost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)

	var req presentationhttprequest.ManagementCreateSpaceRequest
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

	request := domainusecasesmanagement.CreateSpaceRequest{
		Slug:        slug,
		Name:        name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Preferences: req.Preferences,
	}

	id, err := h.spaceUseCase.Create(ctx, caller, request)
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

// SpaceGetById godoc
//
// @Summary Read space by id
// @Tags 03. Spaces
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Success 200 {object} presentationhttpresponse.SpaceResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id} [get]
func (h *handler) SpaceGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spaceUseCase.ReadById(ctx, caller, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.Space(*result))
}

// SpaceGetBySlug godoc
//
// @Summary Read space by slug
// @Tags 03. Spaces
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param slug path string true "slug"
// @Success 200 {object} presentationhttpresponse.SpaceResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/by-slug/{slug} [get]
func (h *handler) SpaceGetBySlug(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	value, err := presentationhttputils.RequiredString(c.Param("slug"), "slug")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spaceUseCase.ReadBySlug(ctx, caller, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.Space(*result))
}

// SpaceGetList godoc
//
// @Summary Read space list
// @Tags 03. Spaces
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param is_active query boolean false "is_active"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.SpaceResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces [get]
func (h *handler) SpaceGetList(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
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

	request := domainusecasesmanagement.FilterSpaceRequest{
		Page:     page.Page,
		Limit:    page.Limit,
		Search:   page.Search,
		IsActive: isActive,
	}

	result, err := h.spaceUseCase.ReadByFilter(ctx, caller, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.SpaceResponse]{Data: presentationhttpresponse.SpaceList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// SpacePatch godoc
//
// @Summary Update space
// @Tags 03. Spaces
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param request body presentationhttprequest.ManagementUpdateSpaceRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id} [patch]
func (h *handler) SpacePatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateSpaceRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateSpaceRequest{
		Name:        req.Name,
		Description: req.Description,
		IsActive:    req.IsActive,
		Preferences: req.Preferences,
	}

	if err := h.spaceUseCase.UpdateById(ctx, caller, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SpaceDelete godoc
//
// @Summary Delete space
// @Tags 03. Spaces
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id} [delete]
func (h *handler) SpaceDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	id, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.spaceUseCase.DeleteById(ctx, caller, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
