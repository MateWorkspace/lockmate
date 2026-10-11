package presentationhttphandlermanagement

import (
	domainusecasesmanagement "github.com/MateWorkspace/lockmate/internal/domain/usecases/management"
)

type handler struct {
	spaceUseCase           domainusecasesmanagement.Space
	permissionUseCase      domainusecasesmanagement.Permission
	roleUseCase            domainusecasesmanagement.Role
	userUseCase            domainusecasesmanagement.User
	classicauthUseCase     domainusecasesmanagement.ClassicAuth
	googleauthUseCase      domainusecasesmanagement.GoogleAuth
	clientUseCase          domainusecasesmanagement.Client
	spacememberUseCase     domainusecasesmanagement.SpaceMember
	rolepermissionUseCase  domainusecasesmanagement.RolePermission
	spacememberroleUseCase domainusecasesmanagement.SpaceMemberRole
}

func NewHandler(spaceUseCase domainusecasesmanagement.Space, permissionUseCase domainusecasesmanagement.Permission, roleUseCase domainusecasesmanagement.Role, userUseCase domainusecasesmanagement.User, classicauthUseCase domainusecasesmanagement.ClassicAuth, googleauthUseCase domainusecasesmanagement.GoogleAuth, clientUseCase domainusecasesmanagement.Client, spacememberUseCase domainusecasesmanagement.SpaceMember, rolepermissionUseCase domainusecasesmanagement.RolePermission, spacememberroleUseCase domainusecasesmanagement.SpaceMemberRole) *handler {
	return &handler{
		spaceUseCase:           spaceUseCase,
		permissionUseCase:      permissionUseCase,
		roleUseCase:            roleUseCase,
		userUseCase:            userUseCase,
		classicauthUseCase:     classicauthUseCase,
		googleauthUseCase:      googleauthUseCase,
		clientUseCase:          clientUseCase,
		spacememberUseCase:     spacememberUseCase,
		rolepermissionUseCase:  rolepermissionUseCase,
		spacememberroleUseCase: spacememberroleUseCase,
	}
}
