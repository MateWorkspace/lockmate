package applicationshared

import (
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

func ClassicAuthView(value *domainmodelsentity.ClassicAuth) *domainusecasesshared.ClassicAuthView {
	if value == nil {
		return nil
	}

	return &domainusecasesshared.ClassicAuthView{
		Id:                      value.Id,
		UserId:                  value.UserId,
		Username:                value.Username,
		Email:                   value.Email,
		Phone:                   value.Phone,
		IsEmailVerified:         value.IsEmailVerified,
		IsPhoneVerified:         value.IsPhoneVerified,
		Preferences:             value.Preferences,
		AuditCreateUpdateDelete: value.AuditCreateUpdateDelete,
	}
}

func ClientView(value *domainmodelsentity.Client) *domainusecasesshared.ClientView {
	if value == nil {
		return nil
	}

	return &domainusecasesshared.ClientView{
		Id:                      value.Id,
		AuthRedirectUri:         value.AuthRedirectUri,
		SpaceId:                 value.SpaceId,
		Slug:                    value.Slug,
		Name:                    value.Name,
		Description:             value.Description,
		SecretKeyRedacted:       value.SecretKeyRedacted,
		IsActive:                value.IsActive,
		Preferences:             value.Preferences,
		AuditCreateUpdateDelete: value.AuditCreateUpdateDelete,
	}
}
