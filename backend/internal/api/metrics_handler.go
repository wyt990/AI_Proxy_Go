package api

import (
	"AI_Proxy_Go/backend/internal/service"
	"net/http"

	"github.com/gin-gonic/gin"
)

type MetricsHandler struct {
	metricsService *service.MetricsService
}

func NewMetricsHandler(metricsService *service.MetricsService) *MetricsHandler {
	return &MetricsHandler{metricsService: metricsService}
}

// GetLatestMetrics 获取最新的系统指标
func (h *MetricsHandler) GetLatestMetrics(c *gin.Context) {
	metrics, err := h.metricsService.GetLatestMetrics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取系统指标失败"})
		return
	}

	c.JSON(http.StatusOK, metrics)
}
