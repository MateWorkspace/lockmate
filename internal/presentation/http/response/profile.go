package presentationhttpresponse

import (
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

type ProfileResponse struct {
	User        UserResponse        `json:"user"`
	Space       SpaceResponse       `json:"space"`
	SpaceMember SpaceMemberResponse `json:"space_member"`
}

type AuthMethodsResponse struct {
	ClassicAuth *ClassicAuthResponse `json:"classic_auth,omitempty"`
	GoogleAuth  *GoogleAuthResponse  `json:"google_auth,omitempty"`
}

func Profile(value domainusecasesprofile.ProfileResult) ProfileResponse {
	return ProfileResponse{User: User(value.User), Space: Space(value.Space), SpaceMember: SpaceMember(value.SpaceMember)}
}

func AuthMethods(value domainusecasesprofile.AuthMethodsResult) AuthMethodsResponse {
	result := AuthMethodsResponse{}
	if value.ClassicAuth != nil {
		auth := ClassicAuth(*value.ClassicAuth)
		result.ClassicAuth = &auth
	}
	if value.GoogleAuth != nil {
		auth := GoogleAuth(*value.GoogleAuth)
		result.GoogleAuth = &auth
	}
	return result
}
