/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: manager.go
 * Project: simple-dsp
 * Description: 预算管理器实现，负责广告预算的控制和管理
 *
 * 主要功能:
 * - 管理广告主预算
 * - 控制预算消耗
 * - 提供预算查询接口
 * - 实现预算预警
 *
 * 实现细节:
 * - 使用Redis存储预算数据
 * - 实现原子预算扣减
 * - 支持多级预算控制
 * - 提供预算统计功能
 *
 * 依赖关系:
 * - simple-dsp/pkg/clients
 * - simple-dsp/pkg/metrics
 * - simple-dsp/pkg/logger
 *
 * 注意事项:
 * - 确保预算扣减的原子性
 * - 注意处理并发访问
 * - 合理设置预算预警阈值
 * - 注意数据一致性
 */

package budget

import (
	"context"
	"sync"
	"time"

	"simple-dsp/pkg/logger"
	"simple-dsp/pkg/metrics"

	"github.com/go-redis/redis/v8"
)

// Type BudgetType 预算类型
type Type string

const (
	// DailyBudget 日预算
	DailyBudget Type = "daily"
	// TotalBudget 总预算
	TotalBudget Type = "total"
)

// Budget 预算信息
type Budget struct {
	ID          string    `json:"id"`
	Type        Type      `json:"type"`
	Amount      float64   `json:"amount"`
	Spent       float64   `json:"spent"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	UpdateTime  time.Time `json:"update_time"`
	Status      string    `json:"status"`
	Description string    `json:"description"`
}

// Manager 预算管理器
type Manager struct {
	budgets     map[string]*Budget
	mu          sync.RWMutex
	logger      *logger.Logger
	metrics     *metrics.Metrics
	redisClient *redis.Client
}

// NewManager 创建新的预算管理器
func NewManager(redisClient *redis.Client, logger *logger.Logger, metrics *metrics.Metrics) *Manager {
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

	newSpent := m.redisClient.IncrBy(ctx, key, int64(amount*100)).Val() // 转换为分
	if err := m.redisClient.IncrBy(ctx, key, int64(amount*100)).Err(); err != nil {
		m.logger.Error("扣除预算失败", "error", err, "budget_id", budgetID)
		return false, err
	}

	// 更新内存中的预算信息
	budget.Spent = float64(newSpent) / 100
	budget.UpdateTime = now

	// 更新指标
	//m.metrics.BudgetSpent.WithLabelValues(budgetID).Set(budget.Spent)
	//m.metrics.BudgetRemaining.WithLabelValues(budgetID).Set(budget.Amount - budget.Spent)

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
	ID          string    `json:"id"`
	Type        Type      `json:"type"`
	Amount      float64   `json:"amount"`
	Spent       float64   `json:"spent"`
	Remaining   float64   `json:"remaining"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	Status      string    `json:"status"`
	UpdateTime  time.Time `json:"update_time"`
	IsActive    bool      `json:"is_active"`
	IsExceeded  bool      `json:"is_exceeded"`
	IsExpired   bool      `json:"is_expired"`
	Description string    `json:"description"`
}

// getBudgetKey 获取预算Redis键
func getBudgetKey(budgetID string) string {
	return "budget:spent:" + budgetID
}
