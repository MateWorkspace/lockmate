package presentationhttproute

import "github.com/gin-gonic/gin"

type JwksHandler interface {
	JwksGet(c *gin.Context)
}

func routeJwks(v1 *gin.RouterGroup, handler JwksHandler) {
	v1.GET("/jwks", handler.JwksGet)
}
