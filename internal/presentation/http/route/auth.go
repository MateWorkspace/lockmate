package presentationhttproute

import (
	"github.com/gin-gonic/gin"
)

type AuthHandler interface {
	AuthSignInPost(c *gin.Context)
	AuthRegisterPost(c *gin.Context)
	AuthRefreshPost(c *gin.Context)
}

func routeAuth(v1 *gin.RouterGroup, handler AuthHandler) {
	v1.POST("/auth/sign-in", handler.AuthSignInPost)
	v1.POST("/auth/register", handler.AuthRegisterPost)
	v1.POST("/auth/refresh", handler.AuthRefreshPost)
}
