package bidding

import (
	"context"
	"fmt"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
	"sort"
	"sync"
	"time"
)

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

	// 获取出价策略列表
	strategies, err, _ := e.repository.ListBidStrategies(ctx, BidStrategyFilter{
		Page:     1,
		PageSize: 100,
	})
	if err != nil {
		e.logger.Error("获取出价策略失败", "error", err)
		return nil, fmt.Errorf("获取出价策略失败: %w", err)
	}

	// 如果没有可用的出价策略
	if len(strategies) == 0 {
		return nil, ErrNoAvailableAds
	}

	// 对每个广告位进行竞价
	for _, slot := range req.AdSlots {
		// 获取候选广告
		candidates := e.getBidCandidates(ctx, req.UserID, slot, strategies)
		if len(candidates) == 0 {
			continue
		}

		// 选择最优出价
		winner := e.selectWinner(candidates)
		if winner == nil {
			continue
		}

		// 检查预算
		ok, err := e.budgetMgr.CheckAndDeduct(ctx, winner.Strategy.ID, winner.BidPrice)
		if err != nil {
			e.logger.Error("检查预算失败", "error", err)
			continue
		}
		if !ok {
			e.logger.Warn("预算不足", "strategy_id", winner.Strategy.ID)
			continue
		}

		// 检查频次
		ok, err = e.freqCtrl.CheckImpression(ctx, req.UserID, winner.Strategy.ID)
		if err != nil {
			e.logger.Error("检查频次失败", "error", err)
			continue
		}
		if !ok {
			e.logger.Warn("频次超限", "strategy_id", winner.Strategy.ID)
			continue
		}

		// 返回竞价响应
		return &BidResponse{
			SlotID:    slot.SlotID,
			AdID:      winner.Strategy.ID,
			BidPrice:  winner.BidPrice,
			BidType:   winner.Strategy.BidType,
			AdMarkup:  "", // TODO: 生成广告物料
			WinNotice: "", // TODO: 生成获胜通知URL
		}, nil
	}

	return nil, ErrNoAvailableAds
}

// getBidCandidates 获取竞价候选
func (e *Engine) getBidCandidates(ctx context.Context, userID string, slot AdSlot, strategies []BidStrategy) []BidCandidate {
	var candidates []BidCandidate

	for _, strategy := range strategies {
		// 检查策略状态
		if strategy.Status != 1 {
			continue
		}

		// 计算出价
		bidPrice := e.calculateBidPrice(strategy, slot)
		if bidPrice < slot.MinPrice || bidPrice > slot.MaxPrice {
			continue
		}

		// 计算CTR
		ctr := e.estimateCTR(strategy, userID, slot)

		candidates = append(candidates, BidCandidate{
			Strategy: strategy,
			BidPrice: bidPrice,
			CTR:      ctr,
		})
	}

	return candidates
}

// selectWinner 选择最优出价
func (e *Engine) selectWinner(candidates []BidCandidate) *BidCandidate {
	if len(candidates) == 0 {
		return nil
	}

	// 按 eCPM 排序
	sort.Slice(candidates, func(i, j int) bool {
		return candidates[i].BidPrice*candidates[i].CTR > candidates[j].BidPrice*candidates[j].CTR
	})

	return &candidates[0]
}

// calculateBidPrice 计算出价
func (e *Engine) calculateBidPrice(strategy BidStrategy, slot AdSlot) float64 {
	// TODO: 实现更复杂的出价逻辑
	return strategy.Price
}

// estimateCTR 预估点击率
func (e *Engine) estimateCTR(strategy BidStrategy, userID string, slot AdSlot) float64 {
	// TODO: 实现更复杂的CTR预估逻辑
	return 0.01
}
