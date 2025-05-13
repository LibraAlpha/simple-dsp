package bidding_test

import (
	"context"
	"testing"

	"simple-dsp/internal/bidding"
	"simple-dsp/pkg/logger"
)

func TestEngine_ProcessBid(t *testing.T) {
	// 创建测试引擎
	engine := bidding.NewEngine(
		nil, // 这里可以传入mock的repository
		nil, // mock的预算管理器
		nil, // mock的频次控制器
		logger.NewLogger(),
		nil, // mock的指标收集器
	)

	tests := []struct {
		name    string
		request bidding.BidRequest
		wantErr bool
	}{
		{
			name: "正常竞价请求",
			request: bidding.BidRequest{
				RequestID: "test-123",
				UserID:    "user-123",
				DeviceID:  "device-123",
				IP:        "127.0.0.1",
				AdSlots: []bidding.AdSlot{
					{
						SlotID:    "slot-123",
						Width:     300,
						Height:    250,
						MinPrice:  1.0,
						MaxPrice:  10.0,
						Position:  "banner",
						AdType:    "display",
						BidType:   "CPM",
					},
				},
			},
			wantErr: false,
		},
		{
			name: "无效的广告位",
			request: bidding.BidRequest{
				RequestID: "test-124",
				UserID:    "user-124",
				DeviceID:  "device-124",
				IP:        "127.0.0.1",
				AdSlots:   []bidding.AdSlot{},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resp, err := engine.ProcessBid(context.Background(), tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("ProcessBid() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && resp == nil {
				t.Error("Expected non-nil response when no error")
			}
		})
	}
} 