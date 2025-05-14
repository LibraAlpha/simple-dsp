package traffic

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"simple-dsp/internal/bidding"
	"simple-dsp/internal/event"
	"simple-dsp/internal/rta"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Request TrafficRequest 表示来自上游的流量请求
type Request struct {
	RequestID   string            `json:"request_id"`
	UserID      string            `json:"user_id"`
	DeviceID    string            `json:"device_id"`
	IP          string            `json:"ip"`
	UserAgent   string            `json:"user_agent"`
	AdSlots     []AdSlot          `json:"ad_slots"`
	Timestamp   int64             `json:"timestamp"`
	ExtraParams map[string]string `json:"extra_params"`
}

// AdSlot 表示广告位信息
type AdSlot struct {
	SlotID   string  `json:"slot_id"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
	Position string  `json:"position"`
	AdType   string  `json:"ad_type"`
}

// Response TrafficResponse 表示返回给上游的响应
type Response struct {
	RequestID string     `json:"request_id"`
	Code      int        `json:"code"`
	Message   string     `json:"message"`
	Data      []AdResult `json:"data"`
}

// AdResult 表示广告结果
type AdResult struct {
	SlotID    string  `json:"slot_id"`
	AdID      string  `json:"ad_id"`
	BidPrice  float64 `json:"bid_price"`
	AdMarkup  string  `json:"ad_markup"`
	WinNotice string  `json:"win_notice"`
}

// Handler 流量处理器
type Handler struct {
	rtaClient     *rta.Client
	biddingEngine *bidding.Engine
	eventHandler  *event.Handler
	logger        *logger.Logger
	metrics       *metrics.Metrics
	limiter       *Limiter
}

// NewHandler 创建新的流量处理器
func NewHandler(
	rtaClient *rta.Client,
	biddingEngine *bidding.Engine,
	eventHandler *event.Handler,
	logger *logger.Logger,
	metrics *metrics.Metrics,
	limiter *Limiter,
) *Handler {
	return &Handler{
		rtaClient:     rtaClient,
		biddingEngine: biddingEngine,
		eventHandler:  eventHandler,
		logger:        logger,
		metrics:       metrics,
		limiter:       limiter,
	}
}

// GetStats 获取流量统计
func (h *Handler) GetStats(c *gin.Context) {
	// TODO: 实现流量统计
	c.JSON(http.StatusOK, gin.H{
		"total_requests":    0,
		"total_impressions": 0,
		"total_clicks":      0,
		"total_conversions": 0,
	})
}

// HandlerConfig 处理器配置
type HandlerConfig struct {
	QPS           float64       // 每秒请求数限制
	Burst         int           // 突发请求数限制
	RTATimeout    time.Duration // RTA服务超时时间
	BidTimeout    time.Duration // 竞价服务超时时间
	MaxAdSlots    int           // 最大广告位数
	MinAdSlotSize int           // 最小广告位尺寸
	MaxAdSlotSize int           // 最大广告位尺寸
}

// HandleRequest 处理流量请求
func (h *Handler) HandleRequest(c *gin.Context) {
	startTime := time.Now()
	requestID := c.GetHeader("X-Request-ID")
	if requestID == "" {
		requestID = generateRequestID()
	}

	// 记录请求开始
	h.logger.Info("收到流量请求",
		"request_id", requestID,
		"remote_addr", c.ClientIP(),
		"user_agent", c.GetHeader("User-Agent"))

	// 限流检查
	if !h.limiter.Allow() {
		h.logger.Warn("请求被限流",
			"request_id", requestID,
			"remote_addr", c.ClientIP())
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "服务繁忙，请稍后重试"})
		return
	}

	defer func() {
		// 记录请求处理时间
		duration := time.Since(startTime)
		h.metrics.RequestDuration.Observe(duration.Seconds())
		h.logger.Info("请求处理完成",
			"request_id", requestID,
			"duration_ms", duration.Milliseconds())
	}()

	// 解析请求
	var req Request
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("解析请求失败",
			"request_id", requestID,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求格式"})
		return
	}

	// 设置请求ID
	req.RequestID = requestID

	// 参数验证
	if err := h.validateRequest(&req); err != nil {
		h.logger.Error("请求参数验证失败",
			"request_id", requestID,
			"error", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 创建上下文
	ctx, cancel := context.WithTimeout(c.Request.Context(), 200*time.Millisecond)
	defer cancel()

	// RTA定向判断
	isTargeted, err := h.rtaClient.CheckTargeting(ctx, req.UserID)
	if err != nil {
		h.logger.Error("RTA定向检查失败",
			"request_id", requestID,
			"user_id", req.UserID,
			"error", err)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "服务暂时不可用"})
		return
	}

	if !isTargeted {
		h.logger.Info("用户不符合RTA定向",
			"request_id", requestID,
			"user_id", req.UserID)
		c.JSON(http.StatusOK, Response{
			RequestID: requestID,
			Code:      0,
			Message:   "用户不符合定向要求",
			Data:      []AdResult{},
		})
		return
	}

	// 转换为竞价请求
	bidReq := bidding.BidRequest{
		RequestID: requestID,
		UserID:    req.UserID,
		AdSlots:   convertToBidSlots(req.AdSlots),
	}

	// 执行竞价
	bidResp, err := h.biddingEngine.ProcessBid(ctx, bidReq)
	if err != nil {
		switch {
		case errors.Is(err, bidding.ErrNoAvailableAds):
			h.logger.Info("没有可用的广告",
				"request_id", requestID,
				"user_id", req.UserID)
			c.JSON(http.StatusOK, Response{
				RequestID: requestID,
				Code:      0,
				Message:   "没有可用的广告",
				Data:      []AdResult{},
			})
		case errors.Is(err, bidding.ErrBudgetExceeded):
			h.logger.Warn("预算已超限",
				"request_id", requestID,
				"user_id", req.UserID)
			c.JSON(http.StatusOK, Response{
				RequestID: requestID,
				Code:      0,
				Message:   "预算已超限",
				Data:      []AdResult{},
			})
		default:
			h.logger.Error("竞价处理失败",
				"request_id", requestID,
				"user_id", req.UserID,
				"error", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "竞价处理失败"})
		}
		return
	}

	// 构造响应
	resp := Response{
		RequestID: requestID,
		Code:      0,
		Message:   "success",
		Data:      convertToAdResults(bidResp),
	}

	// 记录竞价结果
	h.logger.Info("竞价成功",
		"request_id", requestID,
		"user_id", req.UserID,
		"ad_id", bidResp.AdID,
		"bid_price", bidResp.BidPrice)

	c.JSON(http.StatusOK, resp)
}

// validateRequest 验证请求参数
func (h *Handler) validateRequest(req *Request) error {
	if req.RequestID == "" {
		return ErrInvalidRequestID
	}
	if req.UserID == "" {
		return ErrInvalidUserID
	}
	if req.DeviceID == "" {
		return ErrInvalidDeviceID
	}
	if req.IP == "" {
		return ErrInvalidIP
	}
	if len(req.AdSlots) == 0 {
		return ErrNoAdSlots
	}
	if len(req.AdSlots) > 10 { // 限制最大广告位数
		return ErrTooManyAdSlots
	}

	// 验证每个广告位
	for _, slot := range req.AdSlots {
		if err := h.validateAdSlot(&slot); err != nil {
			return err
		}
	}

	return nil
}

// validateAdSlot 验证广告位参数
func (h *Handler) validateAdSlot(slot *AdSlot) error {
	if slot.SlotID == "" {
		return ErrInvalidSlotID
	}
	if slot.Width <= 0 || slot.Height <= 0 {
		return ErrInvalidAdSlotSize
	}
	if slot.MinPrice < 0 || slot.MaxPrice < 0 || slot.MinPrice > slot.MaxPrice {
		return ErrInvalidAdSlotPrice
	}
	if slot.Position == "" {
		return ErrInvalidAdSlotPosition
	}
	if slot.AdType == "" {
		return ErrInvalidAdType
	}
	return nil
}

// sendResponse 发送响应
func (h *Handler) sendResponse(w http.ResponseWriter, resp *Response) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("X-Request-ID", resp.RequestID)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		return
	}
}

// generateRequestID 生成请求ID
func generateRequestID() string {
	return time.Now().Format("20060102150405.000") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// convertToBidSlots 将流量请求的广告位转换为竞价请求的广告位
func convertToBidSlots(slots []AdSlot) []bidding.AdSlot {
	result := make([]bidding.AdSlot, len(slots))
	for i, slot := range slots {
		result[i] = bidding.AdSlot{
			SlotID:   slot.SlotID,
			Width:    slot.Width,
			Height:   slot.Height,
			MinPrice: slot.MinPrice,
			MaxPrice: slot.MaxPrice,
			Position: slot.Position,
			AdType:   slot.AdType,
		}
	}
	return result
}

// convertToAdResults 将竞价响应转换为流量响应
func convertToAdResults(resp *bidding.BidResponse) []AdResult {
	if resp == nil {
		return []AdResult{}
	}
	return []AdResult{
		{
			SlotID:    resp.SlotID,
			AdID:      resp.AdID,
			BidPrice:  resp.BidPrice,
			AdMarkup:  resp.AdMarkup,
			WinNotice: resp.WinNotice,
		},
	}
}
