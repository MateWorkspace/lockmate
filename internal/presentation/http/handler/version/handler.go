package presentationhttphandlerversion

import (
	"net/http"

	domainmodelsapp "github.com/MateWorkspace/lockmate/internal/domain/models/app"
	presentationhttpresponse "github.com/MateWorkspace/lockmate/internal/presentation/http/response"
	"github.com/gin-gonic/gin"
)

type handler struct {
	info domainmodelsapp.Info
}

func NewHandler(info domainmodelsapp.Info) *handler { return &handler{info: info} }

// VersionGet godoc
//
// @Summary Service version
// @Tags 00. Version
// @Produce json
// @Success 200 {object} presentationhttpresponse.VersionResponse
// @Router /version [get]
func (h *handler) VersionGet(c *gin.Context) {
	c.JSON(http.StatusOK, presentationhttpresponse.VersionResponse{Name: h.info.Name, Version: h.info.Version})
}
