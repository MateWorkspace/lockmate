package presentationhttproute

import (
	"github.com/gin-gonic/gin"
)

type ProfileHandler interface {
	ProfileGet(c *gin.Context)
	ProfileRolesGet(c *gin.Context)
	ProfilePermissionsGet(c *gin.Context)
	ProfilePatch(c *gin.Context)
	ProfileMembershipPreferencesPatch(c *gin.Context)
	ProfileAuthMethodsGet(c *gin.Context)
	ProfileClassicAuthPatch(c *gin.Context)
	ProfilePasswordPatch(c *gin.Context)
}

func routeProfile(v1 *gin.RouterGroup, handler ProfileHandler) {
	v1.GET("/profile", handler.ProfileGet)
	v1.GET("/profile/roles", handler.ProfileRolesGet)
	v1.GET("/profile/permissions", handler.ProfilePermissionsGet)
	v1.PATCH("/profile", handler.ProfilePatch)
	v1.PATCH("/profile/membership/preferences", handler.ProfileMembershipPreferencesPatch)
	v1.GET("/profile/auth-methods", handler.ProfileAuthMethodsGet)
	v1.PATCH("/profile/classic-auth", handler.ProfileClassicAuthPatch)
	v1.PATCH("/profile/password", handler.ProfilePasswordPatch)
}
