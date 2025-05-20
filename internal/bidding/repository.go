package bidding

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

// Repository 出价策略存储接口
type Repository interface {
	// ListBidStrategies 获取出价策略列表
	ListBidStrategies(ctx context.Context, filter BidStrategyFilter) ([]BidStrategy, int64, error)
	// GetBidStrategy 获取单个出价策略
	GetBidStrategy(ctx context.Context, id int64) (*BidStrategy, error)
	// CreateBidStrategy 创建出价策略
	CreateBidStrategy(ctx context.Context, strategy *BidStrategy) error
	// UpdateBidStrategy 更新出价策略
	UpdateBidStrategy(ctx context.Context, strategy *BidStrategy) error
	// DeleteBidStrategy 删除出价策略
	DeleteBidStrategy(ctx context.Context, id int64) error
	// UpdateBidStrategyStatus 更新出价策略状态
	UpdateBidStrategyStatus(ctx context.Context, id int64, status int) error
	// AddCreative 关联素材
	AddCreative(ctx context.Context, strategyID int64, creativeID int64) error
	// RemoveCreative 移除素材
	RemoveCreative(ctx context.Context, strategyID int64, creativeID int64) error
	// ListCreatives 获取策略关联的素材列表
	ListCreatives(ctx context.Context, strategyID string) ([]BidStrategyCreative, error)
	// GetStrategyStats 获取策略统计数据
	GetStrategyStats(ctx context.Context, strategyID int64, startDate, endDate string) ([]BidStrategyStats, error)
}

// MySQLRepository MySQL实现
type MySQLRepository struct {
	db *sqlx.DB
}

// NewMySQLRepository 创建MySQL存储实现
func NewMySQLRepository(db *sqlx.DB) Repository {
	return &MySQLRepository{db: db}
}

// ListBidStrategies 获取出价策略列表
func (r *MySQLRepository) ListBidStrategies(ctx context.Context, filter BidStrategyFilter) ([]BidStrategy, int64, error) {
	var conditions []string
	var args []interface{}

	if filter.BidType != "" {
		conditions = append(conditions, "bid_type = ?")
		args = append(args, filter.BidType)
	}

	if filter.MinPrice != nil {
		conditions = append(conditions, "price >= ?")
		args = append(
			args,
			*filter.MinPrice,
		)
	}

	if filter.MaxPrice != nil {
		conditions = append(conditions, "price <= ?")
		args = append(
			args,
			*filter.MaxPrice,
		)
	}

	where := ""
	if len(conditions) > 0 {
		where = "WHERE " + strings.Join(conditions, " AND ")
	}

	// 获取总数
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM bid_strategies %s", where)
	var total int64
	if err := r.db.GetContext(ctx, &total, countQuery, args...); err != nil {
		return nil, 0, err
	}

	// 获取列表
	offset := (filter.Page - 1) * filter.PageSize
	query := fmt.Sprintf(`
		SELECT * FROM bid_strategies 
		%s
		ORDER BY id DESC 
		LIMIT ? OFFSET ?
	`, where)

	args = append(args, filter.PageSize, offset)

	var strategies []BidStrategy
	if err := r.db.SelectContext(ctx, &strategies, query, args...); err != nil {
		return nil, 0, err
	}

	// 获取关联的素材
	//for i := range strategies {
	//	creatives, err := r.ListCreatives(ctx, strategies[i].ID)
	//	if err != nil {
	//		return nil, 0, err
	//	}
	//	//strategies[i].Creatives = creatives
	//}

	return strategies, total, nil
}

// GetBidStrategy 获取单个出价策略
func (r *MySQLRepository) GetBidStrategy(ctx context.Context, id int64) (*BidStrategy, error) {
	var strategy BidStrategy
	err := r.db.GetContext(ctx, &strategy, "SELECT * FROM bid_strategies WHERE id = ?", id)
	if errors.Is(sql.ErrNoRows, err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	// 获取关联的素材
	//creatives, err := r.ListCreatives(ctx, strategy.ID)
	//if err != nil {
	//	return nil, err
	//}
	//strategy.Creatives = creatives

	return &strategy, nil
}

// CreateBidStrategy 创建出价策略
func (r *MySQLRepository) CreateBidStrategy(ctx context.Context, strategy *BidStrategy) error {
	query := `
		INSERT INTO bid_strategies (
			name, bid_type, price, daily_budget, status, is_price_locked, created_at, updated_at
		) VALUES (?, ?, ?, ?, ?, ?, NOW(), NOW())
	`
	result, err := r.db.ExecContext(ctx, query,
		strategy.Name,
		strategy.BidType,
		strategy.Price,
		strategy.DailyBudget,
		strategy.Status,
		strategy.IsPriceLocked,
	)
	if err != nil {
		return err
	}

	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	strategy.ID = strconv.FormatInt(id, 10)

	// 关联素材
	// for _, creative := range strategy.Creatives {
	// 	if err := r.AddCreative(ctx, id, creative.CreativeID); err != nil {
	// 		return err
	// 	}
	// }

	return nil
}

// UpdateBidStrategy 更新出价策略
func (r *MySQLRepository) UpdateBidStrategy(ctx context.Context, strategy *BidStrategy) error {
	// 检查是否允许更新价格
	var isPriceLocked int
	err := r.db.GetContext(ctx, &isPriceLocked, "SELECT is_price_locked FROM bid_strategies WHERE id = ?", strategy.ID)
	if err != nil {
		return err
	}

	if isPriceLocked == 1 {
		// 如果价格已锁定，则不允许更新价格
		query := `
			UPDATE bid_strategies SET 
				name = ?,
				daily_budget = ?,
				status = ?,
				updated_at = NOW()
			WHERE id = ?
		`
		_, err = r.db.ExecContext(ctx, query,
			strategy.Name,
			strategy.DailyBudget,
			strategy.Status,
			strategy.ID,
		)
	} else {
		query := `
			UPDATE bid_strategies SET 
				name = ?,
				price = ?,
				daily_budget = ?,
				status = ?,
				updated_at = NOW()
			WHERE id = ?
		`
		_, err = r.db.ExecContext(ctx, query,
			strategy.Name,
			strategy.Price,
			strategy.DailyBudget,
			strategy.Status,
			strategy.ID,
		)
	}
	return err
}

// DeleteBidStrategy 删除出价策略
func (r *MySQLRepository) DeleteBidStrategy(ctx context.Context, id int64) error {
	_, err := r.db.ExecContext(ctx, "DELETE FROM bid_strategies WHERE id = ?", id)
	return err
}

// UpdateBidStrategyStatus 更新出价策略状态
func (r *MySQLRepository) UpdateBidStrategyStatus(ctx context.Context, id int64, status int) error {
	query := "UPDATE bid_strategies SET status = ?, updated_at = NOW() WHERE id = ?"
	_, err := r.db.ExecContext(ctx, query, status, id)
	return err
}

// AddCreative 关联素材
func (r *MySQLRepository) AddCreative(ctx context.Context, strategyID int64, creativeID int64) error {
	query := `
		INSERT INTO bid_strategy_creatives (
			strategy_id, creative_id, status, created_at, updated_at
		) VALUES (?, ?, 1, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query, strategyID, creativeID)
	return err
}

// RemoveCreative 移除素材
func (r *MySQLRepository) RemoveCreative(ctx context.Context, strategyID int64, creativeID int64) error {
	query := "DELETE FROM bid_strategy_creatives WHERE strategy_id = ? AND creative_id = ?"
	_, err := r.db.ExecContext(ctx, query, strategyID, creativeID)
	return err
}

// ListCreatives 获取策略关联的素材列表
func (r *MySQLRepository) ListCreatives(ctx context.Context, strategyID string) ([]BidStrategyCreative, error) {
	var creatives []BidStrategyCreative
	query := "SELECT * FROM bid_strategy_creatives WHERE strategy_id = ?"
	err := r.db.SelectContext(ctx, &creatives, query, strategyID)
	return creatives, err
}

// GetStrategyStats 获取策略统计数据
func (r *MySQLRepository) GetStrategyStats(ctx context.Context, strategyID int64, startDate, endDate string) ([]BidStrategyStats, error) {
	query := `
		SELECT 
			strategy_id,
			creative_id,
			SUM(impressions) as impressions,
			SUM(clicks) as clicks,
			SUM(spend) as spend,
			date
		FROM bid_strategy_stats
		WHERE strategy_id = ? AND date BETWEEN ? AND ?
		GROUP BY strategy_id, creative_id, date
		ORDER BY date DESC
	`
	var stats []BidStrategyStats
	err := r.db.SelectContext(ctx, &stats, query, strategyID, startDate, endDate)
	return stats, err
}
