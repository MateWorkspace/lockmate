package presentationhttproute

import "github.com/gin-gonic/gin"

type AuthFlowHandler interface {
	AuthGoogleStartPost(c *gin.Context)
	AuthGoogleLaunchGet(c *gin.Context)
	AuthGoogleCallbackGet(c *gin.Context)
	AuthExchangePost(c *gin.Context)
}

func routeAuthFlow(authenticated, browser *gin.RouterGroup, handler AuthFlowHandler) {
	authenticated.POST("/auth/google/start", handler.AuthGoogleStartPost)
	authenticated.POST("/auth/exchange", handler.AuthExchangePost)
	browser.GET("/auth/google/launch", handler.AuthGoogleLaunchGet)
	browser.GET("/auth/google/callback", handler.AuthGoogleCallbackGet)
}
