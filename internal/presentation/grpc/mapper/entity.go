package presentationgrpcmapper

import (
	lockmatev1 "github.com/MateWorkspace/lockmate-proto/gen/go/lockmate/v1"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainusecasesshared "github.com/MateWorkspace/lockmate/internal/domain/usecases/shared"
)

func User(value domainmodelsentity.User) *lockmatev1.User {
	return &lockmatev1.User{
		Id:          value.Id,
		Name:        value.Name,
		Bio:         value.Bio,
		AvatarPath:  value.AvatarPath,
		Preferences: Preferences(value.Preferences),
		Audit:       Audit(value.AuditCreateUpdateDelete),
	}
}

func Space(value domainmodelsentity.Space) *lockmatev1.Space {
	return &lockmatev1.Space{
		Id:          value.Id,
		Slug:        value.Slug,
		Name:        value.Name,
		Description: value.Description,
		IsActive:    value.IsActive,
		Preferences: Preferences(value.Preferences),
		Audit:       Audit(value.AuditCreateUpdateDelete),
	}
}

func SpaceMember(value domainmodelsentity.SpaceMember) *lockmatev1.SpaceMember {
	return &lockmatev1.SpaceMember{
		Id:          value.Id,
		SpaceId:     value.SpaceId,
		UserId:      value.UserId,
		IsActive:    value.IsActive,
		Preferences: Preferences(value.Preferences),
		Audit:       Audit(value.AuditCreateUpdateDelete),
	}
}

func Role(value domainmodelsentity.Role) *lockmatev1.Role {
	return &lockmatev1.Role{
		Id:          value.Id,
		SpaceId:     value.SpaceId,
		Slug:        value.Slug,
		Name:        value.Name,
		Description: value.Description,
		IsDefault:   value.IsDefault,
		Preferences: Preferences(value.Preferences),
		Audit:       Audit(value.AuditCreateUpdateDelete),
	}
}

func Permission(value domainmodelsentity.Permission) *lockmatev1.Permission {
	return &lockmatev1.Permission{
		Id:          value.Id,
		SpaceId:     value.SpaceId,
		Slug:        value.Slug,
		Name:        value.Name,
		Description: value.Description,
		Preferences: Preferences(value.Preferences),
		Audit:       Audit(value.AuditCreateUpdateDelete),
	}
}

func Client(value domainusecasesshared.ClientView) *lockmatev1.Client {
	return &lockmatev1.Client{
		Id:                value.Id,
		SpaceId:           value.SpaceId,
		Slug:              value.Slug,
		Name:              value.Name,
		Description:       value.Description,
		AuthRedirectUri:   value.AuthRedirectUri,
		SecretKeyRedacted: value.SecretKeyRedacted,
		IsActive:          value.IsActive,
		Preferences:       Preferences(value.Preferences),
		Audit:             Audit(value.AuditCreateUpdateDelete),
	}
}

func ClientIpAllowlist(value domainmodelsentity.ClientIpAllowlist) *lockmatev1.ClientIpAllowlist {
	return &lockmatev1.ClientIpAllowlist{
		Id:       value.Id,
		ClientId: value.ClientId,
		Cidr:     value.Cidr,
		IsActive: value.IsActive,
		Audit:    Audit(value.AuditCreateUpdateDelete),
	}
}

func ClassicAuth(value domainusecasesshared.ClassicAuthView) *lockmatev1.ClassicAuth {
	return &lockmatev1.ClassicAuth{
		Id:              value.Id,
		UserId:          value.UserId,
		Username:        value.Username,
		Email:           value.Email,
		Phone:           value.Phone,
		IsEmailVerified: value.IsEmailVerified,
		IsPhoneVerified: value.IsPhoneVerified,
		Preferences:     Preferences(value.Preferences),
		Audit:           Audit(value.AuditCreateUpdateDelete),
	}
}

func GoogleAuth(value domainmodelsentity.GoogleAuth) *lockmatev1.GoogleAuth {
	return &lockmatev1.GoogleAuth{
		Id:              value.Id,
		UserId:          value.UserId,
		Subject:         value.Subject,
		Email:           value.Email,
		IsEmailVerified: value.IsEmailVerified,
		Preferences:     Preferences(value.Preferences),
		Audit:           Audit(value.AuditCreateUpdateDelete),
	}
}

func RolePermission(value domainmodelsentity.RolePermission) *lockmatev1.RolePermission {
	return &lockmatev1.RolePermission{
		Id:           value.Id,
		SpaceId:      value.SpaceId,
		RoleId:       value.RoleId,
		PermissionId: value.PermissionId,
		Audit:        AuditCreateOnly(value.AuditCreate),
	}
}

func SpaceMemberRole(value domainmodelsentity.SpaceMemberRole) *lockmatev1.SpaceMemberRole {
	return &lockmatev1.SpaceMemberRole{
		Id:            value.Id,
		SpaceId:       value.SpaceId,
		SpaceMemberId: value.SpaceMemberId,
		RoleId:        value.RoleId,
		Audit:         AuditCreateOnly(value.AuditCreate),
	}
}

func Roles(values []domainmodelsentity.Role) []*lockmatev1.Role {
	result := make([]*lockmatev1.Role, 0, len(values))
	for _, value := range values {
		result = append(result, Role(value))
	}
	return result
}

func Permissions(values []domainmodelsentity.Permission) []*lockmatev1.Permission {
	result := make([]*lockmatev1.Permission, 0, len(values))
	for _, value := range values {
		result = append(result, Permission(value))
	}
	return result
}

func Spaces(values []domainmodelsentity.Space) []*lockmatev1.Space {
	result := make([]*lockmatev1.Space, 0, len(values))
	for _, value := range values {
		result = append(result, Space(value))
	}
	return result
}

func Users(values []domainmodelsentity.User) []*lockmatev1.User {
	result := make([]*lockmatev1.User, 0, len(values))
	for _, value := range values {
		result = append(result, User(value))
	}
	return result
}

func Clients(values []domainusecasesshared.ClientView) []*lockmatev1.Client {
	result := make([]*lockmatev1.Client, 0, len(values))
	for _, value := range values {
		result = append(result, Client(value))
	}
	return result
}

func ClassicAuths(values []domainusecasesshared.ClassicAuthView) []*lockmatev1.ClassicAuth {
	result := make([]*lockmatev1.ClassicAuth, 0, len(values))
	for _, value := range values {
		result = append(result, ClassicAuth(value))
	}
	return result
}

func GoogleAuths(values []domainmodelsentity.GoogleAuth) []*lockmatev1.GoogleAuth {
	result := make([]*lockmatev1.GoogleAuth, 0, len(values))
	for _, value := range values {
		result = append(result, GoogleAuth(value))
	}
	return result
}

func ClientIpAllowlists(values []domainmodelsentity.ClientIpAllowlist) []*lockmatev1.ClientIpAllowlist {
	result := make([]*lockmatev1.ClientIpAllowlist, 0, len(values))
	for _, value := range values {
		result = append(result, ClientIpAllowlist(value))
	}
	return result
}
