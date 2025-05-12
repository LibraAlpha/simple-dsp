package budget

import (
	"context"
	"sync"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"
)

// BudgetType 预算类型
type BudgetType string

const (
	// DailyBudget 日预算
	DailyBudget BudgetType = "daily"
	// TotalBudget 总预算
	TotalBudget BudgetType = "total"
)

// Budget 预算信息
type Budget struct {
	ID          string     `json:"id"`
	Type        BudgetType `json:"type"`
	Amount      float64    `json:"amount"`
	Spent       float64    `json:"spent"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	UpdateTime  time.Time  `json:"update_time"`
	Status      string     `json:"status"`
	Description string     `json:"description"`
}

// Manager 预算管理器
type Manager struct {
	budgets     map[string]*Budget
	mu          sync.RWMutex
	logger      *logger.Logger
	metrics     *metrics.Metrics
	redisClient RedisClient
}

// RedisClient 定义Redis客户端接口
type RedisClient interface {
	IncrBy(ctx context.Context, key string, value int64) (int64, error)
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
}

// NewManager 创建新的预算管理器
func NewManager(redisClient RedisClient, logger *logger.Logger, metrics *metrics.Metrics) *Manager {
	return &Manager{
		budgets:     make(map[string]*Budget),
		logger:      logger,
		metrics:     metrics,
		redisClient: redisClient,
	}
}

// AddBudget 添加预算
func (m *Manager) AddBudget(budget *Budget) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.budgets[budget.ID]; exists {
		return ErrBudgetAlreadyExists
	}

	m.budgets[budget.ID] = budget
	return nil
}

// UpdateBudget 更新预算
func (m *Manager) UpdateBudget(budget *Budget) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.budgets[budget.ID]; !exists {
		return ErrBudgetNotFound
	}

	m.budgets[budget.ID] = budget
	return nil
}

// GetBudget 获取预算信息
func (m *Manager) GetBudget(id string) (*Budget, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budget, exists := m.budgets[id]
	if !exists {
		return nil, ErrBudgetNotFound
	}

	return budget, nil
}

// CheckAndDeduct 检查并扣除预算
func (m *Manager) CheckAndDeduct(ctx context.Context, budgetID string, amount float64) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	budget, exists := m.budgets[budgetID]
	if !exists {
		return false, ErrBudgetNotFound
	}

	// 检查预算状态
	if budget.Status != "active" {
		return false, ErrBudgetInactive
	}

	// 检查预算时间
	now := time.Now()
	if now.Before(budget.StartTime) || now.After(budget.EndTime) {
		return false, ErrBudgetExpired
	}

	// 检查预算余额
	if budget.Spent+amount > budget.Amount {
		return false, ErrBudgetExceeded
	}

	// 使用Redis进行原子性扣除
	key := getBudgetKey(budgetID)
	newSpent, err := m.redisClient.IncrBy(ctx, key, int64(amount*100)) // 转换为分
	if err != nil {
		m.logger.Error("扣除预算失败", "error", err, "budget_id", budgetID)
		return false, err
	}

	// 更新内存中的预算信息
	budget.Spent = float64(newSpent) / 100
	budget.UpdateTime = now

	// 更新指标
	m.metrics.BudgetSpent.WithLabelValues(budgetID).Set(budget.Spent)
	m.metrics.BudgetRemaining.WithLabelValues(budgetID).Set(budget.Amount - budget.Spent)

	return true, nil
}

// GetBudgetStatus 获取预算状态
func (m *Manager) GetBudgetStatus(budgetID string) (*BudgetStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	budget, exists := m.budgets[budgetID]
	if !exists {
		return nil, ErrBudgetNotFound
	}

	now := time.Now()
	status := &BudgetStatus{
		ID:          budget.ID,
		Type:        budget.Type,
		Amount:      budget.Amount,
		Spent:       budget.Spent,
		Remaining:   budget.Amount - budget.Spent,
		StartTime:   budget.StartTime,
		EndTime:     budget.EndTime,
		Status:      budget.Status,
		UpdateTime:  budget.UpdateTime,
		IsActive:    budget.Status == "active" && now.After(budget.StartTime) && now.Before(budget.EndTime),
		IsExceeded:  budget.Spent >= budget.Amount,
		IsExpired:   now.After(budget.EndTime),
		Description: budget.Description,
	}

	return status, nil
}

// BudgetStatus 预算状态信息
type BudgetStatus struct {
	ID          string     `json:"id"`
	Type        BudgetType `json:"type"`
	Amount      float64    `json:"amount"`
	Spent       float64    `json:"spent"`
	Remaining   float64    `json:"remaining"`
	StartTime   time.Time  `json:"start_time"`
	EndTime     time.Time  `json:"end_time"`
	Status      string     `json:"status"`
	UpdateTime  time.Time  `json:"update_time"`
	IsActive    bool       `json:"is_active"`
	IsExceeded  bool       `json:"is_exceeded"`
	IsExpired   bool       `json:"is_expired"`
	Description string     `json:"description"`
}

// getBudgetKey 获取预算Redis键
func getBudgetKey(budgetID string) string {
	return "budget:spent:" + budgetID
} 