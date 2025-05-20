package bidding_test

import (
	"context"
	"testing"

	"simple-dsp/internal/bidding"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"

	"go.uber.org/zap"
)

// mockRepository 实现 bidding.Repository 只需 ListBidStrategies
type mockRepository struct{}

func (m *mockRepository) ListBidStrategies(ctx context.Context, filter bidding.BidStrategyFilter) ([]bidding.BidStrategy, int64, error) {
	return []bidding.BidStrategy{{ID: "strategy-1", BidType: "CPM", Price: 2.0, Status: 1}}, 1, nil
}

func (m *mockRepository) GetBidStrategy(ctx context.Context, id int64) (*bidding.BidStrategy, error) {
	return nil, nil
}
func (m *mockRepository) CreateBidStrategy(ctx context.Context, strategy *bidding.BidStrategy) error {
	return nil
}
func (m *mockRepository) UpdateBidStrategy(ctx context.Context, strategy *bidding.BidStrategy) error {
	return nil
}
func (m *mockRepository) DeleteBidStrategy(ctx context.Context, id int64) error { return nil }
func (m *mockRepository) UpdateBidStrategyStatus(ctx context.Context, id int64, status int) error {
	return nil
}
func (m *mockRepository) AddCreative(ctx context.Context, strategyID int64, creativeID int64) error {
	return nil
}
func (m *mockRepository) RemoveCreative(ctx context.Context, strategyID int64, creativeID int64) error {
	return nil
}
func (m *mockRepository) ListCreatives(ctx context.Context, strategyID string) ([]bidding.BidStrategyCreative, error) {
	return nil, nil
}
func (m *mockRepository) GetStrategyStats(ctx context.Context, strategyID int64, startDate, endDate string) ([]bidding.BidStrategyStats, error) {
	return nil, nil
}

// mockBudgetManager 实现 bidding.BudgetManager
type mockBudgetManager struct{}

func (m *mockBudgetManager) CheckAndDeduct(ctx context.Context, budgetID string, amount float64) (bool, error) {
	return true, nil
}

// mockFreqCtrl 实现 bidding.FrequencyController
type mockFreqCtrl struct{}

func (m *mockFreqCtrl) CheckImpression(ctx context.Context, userID, adID string) (bool, error) {
	return true, nil
}

func (m *mockFreqCtrl) RecordImpression(ctx context.Context, userID, adID string) error {
	return nil
}

// mockHistogram 实现 prometheus.Histogram、prometheus.Metric、prometheus.Collector
type mockHistogram struct{}

func (m *mockHistogram) Observe(float64)                            {}
func (m *mockHistogram) Desc() *prometheus.Desc                     { return nil }
func (m *mockHistogram) Write(_ *io_prometheus_client.Metric) error { return nil }
func (m *mockHistogram) Collect(chan<- prometheus.Metric)           {}
func (m *mockHistogram) Describe(chan<- *prometheus.Desc)           {}

func TestEngine_ProcessBid(t *testing.T) {

	// 创建测试引擎
	engine := bidding.NewEngine(
		&mockRepository{},
		&mockBudgetManager{},
		&mockFreqCtrl{},
		logger.NewLogger(zap.NewNop()),
		&metrics.Metrics{Bid: &metrics.BidMetrics{Duration: &mockHistogram{}}},
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
						SlotID:   "slot-123",
						Width:    300,
						Height:   250,
						MinPrice: 1.0,
						MaxPrice: 10.0,
						Position: "banner",
						AdType:   "display",
						BidType:  "CPM",
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
