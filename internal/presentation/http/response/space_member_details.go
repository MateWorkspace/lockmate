package presentationhttpresponse

import (
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

type SpaceMemberWithUserResponse struct {
	SpaceMember SpaceMemberResponse `json:"space_member"`
	User        UserResponse        `json:"user"`
}

func SpaceMemberWithUser(value domainusecasesshared.SpaceMemberWithUser) SpaceMemberWithUserResponse {
	return SpaceMemberWithUserResponse{
		SpaceMember: SpaceMember(value.SpaceMember),
		User:        User(value.User),
	}
}

func SpaceMemberWithUserList(values []domainusecasesshared.SpaceMemberWithUser) []SpaceMemberWithUserResponse {
	result := make([]SpaceMemberWithUserResponse, 0, len(values))
	for _, value := range values {
		result = append(result, SpaceMemberWithUser(value))
	}
	return result
}
