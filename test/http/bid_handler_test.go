package http_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"simple-dsp/internal/bidding"
	"simple-dsp/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

func TestBidHandler_ProcessBid(t *testing.T) {
	// 设置Gin测试模式
	gin.SetMode(gin.TestMode)

	// 创建路由
	router := gin.New()
	engine := bidding.NewEngine(nil, nil, nil, logger.NewLogger(zap.NewNop()), nil)

	// 注册路由
	router.POST("/bid", func(c *gin.Context) {
		var req struct {
			RequestID string           `json:"request_id"`
			UserID    string           `json:"user_id"`
			DeviceID  string           `json:"device_id"`
			IP        string           `json:"ip"`
			AdSlots   []bidding.AdSlot `json:"ad_slots"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		bidReq := bidding.BidRequest{
			RequestID: req.RequestID,
			UserID:    req.UserID,
			DeviceID:  req.DeviceID,
			IP:        req.IP,
			AdSlots:   req.AdSlots,
		}

		resp, err := engine.ProcessBid(c.Request.Context(), bidReq)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, resp)
	})

	tests := []struct {
		name       string
		request    map[string]interface{}
		wantStatus int
	}{
		{
			name: "正常请求",
			request: map[string]interface{}{
				"request_id": "test-123",
				"user_id":    "user-123",
				"device_id":  "device-123",
				"ip":         "127.0.0.1",
				"ad_slots": []map[string]interface{}{
					{
						"slot_id":   "slot-123",
						"width":     300,
						"height":    250,
						"min_price": 1.0,
						"max_price": 10.0,
						"position":  "banner",
						"ad_type":   "display",
						"bid_type":  "CPM",
					},
				},
			},
			wantStatus: http.StatusOK,
		},
		{
			name: "无效请求",
			request: map[string]interface{}{
				"request_id": "test-124",
			},
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jsonData, err := json.Marshal(tt.request)
			if err != nil {
				t.Fatalf("Failed to marshal request: %v", err)
			}

			req := httptest.NewRequest("POST", "/bid", bytes.NewBuffer(jsonData))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("ProcessBid() status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}
