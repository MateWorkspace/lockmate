package utility_test

import (
	"errors"
	domainmodelsentity "github.com/MateWorkspace/lockmate/internal/domain/models/entity"
	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	infrastructureutilitypermissionguard "github.com/MateWorkspace/lockmate/internal/infrastructure/utility/permission_guard"
	"testing"
)

func TestPermissionGuard(t *testing.T) {
	guard := infrastructureutilitypermissionguard.NewClaimsImpl()
	claims := &domainmodelsentity.AccessClaims{Roles: []string{"super"}, Permissions: []string{"user:get", "user:set"}}
	for _, permissions := range [][]string{nil, {}, {"user:get"}, {"user:set", "user:get", "user:get"}} {
		if err := guard.Check(claims, permissions); err != nil {
			t.Fatal(err)
		}
	}
	for _, permissions := range [][]string{{"user:delete"}, {"user:get", "user:delete"}, {"USER:GET"}, {"user:*"}, {"user:get "}} {
		if !errors.Is(guard.Check(claims, permissions), domainmodelserror.Forbidden) {
			t.Fatalf("accepted permissions %v", permissions)
		}
	}
	if !errors.Is(guard.Check(nil, nil), domainmodelserror.BadArgs) {
		t.Fatal("nil claims accepted")
	}
}
