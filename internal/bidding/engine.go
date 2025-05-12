package bidding

import (
	"context"
	"errors"
	"sort"
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
}

// BidResponse 竞价响应
type BidResponse struct {
	SlotID    string  `json:"slot_id"`
	AdID      string  `json:"ad_id"`
	BidPrice  float64 `json:"bid_price"`
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
	Ad       Ad
	BidPrice float64
}

// Engine 竞价引擎
type Engine struct {
	adService    AdService
	budgetMgr    BudgetManager
	freqCtrl     *frequency.Controller
	logger       *logger.Logger
	metrics      *metrics.Metrics
	mu           sync.RWMutex
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

// NewEngine 创建新的竞价引擎
func NewEngine(adService AdService, budgetMgr BudgetManager, freqCtrl *frequency.Controller, logger *logger.Logger, metrics *metrics.Metrics) *Engine {
	return &Engine{
		adService: adService,
		budgetMgr: budgetMgr,
		freqCtrl:  freqCtrl,
		logger:    logger,
		metrics:   metrics,
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

	// 并发安全设计
	var wg sync.WaitGroup
	resultChan := make(chan BidCandidate, 10)

	// 获取广告候选集
	candidateAds := e.adService.GetCandidateAds(req.UserID)
	if len(candidateAds) == 0 {
		return nil, ErrNoAvailableAds
	}

	// 并发处理每个广告候选
	for _, ad := range candidateAds {
		wg.Add(1)
		go func(a Ad) {
			defer wg.Done()
			// 异常捕获
			defer func() {
				if r := recover(); r != nil {
					e.logger.Error("竞价处理异常",
						"panic", r,
						"ad_id", a.ID,
						"user_id", req.UserID)
				}
			}()

			// 检查广告状态
			if a.Status != "active" {
				return
			}

			// 检查频次限制
			ok, err := e.freqCtrl.CheckImpression(ctx, req.UserID, a.ID)
			if err != nil || !ok {
				e.logger.Info("频次限制",
					"user_id", req.UserID,
					"ad_id", a.ID,
					"error", err)
				return
			}

			// 预测CTR
			ctr := e.predictCTR(a, req.UserID)
			if ctr <= 0 {
				return
			}

			// 计算eCPM
			bidPrice := e.calculateECPM(a, ctr)
			if bidPrice <= 0 {
				return
			}

			// 检查预算
			ok, err = e.budgetMgr.CheckAndDeduct(ctx, a.BudgetID, bidPrice)
			if err != nil || !ok {
				return
			}

			resultChan <- BidCandidate{a, bidPrice}
		}(ad)
	}

	// 等待所有候选处理完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集并排序结果
	var candidates []BidCandidate
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

	// 选择最高出价的广告
	winner := candidates[0]
	slot := req.AdSlots[0] // 简化处理，只返回第一个广告位的结果

	// 记录曝光频次
	if err := e.freqCtrl.RecordImpression(ctx, req.UserID, winner.Ad.ID); err != nil {
		e.logger.Error("记录曝光频次失败",
			"user_id", req.UserID,
			"ad_id", winner.Ad.ID,
			"error", err)
	}

	// 更新监控指标
	e.metrics.BidPrice.WithLabelValues(winner.Ad.ID).Observe(winner.BidPrice)
	e.metrics.WinPrice.WithLabelValues(winner.Ad.ID).Observe(winner.BidPrice)

	return &BidResponse{
		SlotID:    slot.SlotID,
		AdID:      winner.Ad.ID,
		BidPrice:  winner.BidPrice,
		AdMarkup:  e.renderAd(winner.Ad),
		WinNotice: e.generateWinNotice(winner.Ad, winner.BidPrice),
	}, nil
}

// predictCTR 预测点击率
func (e *Engine) predictCTR(ad Ad, userID string) float64 {
	// TODO: 实现CTR预测模型
	// 这里使用简单的模拟实现
	return 0.01
}

// calculateECPM 计算eCPM
func (e *Engine) calculateECPM(ad Ad, ctr float64) float64 {
	// TODO: 实现eCPM计算逻辑
	// 这里使用简单的模拟实现
	basePrice := 1.0
	return basePrice * ctr * 1000
}

// renderAd 渲染广告
func (e *Engine) renderAd(ad Ad) string {
	// TODO: 实现广告渲染逻辑
	// 这里返回简单的HTML模板
	return `<div class="ad-container">
		<img src="` + ad.ImageURL + `" alt="` + ad.Title + `">
		<h3>` + ad.Title + `</h3>
		<p>` + ad.Description + `</p>
		<a href="` + ad.LandingURL + `">了解更多</a>
	</div>`
}

// generateWinNotice 生成竞价获胜通知URL
func (e *Engine) generateWinNotice(ad Ad, price float64) string {
	// TODO: 实现竞价获胜通知URL生成逻辑
	return "/api/v1/win-notice?ad_id=" + ad.ID + "&price=" + string(price)
}

var (
	// ErrInvalidBidRequest 表示无效的竞价请求
	ErrInvalidBidRequest = errors.New("无效的竞价请求")

	// ErrNoAvailableAds 表示没有可用的广告
	ErrNoAvailableAds = errors.New("没有可用的广告")

	// ErrBudgetExceeded 表示预算超限
	ErrBudgetExceeded = errors.New("预算已超限")
)
