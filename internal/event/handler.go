package event

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simple-dsp/internal/stats"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Handler 事件处理器
type Handler struct {
	statsCollector *stats.Collector
	logger         *logger.Logger
	metrics        *metrics.Metrics
}

// NewHandler 创建新的事件处理器
func NewHandler(
	statsCollector *stats.Collector,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *Handler {
	return &Handler{
		statsCollector: statsCollector,
		logger:         logger,
		metrics:        metrics,
	}
}

// HandleImpression 处理展示事件
func (h *Handler) HandleImpression(c *gin.Context) {
	var event stats.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("解析展示事件失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	event.EventType = stats.EventImpression
	event.Timestamp = time.Now()

	if err := h.statsCollector.CollectEvent(c.Request.Context(), &event); err != nil {
		h.logger.Error("记录展示事件失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录展示事件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleClick 处理点击事件
func (h *Handler) HandleClick(c *gin.Context) {
	var event stats.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("解析点击事件失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	event.EventType = stats.EventClick
	event.Timestamp = time.Now()

	if err := h.statsCollector.CollectEvent(c.Request.Context(), &event); err != nil {
		h.logger.Error("记录点击事件失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录点击事件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// HandleConversion 处理转化事件
func (h *Handler) HandleConversion(c *gin.Context) {
	var event stats.Event
	if err := c.ShouldBindJSON(&event); err != nil {
		h.logger.Error("解析转化事件失败", "error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	event.EventType = stats.EventConversion
	event.Timestamp = time.Now()

	if err := h.statsCollector.CollectEvent(c.Request.Context(), &event); err != nil {
		h.logger.Error("记录转化事件失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "记录转化事件失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// GetEventStats 获取事件统计
func (h *Handler) GetEventStats(c *gin.Context) {
	adID := c.Query("ad_id")
	if adID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "缺少广告ID参数"})
		return
	}

	stats, err := h.statsCollector.GetRealtimeStats(c.Request.Context(), adID)
	if err != nil {
		h.logger.Error("获取事件统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取事件统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}
