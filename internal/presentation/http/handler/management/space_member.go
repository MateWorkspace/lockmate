package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// SpaceMemberPost godoc
//
// @Summary Create space member
// @Tags 11. Space Members
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param request body presentationhttprequest.ManagementCreateSpaceMemberRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members [post]
func (h *handler) SpaceMemberPost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementCreateSpaceMemberRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateSpaceMemberRequest{
		UserId:      req.UserId,
		IsActive:    req.IsActive,
		Preferences: req.Preferences,
	}

	id, err := h.spacememberUseCase.Create(ctx, caller, spaceId, request)
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

// SpaceMemberGetById godoc
//
// @Summary Read space member by id
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Success 200 {object} presentationhttpresponse.SpaceMemberResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/{space_member_id} [get]
func (h *handler) SpaceMemberGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberUseCase.ReadById(ctx, caller, spaceId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space_member"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMember(*result))
}

// SpaceMemberGetByUserId godoc
//
// @Summary Read space member by user id
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param user_id path int true "user_id"
// @Success 200 {object} presentationhttpresponse.SpaceMemberResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/by-user/{user_id} [get]
func (h *handler) SpaceMemberGetByUserId(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	value, err := presentationhttputils.ParamInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberUseCase.ReadByUserId(ctx, caller, spaceId, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space_member"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMember(*result))
}

// SpaceMemberGetList godoc
//
// @Summary Read space member list
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Page size, maximum 100" default(10)
// @Param search query string false "Search term"
// @Param space_id path int true "space_id"
// @Param user_id query int false "user_id"
// @Param role_id query int false "role_id"
// @Param is_active query boolean false "is_active"
// @Success 200 {object} presentationhttpresponse.PageDataResponse[presentationhttpresponse.SpaceMemberWithUserResponse]
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members [get]
func (h *handler) SpaceMemberGetList(c *gin.Context) {
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

	userId, err := presentationhttputils.QueryInt64(c, "user_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	roleId, err := presentationhttputils.QueryInt64(c, "role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	isActive, err := presentationhttputils.QueryBool(c, "is_active")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.FilterSpaceMemberRequest{
		Page:     page.Page,
		Limit:    page.Limit,
		Search:   page.Search,
		UserId:   userId,
		RoleId:   roleId,
		IsActive: isActive,
	}

	result, err := h.spacememberUseCase.ReadByFilter(ctx, caller, spaceId, request)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PageDataResponse[presentationhttpresponse.SpaceMemberWithUserResponse]{Data: presentationhttpresponse.SpaceMemberWithUserList(result.Items), Page: presentationhttputils.PageResponse(page, result.Total)})
}

// SpaceMemberPatch godoc
//
// @Summary Update space member
// @Tags 11. Space Members
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Param request body presentationhttprequest.ManagementUpdateSpaceMemberRequest true "Request"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/{space_member_id} [patch]
func (h *handler) SpaceMemberPatch(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementUpdateSpaceMemberRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.UpdateSpaceMemberRequest{
		IsActive:    req.IsActive,
		Preferences: req.Preferences,
	}

	if err := h.spacememberUseCase.UpdateById(ctx, caller, spaceId, id, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SpaceMemberEffectiveRolesGet godoc
//
// @Summary Read space member effective roles
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Success 200 {array} presentationhttpresponse.RoleResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/{space_member_id}/roles [get]
func (h *handler) SpaceMemberEffectiveRolesGet(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	spaceMemberId, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberUseCase.ReadEffectiveRolesBySpaceMemberId(ctx, caller, spaceId, spaceMemberId)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.RoleList(result))
}

// SpaceMemberEffectivePermissionsGet godoc
//
// @Summary Read space member effective permissions
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Success 200 {array} presentationhttpresponse.PermissionResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/{space_member_id}/permissions [get]
func (h *handler) SpaceMemberEffectivePermissionsGet(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	spaceMemberId, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberUseCase.ReadEffectivePermissionsBySpaceMemberId(ctx, caller, spaceId, spaceMemberId)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.PermissionList(result))
}

// SpaceMemberDelete godoc
//
// @Summary Delete space member
// @Tags 11. Space Members
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-members/{space_member_id} [delete]
func (h *handler) SpaceMemberDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.spacememberUseCase.DeleteById(ctx, caller, spaceId, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
