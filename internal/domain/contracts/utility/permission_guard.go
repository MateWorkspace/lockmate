package domaincontractsutility

import domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"

type PermissionGuard interface {
	Check(claims *domainmodelsentity.AccessClaims, permissions []string) error
}
