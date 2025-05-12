package event

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

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
func NewHandler(statsCollector *stats.Collector, logger *logger.Logger, metrics *metrics.Metrics) *Handler {
	return &Handler{
		statsCollector: statsCollector,
		logger:         logger,
		metrics:        metrics,
	}
}

// HandleImpression 处理展示事件
func (h *Handler) HandleImpression(w http.ResponseWriter, r *http.Request) {
	h.handleEvent(w, r, stats.EventImpression)
}

// HandleClick 处理点击事件
func (h *Handler) HandleClick(w http.ResponseWriter, r *http.Request) {
	h.handleEvent(w, r, stats.EventClick)
}

// HandleConversion 处理转化事件
func (h *Handler) HandleConversion(w http.ResponseWriter, r *http.Request) {
	h.handleEvent(w, r, stats.EventConversion)
}

// handleEvent 处理通用事件
func (h *Handler) handleEvent(w http.ResponseWriter, r *http.Request, eventType stats.EventType) {
	startTime := time.Now()
	defer func() {
		h.metrics.EventHandleDuration.WithLabelValues(string(eventType)).Observe(time.Since(startTime).Seconds())
	}()

	// 解析请求参数
	var req struct {
		RequestID   string            `json:"request_id"`
		UserID      string            `json:"user_id"`
		AdID        string            `json:"ad_id"`
		SlotID      string            `json:"slot_id"`
		BidPrice    float64           `json:"bid_price"`
		WinPrice    float64           `json:"win_price"`
		IP          string            `json:"ip"`
		UserAgent   string            `json:"user_agent"`
		ExtraParams map[string]string `json:"extra_params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("解析事件请求失败", "error", err, "event_type", eventType)
		http.Error(w, "无效的请求格式", http.StatusBadRequest)
		return
	}

	// 参数验证
	if err := h.validateEventRequest(&req, eventType); err != nil {
		h.logger.Error("事件请求参数验证失败", "error", err, "event_type", eventType)
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 构造事件数据
	event := &stats.Event{
		EventType:   eventType,
		RequestID:   req.RequestID,
		UserID:      req.UserID,
		AdID:        req.AdID,
		SlotID:      req.SlotID,
		BidPrice:    req.BidPrice,
		WinPrice:    req.WinPrice,
		Timestamp:   time.Now(),
		IP:          req.IP,
		UserAgent:   req.UserAgent,
		ExtraParams: req.ExtraParams,
	}

	// 收集事件数据
	if err := h.statsCollector.CollectEvent(r.Context(), event); err != nil {
		h.logger.Error("收集事件数据失败", "error", err, "event_type", eventType)
		http.Error(w, "处理事件失败", http.StatusInternalServerError)
		return
	}

	// 返回成功响应
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
	})
}

// validateEventRequest 验证事件请求参数
func (h *Handler) validateEventRequest(req *struct {
	RequestID   string            `json:"request_id"`
	UserID      string            `json:"user_id"`
	AdID        string            `json:"ad_id"`
	SlotID      string            `json:"slot_id"`
	BidPrice    float64           `json:"bid_price"`
	WinPrice    float64           `json:"win_price"`
	IP          string            `json:"ip"`
	UserAgent   string            `json:"user_agent"`
	ExtraParams map[string]string `json:"extra_params"`
}, eventType stats.EventType) error {
	if req.RequestID == "" {
		return ErrInvalidRequestID
	}
	if req.UserID == "" {
		return ErrInvalidUserID
	}
	if req.AdID == "" {
		return ErrInvalidAdID
	}
	if req.SlotID == "" {
		return ErrInvalidSlotID
	}

	// 展示事件特殊验证
	if eventType == stats.EventImpression {
		if req.WinPrice < 0 {
			return ErrInvalidWinPrice
		}
	}

	return nil
}

// GetEventStats 获取事件统计数据
func (h *Handler) GetEventStats(w http.ResponseWriter, r *http.Request) {
	adID := r.URL.Query().Get("ad_id")
	if adID == "" {
		http.Error(w, "缺少广告ID参数", http.StatusBadRequest)
		return
	}

	stats, err := h.statsCollector.GetRealtimeStats(r.Context(), adID)
	if err != nil {
		h.logger.Error("获取事件统计数据失败", "error", err, "ad_id", adID)
		http.Error(w, "获取统计数据失败", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"code":    0,
		"message": "success",
		"data":    stats,
	})
} 