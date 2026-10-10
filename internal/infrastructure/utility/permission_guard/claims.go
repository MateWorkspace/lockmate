package infrastructureutilitypermissionguard

import (
	"slices"

	domaincontractsutility "github.com/MateWorkspace/lockmate/internal/domain/contracts/utility"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
)

type claimsImpl struct{}

func NewClaimsImpl() domaincontractsutility.PermissionGuard { return &claimsImpl{} }

func (g *claimsImpl) Check(claims *domainmodelsentity.AccessClaims, permissions []string) error {
	if claims == nil {
		return domainmodelserror.NewError("access claims are required", domainmodelserror.BadArgs, nil)
	}
	for _, permission := range permissions {
		if !slices.Contains(claims.Permissions, permission) {
			return domainmodelserror.NewError("access is forbidden", domainmodelserror.Forbidden, nil)
		}
	}
	return nil
}
