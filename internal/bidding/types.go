package bidding

import (
	"time"
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
	SlotID   string  `json:"slot_id"`
	Width    int     `json:"width"`
	Height   int     `json:"height"`
	MinPrice float64 `json:"min_price"`
	MaxPrice float64 `json:"max_price"`
	Position string  `json:"position"`
	AdType   string  `json:"ad_type"`
	BidType  string  `json:"bid_type"`
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

// BidStrategy 出价策略
type BidStrategy struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	BidType       string    `json:"bid_type"`
	Price         float64   `json:"price"`
	Status        int       `json:"status"`
	DailyBudget   int       `json:"daily_budget"`
	IsPriceLocked bool      `json:"is_price_locked"`
	CreateTime    time.Time `json:"create_time"`
	UpdateTime    time.Time `json:"update_time"`
}

// BidStrategyFilter 出价策略过滤条件
type BidStrategyFilter struct {
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
	BidType  string `json:"bid_type"`
	MinPrice int    `json:"min_price"`
	MaxPrice int    `json:"max_type"`
}

// BiddingError 竞价错误
type BiddingError struct {
	message string
}

// NewBiddingError 创建新的竞价错误
func NewBiddingError(message string) *BiddingError {
	return &BiddingError{message: message}
}

// Error 实现 error 接口
func (e *BiddingError) Error() string {
	return e.message
}
