package bidding

import (
	"time"
)

// BidStrategy 出价策略
type BidStrategy struct {
	ID           int64              `json:"id" db:"id"`
	Name         string             `json:"name" db:"name"`
	BidType      string             `json:"bidType" db:"bid_type"` // CPC或CPM
	Price        float64            `json:"price" db:"price"`      // CPC单位为元，CPM单位为分
	DailyBudget  float64            `json:"dailyBudget" db:"daily_budget"`
	Status       int                `json:"status" db:"status"`
	IsPriceLocked int               `json:"isPriceLocked" db:"is_price_locked"`
	Creatives    []BidStrategyCreative `json:"creatives,omitempty"`
	CreatedAt    time.Time          `json:"createdAt" db:"created_at"`
	UpdatedAt    time.Time          `json:"updatedAt" db:"updated_at"`
}

// BidStrategyCreative 出价策略素材关联
type BidStrategyCreative struct {
	ID         int64     `json:"id" db:"id"`
	StrategyID int64     `json:"strategyId" db:"strategy_id"`
	CreativeID int64     `json:"creativeId" db:"creative_id"`
	Status     int       `json:"status" db:"status"`
	CreatedAt  time.Time `json:"createdAt" db:"created_at"`
	UpdatedAt  time.Time `json:"updatedAt" db:"updated_at"`
}

// BidStrategyFilter 出价策略筛选条件
type BidStrategyFilter struct {
	BidType  string   `json:"bidType"`
	MinPrice *float64 `json:"minPrice"`
	MaxPrice *float64 `json:"maxPrice"`
	Page     int      `json:"page"`
	PageSize int      `json:"pageSize"`
}

// BidResponse 竞价响应
type BidResponse struct {
	AdID      string  `json:"adId"`
	BidPrice  float64 `json:"bidPrice"`
	BidType   string  `json:"bidType"`
	AdMarkup  string  `json:"adMarkup"`
	WinNotice string  `json:"winNotice"`
}

// BidRequest 竞价请求
type BidRequest struct {
	RequestID string   `json:"requestId"`
	UserID    string   `json:"userId"`
	AdSlots   []AdSlot `json:"adSlots"`
}

// AdSlot 广告位
type AdSlot struct {
	SlotID    string  `json:"slotId"`
	Width     int     `json:"width"`
	Height    int     `json:"height"`
	MinPrice  float64 `json:"minPrice"`
	MaxPrice  float64 `json:"maxPrice"`
	BidType   string  `json:"bidType"` // 支持的计费类型：CPC,CPM
}

// BidStrategyStats 出价策略统计数据
type BidStrategyStats struct {
	StrategyID   int64   `json:"strategyId" db:"strategy_id"`
	CreativeID   int64   `json:"creativeId" db:"creative_id"`
	Impressions  int64   `json:"impressions" db:"impressions"`
	Clicks       int64   `json:"clicks" db:"clicks"`
	Spend        float64 `json:"spend" db:"spend"`
	Date         string  `json:"date" db:"date"`
} 