package router

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"simple-dsp/internal/event"
	"simple-dsp/internal/traffic"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Handler 路由处理器
type Handler struct {
	trafficHandler *traffic.Handler
	eventHandler   *event.Handler
	logger         *logger.Logger
	metrics        *metrics.Metrics
}

// NewHandler 创建新的路由处理器
func NewHandler(
	trafficHandler *traffic.Handler,
	eventHandler *event.Handler,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *Handler {
	return &Handler{
		trafficHandler: trafficHandler,
		eventHandler:   eventHandler,
		logger:         logger,
		metrics:        metrics,
	}
}

// RegisterRoutes 注册路由
func (h *Handler) RegisterRoutes(router *gin.Engine) {
	// 流量接入接口
	router.POST("/api/v1/traffic", h.trafficHandler.HandleRequest)

	// 事件处理接口
	router.POST("/api/v1/events/impression", h.eventHandler.HandleImpression)
	router.POST("/api/v1/events/click", h.eventHandler.HandleClick)
	router.POST("/api/v1/events/conversion", h.eventHandler.HandleConversion)
	router.GET("/api/v1/events/stats", h.eventHandler.GetEventStats)

	// 健康检查接口
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}
