package presentationhttpmiddleware

import (
	domainusecasesmisc "github.com/MateWorkspace/lockmate/internal/domain/usecases/misc"
	presentationhttputils "github.com/MateWorkspace/lockmate/internal/presentation/http/utils"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func Trace(tracer domainusecasesmisc.Tracer) gin.HandlerFunc {
	return func(c *gin.Context) {
		execId, _ := uuid.Parse(c.GetHeader(ClientExecIdHeader))
		ctx, err := tracer.Inject(c.Request.Context(), c.GetHeader(ClientSlugHeader), execId)
		if err != nil {
			presentationhttputils.Error(c, err)
			return
		}
		if ctx == nil {
			presentationhttputils.Error(c, presentationhttputils.MissingResponse("trace context"))
			return
		}
		c.Request = c.Request.WithContext(ctx)
		c.Next()
	}
}
