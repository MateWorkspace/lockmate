package presentationhttproute

import "github.com/gin-gonic/gin"

type VersionHandler interface{ VersionGet(c *gin.Context) }
