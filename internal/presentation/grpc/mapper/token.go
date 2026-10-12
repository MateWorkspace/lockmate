package presentationgrpcmapper

import (
	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainusecasesauth "github.com/MateWorkspace/lockmate/internal/domain/usecases/auth"
	domainusecasesprofile "github.com/MateWorkspace/lockmate/internal/domain/usecases/profile"
)

func Token(value domainusecasesauth.TokenResult) *lockmatev1.TokenResponse {
	return &lockmatev1.TokenResponse{
		AccessToken:  value.AccessToken,
		RefreshToken: value.RefreshToken,
		IsNewUser:    value.IsNewUser,
	}
}

func Profile(value domainusecasesprofile.ProfileResult) *lockmatev1.Profile {
	return &lockmatev1.Profile{
		User:        User(value.User),
		Space:       Space(value.Space),
		SpaceMember: SpaceMember(value.SpaceMember),
	}
}

func AuthMethods(value domainusecasesprofile.AuthMethodsResult) *lockmatev1.AuthMethods {
	result := &lockmatev1.AuthMethods{}
	if value.ClassicAuth != nil {
		result.ClassicAuth = ClassicAuth(*value.ClassicAuth)
	}
	if value.GoogleAuth != nil {
		result.GoogleAuth = GoogleAuth(*value.GoogleAuth)
	}
	return result
}
