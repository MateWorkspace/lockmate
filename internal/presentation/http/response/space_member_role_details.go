package presentationhttpresponse

import (
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

type SpaceMemberRoleDetailsResponse struct {
	SpaceMemberRole SpaceMemberRoleResponse `json:"space_member_role"`
	SpaceMember     SpaceMemberResponse     `json:"space_member"`
	Role            RoleResponse            `json:"role"`
}

func SpaceMemberRoleDetails(value domainusecasesshared.SpaceMemberRoleDetails) SpaceMemberRoleDetailsResponse {
	return SpaceMemberRoleDetailsResponse{
		SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
		SpaceMember:     SpaceMember(value.SpaceMember),
		Role:            Role(value.Role),
	}
}

func SpaceMemberRoleDetailsList(values []domainusecasesshared.SpaceMemberRoleDetails) []SpaceMemberRoleDetailsResponse {
	result := make([]SpaceMemberRoleDetailsResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberRoleDetails(value))
	}
	return result
}

type SpaceMemberRoleWithSpaceMemberResponse struct {
	SpaceMemberRole SpaceMemberRoleResponse `json:"space_member_role"`
	SpaceMember     SpaceMemberResponse     `json:"space_member"`
}

func SpaceMemberRoleWithSpaceMember(value domainusecasesshared.SpaceMemberRoleWithSpaceMember) SpaceMemberRoleWithSpaceMemberResponse {
	return SpaceMemberRoleWithSpaceMemberResponse{
		SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
		SpaceMember:     SpaceMember(value.SpaceMember),
	}
}

func SpaceMemberRoleWithSpaceMemberList(values []domainusecasesshared.SpaceMemberRoleWithSpaceMember) []SpaceMemberRoleWithSpaceMemberResponse {
	result := make([]SpaceMemberRoleWithSpaceMemberResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberRoleWithSpaceMember(value))
	}
	return result
}

type SpaceMemberRoleWithRoleResponse struct {
	SpaceMemberRole SpaceMemberRoleResponse `json:"space_member_role"`
	Role            RoleResponse            `json:"role"`
}

func SpaceMemberRoleWithRole(value domainusecasesshared.SpaceMemberRoleWithRole) SpaceMemberRoleWithRoleResponse {
	return SpaceMemberRoleWithRoleResponse{
		SpaceMemberRole: SpaceMemberRole(value.SpaceMemberRole),
		Role:            Role(value.Role),
	}
}

func SpaceMemberRoleWithRoleList(values []domainusecasesshared.SpaceMemberRoleWithRole) []SpaceMemberRoleWithRoleResponse {
	result := make([]SpaceMemberRoleWithRoleResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberRoleWithRole(value))
	}
	return result
}
