package bidding

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
	"simple-dsp/internal/frequency"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// BidRequest 竞价请求
type BidRequest struct {
	RequestID string   `json:"request_id"`
	UserID    string   `json:"user_id"`
	DeviceID  string   `json:"device_id"`
	IP        string   `json:"ip"`
	AdSlots   []AdSlot `json:"ad_slots"`
}

// AdSlot 广告位信息
type AdSlot struct {
	SlotID    string  `json:"slot_id"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	MinPrice  float64 `json:"min_price"`
	MaxPrice  float64 `json:"max_price"`
	Position  string  `json:"position"`
	AdType    string  `json:"ad_type"`
	BidType   string  `json:"bid_type"`
}

// BidResponse 竞价响应
type BidResponse struct {
	SlotID    string  `json:"slot_id"`
	AdID      string  `json:"ad_id"`
	BidPrice  float64 `json:"bid_price"`
	BidType   string  `json:"bid_type"`
	AdMarkup  string  `json:"ad_markup"`
	WinNotice string  `json:"win_notice"`
}

// Ad 广告信息
type Ad struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	ImageURL    string    `json:"image_url"`
	LandingURL  string    `json:"landing_url"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	BudgetID    string    `json:"budget_id"`
	Status      string    `json:"status"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// BidCandidate 竞价候选
type BidCandidate struct {
	Strategy BidStrategy
	BidPrice float64
	CTR      float64
}

// Engine 竞价引擎
type Engine struct {
	repository Repository
	budgetMgr  BudgetManager
	freqCtrl   FrequencyController
	logger     *logger.Logger
	metrics    *metrics.Metrics
	mu         sync.RWMutex
}

// AdService 广告服务接口
type AdService interface {
	GetCandidateAds(userID string) []Ad
	GetAdByID(adID string) (*Ad, error)
}

// BudgetManager 预算管理接口
type BudgetManager interface {
	CheckAndDeduct(ctx context.Context, budgetID string, amount float64) (bool, error)
}

// FrequencyController 频率控制接口
type FrequencyController interface {
	CheckImpression(ctx context.Context, userID, adID string) (bool, error)
	RecordImpression(ctx context.Context, userID, adID string) error
}

// NewEngine 创建新的竞价引擎
func NewEngine(
	repository Repository,
	budgetMgr BudgetManager,
	freqCtrl FrequencyController,
	logger *logger.Logger,
	metrics *metrics.Metrics,
) *Engine {
	return &Engine{
		repository: repository,
		budgetMgr:  budgetMgr,
		freqCtrl:   freqCtrl,
		logger:     logger,
		metrics:    metrics,
	}
}

// ProcessBid 处理竞价请求
func (e *Engine) ProcessBid(ctx context.Context, req BidRequest) (*BidResponse, error) {
	startTime := time.Now()
	defer func() {
		e.metrics.BidDuration.Observe(time.Since(startTime).Seconds())
	}()

	// 防御性编程：空请求检查
	if req.UserID == "" || len(req.AdSlots) == 0 {
		return nil, ErrInvalidBidRequest
	}

	// 获取所有活跃的出价策略
	strategies, err := e.repository.ListBidStrategies(ctx, BidStrategyFilter{
		Page:     1,
		PageSize: 1000,
	})
	if err != nil {
		return nil, err
	}

	// 并发安全设计
	var wg sync.WaitGroup
	resultChan := make(chan *BidCandidate, len(strategies))

	// 获取广告位
	slot := req.AdSlots[0] // 简化处理，只处理第一个广告位

	// 并发处理每个出价策略
	for _, strategy := range strategies {
		if strategy.Status != 1 {
			continue
		}

		// 检查计费类型是否匹配
		if !strings.Contains(slot.BidType, strategy.BidType) {
			continue
		}

		wg.Add(1)
		go func(s BidStrategy) {
			defer wg.Done()
			// 异常捕获
			defer func() {
				if r := recover(); r != nil {
					e.logger.Error("竞价处理异常",
						"panic", r,
						"strategy_id", s.ID,
						"user_id", req.UserID)
				}
			}()

			// 检查频次限制
			ok, err := e.freqCtrl.CheckImpression(ctx, req.UserID, s.ID)
			if err != nil || !ok {
				e.logger.Info("频次限制",
					"user_id", req.UserID,
					"strategy_id", s.ID,
					"error", err)
				return
			}

			// 预测CTR
			ctr := e.predictCTR(s, req.UserID)
			if ctr <= 0 {
				return
			}

			// 计算最终出价
			bidPrice := e.calculateFinalBidPrice(s, ctr)
			if bidPrice <= 0 {
				return
			}

			// 检查出价范围
			if bidPrice < slot.MinPrice || bidPrice > slot.MaxPrice {
				return
			}

			// 检查预算
			ok, err = e.budgetMgr.CheckAndDeduct(ctx, s.ID, bidPrice)
			if err != nil || !ok {
				return
			}

			resultChan <- &BidCandidate{
				Strategy: s,
				BidPrice: bidPrice,
				CTR:      ctr,
			}
		}(strategy)
	}

	// 等待所有候选处理完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集并排序结果
	var candidates []*BidCandidate
	for c := range resultChan {
		candidates = append(candidates, c)
	}
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].BidPrice > candidates[j].BidPrice
	})

	// 空结果处理
	if len(candidates) == 0 {
		return nil, ErrNoAvailableAds
	}

	// 选择最高出价的策略
	winner := candidates[0]

	// 记录曝光频次
	if err := e.freqCtrl.RecordImpression(ctx, req.UserID, winner.Strategy.ID); err != nil {
		e.logger.Error("记录曝光频次失败",
			"user_id", req.UserID,
			"strategy_id", winner.Strategy.ID,
			"error", err)
	}

	// 更新监控指标
	e.metrics.BidPrice.WithLabelValues(winner.Strategy.BidType).Observe(winner.BidPrice)
	e.metrics.WinPrice.WithLabelValues(winner.Strategy.BidType).Observe(winner.BidPrice)

	return &BidResponse{
		AdID:      fmt.Sprintf("%d", winner.Strategy.ID),
		BidPrice:  winner.BidPrice,
		BidType:   winner.Strategy.BidType,
		AdMarkup:  e.renderAd(winner.Strategy),
		WinNotice: e.generateWinNotice(winner.Strategy, winner.BidPrice),
	}, nil
}

// calculateFinalBidPrice 计算最终出价
func (e *Engine) calculateFinalBidPrice(strategy BidStrategy, ctr float64) float64 {
	switch strategy.BidType {
	case "CPC":
		// CPC出价 = 原始出价(元) * CTR
		return strategy.Price * ctr
	case "CPM":
		// CPM出价 = 原始出价(分) / 1000
		return strategy.Price / 1000
	default:
		return 0
	}
}

// predictCTR 预测点击率
func (e *Engine) predictCTR(strategy BidStrategy, userID string) float64 {
	// TODO: 实现CTR预测模型
	// 这里使用简单的模拟实现
	return 0.01
}

// renderAd 渲染广告
func (e *Engine) renderAd(strategy BidStrategy) string {
	// TODO: 实现广告渲染逻辑
	return fmt.Sprintf(`<div class="ad" data-strategy="%d">广告内容</div>`, strategy.ID)
}

// generateWinNotice 生成竞价获胜通知URL
func (e *Engine) generateWinNotice(strategy BidStrategy, price float64) string {
	return fmt.Sprintf("/api/v1/win-notice?strategy_id=%d&price=%.4f&type=%s",
		strategy.ID, price, strategy.BidType)
}

var (
	// ErrInvalidBidRequest 表示无效的竞价请求
	ErrInvalidBidRequest = errors.New("无效的竞价请求")

	// ErrNoAvailableAds 表示没有可用的广告
	ErrNoAvailableAds = errors.New("没有可用的广告")

	// ErrBudgetExceeded 表示预算超限
	ErrBudgetExceeded = errors.New("预算已超限")
)
