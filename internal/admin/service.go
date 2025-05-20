package admin

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"

	"simple-dsp/internal/budget"
	"simple-dsp/internal/frequency"
	"simple-dsp/internal/stats"
	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// Service 管理后台服务
type Service struct {
	budgetMgr    *budget.Manager
	statsService *stats.Service
	logger       *logger.Logger
	metrics      *metrics.Metrics
	redis        *redis.Client
	freqCtrl     *frequency.Controller
}

// NewService 创建管理后台服务
func NewService(
	budgetMgr *budget.Manager,
	statsService *stats.Service,
	logger *logger.Logger,
	metrics *metrics.Metrics,
	freqCtrl *frequency.Controller,
) *Service {
	return &Service{
		budgetMgr:    budgetMgr,
		statsService: statsService,
		logger:       logger,
		metrics:      metrics,
		freqCtrl:     freqCtrl,
	}
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

// Budget 预算信息
type Budget struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Amount      float64   `json:"amount"`
	UsedAmount  float64   `json:"used_amount"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	AutoRenewal bool      `json:"auto_renewal"`
	CreateTime  time.Time `json:"create_time"`
	UpdateTime  time.Time `json:"update_time"`
}

// StatsOverview 统计概览
type StatsOverview struct {
	TotalImpressions int64   `json:"total_impressions"`
	TotalClicks      int64   `json:"total_clicks"`
	TotalConversions int64   `json:"total_conversions"`
	TotalSpend       float64 `json:"total_spend"`
	CTR              float64 `json:"ctr"`
	CVR              float64 `json:"cvr"`
	CPC              float64 `json:"cpc"`
	CPM              float64 `json:"cpm"`
}

// CreateAd 创建广告
func (s *Service) CreateAd(c *gin.Context) {
	var ad Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 生成广告ID
	ad.ID = generateID()
	ad.CreateTime = time.Now()
	ad.UpdateTime = time.Now()
	ad.Status = "active"

	// 保存广告信息
	ctx := c.Request.Context()
	if err := s.saveAd(ctx, &ad); err != nil {
		s.logger.Error("保存广告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存广告失败"})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// UpdateAd 更新广告
func (s *Service) UpdateAd(c *gin.Context) {
	id := c.Param("id")
	var ad Ad
	if err := c.ShouldBindJSON(&ad); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取现有广告
	ctx := c.Request.Context()
	existingAd, err := s.getAd(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "广告不存在"})
		return
	}

	// 更新广告信息
	ad.ID = id
	ad.CreateTime = existingAd.CreateTime
	ad.UpdateTime = time.Now()

	// 保存更新后的广告
	if err := s.saveAd(ctx, &ad); err != nil {
		s.logger.Error("更新广告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新广告失败"})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// DeleteAd 删除广告
func (s *Service) DeleteAd(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	// 获取广告信息
	ad, err := s.getAd(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "广告不存在"})
		return
	}

	// 标记广告为删除状态
	ad.Status = "deleted"
	ad.UpdateTime = time.Now()

	// 保存更新后的广告
	if err := s.saveAd(ctx, ad); err != nil {
		s.logger.Error("删除广告失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "删除广告失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "广告已删除"})
}

// GetAd 获取广告信息
func (s *Service) GetAd(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	ad, err := s.getAd(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "广告不存在"})
		return
	}

	c.JSON(http.StatusOK, ad)
}

// ListAds 获取广告列表
func (s *Service) ListAds(c *gin.Context) {
	ctx := c.Request.Context()
	ads, err := s.getAllAds(ctx)
	if err != nil {
		s.logger.Error("获取广告列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取广告列表失败"})
		return
	}

	c.JSON(http.StatusOK, ads)
}

// CreateBudget 创建预算
func (s *Service) CreateBudget(c *gin.Context) {
	var budget Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 生成预算ID
	budget.ID = generateID()
	budget.CreateTime = time.Now()
	budget.UpdateTime = time.Now()
	budget.Status = "active"
	budget.UsedAmount = 0

	// 保存预算信息
	ctx := c.Request.Context()
	if err := s.saveBudget(ctx, &budget); err != nil {
		s.logger.Error("保存预算失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "保存预算失败"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// UpdateBudget 更新预算
func (s *Service) UpdateBudget(c *gin.Context) {
	id := c.Param("id")
	var budget Budget
	if err := c.ShouldBindJSON(&budget); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取现有预算
	ctx := c.Request.Context()
	existingBudget, err := s.getBudget(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预算不存在"})
		return
	}

	// 更新预算信息
	budget.ID = id
	budget.CreateTime = existingBudget.CreateTime
	budget.UpdateTime = time.Now()
	budget.UsedAmount = existingBudget.UsedAmount

	// 保存更新后的预算
	if err := s.saveBudget(ctx, &budget); err != nil {
		s.logger.Error("更新预算失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新预算失败"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// GetBudget 获取预算信息
func (s *Service) GetBudget(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	budget, err := s.getBudget(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预算不存在"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// ListBudgets 获取预算列表
func (s *Service) ListBudgets(c *gin.Context) {
	ctx := c.Request.Context()
	budgets, err := s.getAllBudgets(ctx)
	if err != nil {
		s.logger.Error("获取预算列表失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取预算列表失败"})
		return
	}

	c.JSON(http.StatusOK, budgets)
}

// RenewBudget 续费预算
func (s *Service) RenewBudget(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	// 获取预算信息
	budget, err := s.getBudget(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "预算不存在"})
		return
	}

	// 更新预算时间
	budget.StartTime = time.Now()
	budget.EndTime = budget.EndTime.AddDate(0, 1, 0) // 续费一个月
	budget.UpdateTime = time.Now()

	// 保存更新后的预算
	if err := s.saveBudget(ctx, budget); err != nil {
		s.logger.Error("续费预算失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "续费预算失败"})
		return
	}

	c.JSON(http.StatusOK, budget)
}

// GetStatsOverview 获取统计概览
func (s *Service) GetStatsOverview(c *gin.Context) {
	ctx := c.Request.Context()
	overview, err := s.statsService.GetOverview(ctx)
	if err != nil {
		s.logger.Error("获取统计概览失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取统计概览失败"})
		return
	}

	c.JSON(http.StatusOK, overview)
}

// GetAdStats 获取广告统计
func (s *Service) GetAdStats(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	stats, err := s.statsService.GetAdStats(ctx, id)
	if err != nil {
		s.logger.Error("获取广告统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取广告统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetBudgetStats 获取预算统计
func (s *Service) GetBudgetStats(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	stats, err := s.statsService.GetBudgetStats(ctx, id)
	if err != nil {
		s.logger.Error("获取预算统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取预算统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetDailyStats 获取每日统计
func (s *Service) GetDailyStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := s.statsService.GetDailyStats(ctx)
	if err != nil {
		s.logger.Error("获取每日统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每日统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetHourlyStats 获取每小时统计
func (s *Service) GetHourlyStats(c *gin.Context) {
	ctx := c.Request.Context()
	stats, err := s.statsService.GetHourlyStats(ctx)
	if err != nil {
		s.logger.Error("获取每小时统计失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取每小时统计失败"})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetSystemStatus 获取系统状态
func (s *Service) GetSystemStatus(c *gin.Context) {
	ctx := c.Request.Context()
	status := gin.H{
		"redis": s.checkRedisStatus(ctx),
		"time":  time.Now().Format(time.RFC3339),
	}

	c.JSON(http.StatusOK, status)
}

// GetSystemMetrics 获取系统指标
// func (s *Service) GetSystemMetrics(c *gin.Context) {
// 	metrics := s.metrics.GetMetrics()
// 	c.JSON(http.StatusOK, metrics)
// }

// UpdateFrequencyConfig 更新频次控制配置
func (s *Service) UpdateFrequencyConfig(c *gin.Context) {
	id := c.Param("id")
	var config frequency.Config
	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的请求参数"})
		return
	}

	// 获取广告信息
	ctx := c.Request.Context()
	ad, err := s.getAd(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "广告不存在"})
		return
	}

	// 更新频次控制配置
	if err := s.freqCtrl.UpdateConfig(ctx, ad.ID, &config); err != nil {
		s.logger.Error("更新频次控制配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "更新频次控制配置失败"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "频次控制配置已更新"})
}

// GetFrequencyConfig 获取频次控制配置
func (s *Service) GetFrequencyConfig(c *gin.Context) {
	id := c.Param("id")
	ctx := c.Request.Context()

	// 获取广告信息
	ad, err := s.getAd(ctx, id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "广告不存在"})
		return
	}

	// 获取频次控制配置
	config, err := s.freqCtrl.GetConfig(ctx, ad.ID)
	if err != nil {
		s.logger.Error("获取频次控制配置失败", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取频次控制配置失败"})
		return
	}

	c.JSON(http.StatusOK, config)
}

// 内部辅助方法

func (s *Service) saveAd(ctx context.Context, ad *Ad) error {
	data, err := json.Marshal(ad)
	if err != nil {
		return err
	}

	key := "ad:" + ad.ID
	return s.redis.Set(ctx, key, data, 0).Err()
}

func (s *Service) getAd(ctx context.Context, id string) (*Ad, error) {
	key := "ad:" + id
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var ad Ad
	if err := json.Unmarshal(data, &ad); err != nil {
		return nil, err
	}

	return &ad, nil
}

func (s *Service) getAllAds(ctx context.Context) ([]Ad, error) {
	keys, err := s.redis.Keys(ctx, "ad:*").Result()
	if err != nil {
		return nil, err
	}

	var ads []Ad
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var ad Ad
		if err := json.Unmarshal(data, &ad); err != nil {
			continue
		}

		if ad.Status != "deleted" {
			ads = append(ads, ad)
		}
	}

	return ads, nil
}

func (s *Service) saveBudget(ctx context.Context, budget *Budget) error {
	data, err := json.Marshal(budget)
	if err != nil {
		return err
	}

	key := "budget:" + budget.ID
	return s.redis.Set(ctx, key, data, 0).Err()
}

func (s *Service) getBudget(ctx context.Context, id string) (*Budget, error) {
	key := "budget:" + id
	data, err := s.redis.Get(ctx, key).Bytes()
	if err != nil {
		return nil, err
	}

	var budget Budget
	if err := json.Unmarshal(data, &budget); err != nil {
		return nil, err
	}

	return &budget, nil
}

func (s *Service) getAllBudgets(ctx context.Context) ([]Budget, error) {
	keys, err := s.redis.Keys(ctx, "budget:*").Result()
	if err != nil {
		return nil, err
	}

	var budgets []Budget
	for _, key := range keys {
		data, err := s.redis.Get(ctx, key).Bytes()
		if err != nil {
			continue
		}

		var budget Budget
		if err := json.Unmarshal(data, &budget); err != nil {
			continue
		}

		budgets = append(budgets, budget)
	}

	return budgets, nil
}

func (s *Service) checkRedisStatus(ctx context.Context) string {
	if err := s.redis.Ping(ctx).Err(); err != nil {
		return "disconnected"
	}
	return "connected"
}

func generateID() string {
	return time.Now().Format("20060102150405") + randomString(6)
}

func randomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}
