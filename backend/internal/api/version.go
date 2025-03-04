package api

import (
	"github.com/gin-gonic/gin"
)

type VersionHandler struct {
	Version   string
	BuildTime string
}

func NewVersionHandler(version, buildTime string) *VersionHandler {
	return &VersionHandler{
		Version:   version,
		BuildTime: buildTime,
	}
}

func (h *VersionHandler) GetVersion(c *gin.Context) {
	c.JSON(200, gin.H{
		"version":   h.Version,
		"buildTime": h.BuildTime,
	})
}
