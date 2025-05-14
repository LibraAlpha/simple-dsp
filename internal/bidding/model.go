package bidding

import (
	"time"
)

// BidStrategyCreative 出价策略素材关联
type BidStrategyCreative struct {
	ID         int64     `json:"id" db:"id"`
	StrategyID int64     `json:"strategyId" db:"strategy_id"`
	CreativeID int64     `json:"creativeId" db:"creative_id"`
	Status     int       `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// BidStrategyStats 出价策略统计数据
type BidStrategyStats struct {
	StrategyID  int64   `json:"strategyId" db:"strategy_id"`
	CreativeID  int64   `json:"creativeId" db:"creative_id"`
	Impressions int64   `json:"impressions" db:"impressions"`
	Clicks      int64   `json:"clicks" db:"clicks"`
	Spend       float64 `json:"spend" db:"spend"`
	Date        string  `json:"date" db:"date"`
}
