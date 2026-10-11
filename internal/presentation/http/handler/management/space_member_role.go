package presentationhttphandlermanagement

import (
	"net/http"

	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
	presentationhttprequest "github.com/MateWorkspace/lockmate/internal/presentation/http/request"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

// SpaceMemberRolePost godoc
//
// @Summary Create space member role
// @Tags 13. Space Member Roles
// @Accept json
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param request body presentationhttprequest.ManagementCreateSpaceMemberRoleRequest true "Request"
// @Success 201 {object} presentationhttpresponse.IdResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles [post]
func (h *handler) SpaceMemberRolePost(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	var req presentationhttprequest.ManagementCreateSpaceMemberRoleRequest
	if err := presentationhttputils.Bind(c, &req); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.CreateSpaceMemberRoleRequest{
		SpaceMemberId: req.SpaceMemberId,
		RoleId:        req.RoleId,
	}

	id, err := h.spacememberroleUseCase.Create(ctx, caller, spaceId, request)
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

// SpaceMemberRoleGetById godoc
//
// @Summary Read space member role by id
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_role_id path int true "space_member_role_id"
// @Success 200 {object} presentationhttpresponse.SpaceMemberRoleDetailsResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles/{space_member_role_id} [get]
func (h *handler) SpaceMemberRoleGetById(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "space_member_role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberroleUseCase.ReadById(ctx, caller, spaceId, id)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space_member_role_details"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMemberRoleDetails(*result))
}

// SpaceMemberRoleGetByPair godoc
//
// @Summary Read space member role by pair
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id query int true "space_member_id"
// @Param role_id query int true "role_id"
// @Success 200 {object} presentationhttpresponse.SpaceMemberRoleDetailsResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles/by-pair [get]
func (h *handler) SpaceMemberRoleGetByPair(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	spaceMemberId, err := presentationhttputils.RequiredInt64(c.Query("space_member_id"), "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	roleId, err := presentationhttputils.RequiredInt64(c.Query("role_id"), "role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberroleUseCase.ReadBySpaceMemberIdAndRoleId(ctx, caller, spaceId, spaceMemberId, roleId)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if result == nil {
		presentationhttputils.Error(c, presentationhttputils.MissingResponse("space_member_role_details"))
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMemberRoleDetails(*result))
}

// SpaceMemberRoleGetBySpaceMemberId godoc
//
// @Summary Read space member role by space member id
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id path int true "space_member_id"
// @Success 200 {array} presentationhttpresponse.SpaceMemberRoleWithRoleResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles/by-space-member/{space_member_id} [get]
func (h *handler) SpaceMemberRoleGetBySpaceMemberId(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	value, err := presentationhttputils.ParamInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberroleUseCase.ReadBySpaceMemberId(ctx, caller, spaceId, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMemberRoleWithRoleList(result))
}

// SpaceMemberRoleGetByRoleId godoc
//
// @Summary Read space member role by role id
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param role_id path int true "role_id"
// @Success 200 {array} presentationhttpresponse.SpaceMemberRoleWithSpaceMemberResponse
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles/by-role/{role_id} [get]
func (h *handler) SpaceMemberRoleGetByRoleId(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	value, err := presentationhttputils.ParamInt64(c, "role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	result, err := h.spacememberroleUseCase.ReadByRoleId(ctx, caller, spaceId, value)
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.JSON(http.StatusOK, presentationhttpresponse.SpaceMemberRoleWithSpaceMemberList(result))
}

// SpaceMemberRoleDelete godoc
//
// @Summary Delete space member role
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_role_id path int true "space_member_role_id"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles/{space_member_role_id} [delete]
func (h *handler) SpaceMemberRoleDelete(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	id, err := presentationhttputils.ParamInt64(c, "space_member_role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	if err := h.spacememberroleUseCase.DeleteById(ctx, caller, spaceId, id); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}

// SpaceMemberRoleDeleteBySelectors godoc
//
// @Summary Read space member role delete by selectors
// @Tags 13. Space Member Roles
// @Produce json
// @Security ClientSlug && ClientSecret && ClientExecId && BearerAuth
// @Param X-Client-Exec-Id header string true "Nonzero execution UUID" format(uuid)
// @Param space_id path int true "space_id"
// @Param space_member_id query int false "Deletion selector (two selectors use OR)"
// @Param role_id query int false "Deletion selector (two selectors use OR)"
// @Success 204 "Operation completed"
// @Failure 400 {object} presentationhttpresponse.ErrorResponse
// @Failure 401 {object} presentationhttpresponse.ErrorResponse
// @Failure 403 {object} presentationhttpresponse.ErrorResponse
// @Failure 404 {object} presentationhttpresponse.ErrorResponse
// @Failure 409 {object} presentationhttpresponse.ErrorResponse
// @Failure 412 {object} presentationhttpresponse.ErrorResponse
// @Failure 500 {object} presentationhttpresponse.ErrorResponse
// @Failure 504 {object} presentationhttpresponse.ErrorResponse
// @Router /v1/management/spaces/{space_id}/space-member-roles [delete]
func (h *handler) SpaceMemberRoleDeleteBySelectors(c *gin.Context) {
	ctx := c.Request.Context()
	caller := presentationhttputils.AccessClaims(ctx)
	spaceId, err := presentationhttputils.ParamInt64(c, "space_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	spaceMemberId, err := presentationhttputils.QueryInt64(c, "space_member_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	roleId, err := presentationhttputils.QueryInt64(c, "role_id")
	if err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	request := domainusecasesmanagement.DeleteSpaceMemberRoleRequest{
		SpaceMemberId: spaceMemberId,
		RoleId:        roleId,
	}

	if err := h.spacememberroleUseCase.DeleteBySpaceMemberIdOrRoleId(ctx, caller, spaceId, request); err != nil {
		presentationhttputils.Error(c, err)
		return
	}

	c.Status(http.StatusNoContent)
}
