package presentationhttpmiddleware

import (
	"io"

	domainmodelserror "github.com/MateWorkspace/lockmate/internal/domain/models/error"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
)

func Recovery() gin.HandlerFunc {
	return gin.CustomRecoveryWithWriter(io.Discard, func(c *gin.Context, _ any) {
		presentationhttputils.Error(c, domainmodelserror.NewError("request failed", domainmodelserror.Failure, nil))
	})
}
