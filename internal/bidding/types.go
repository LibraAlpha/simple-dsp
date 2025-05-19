/*
 * Copyright (c) 2024 Simple DSP
 *
 * File: types.go
 * Project: simple-dsp
 * Description: 竞价引擎相关的类型定义和接口声明
 *
 * 主要功能:
 * - 定义竞价请求和响应结构
 * - 定义广告位和策略类型
 * - 声明数据访问接口
 * - 定义错误类型和常量
 *
 * 实现细节:
 * - 使用Go标准类型定义
 * - 实现错误类型封装
 * - 定义通用接口
 * - 提供类型转换方法
 *
 * 依赖关系:
 * - time
 * - simple-dsp/pkg/metrics
 *
 * 注意事项:
 * - 保持类型定义的一致性
 * - 注意字段的序列化
 * - 合理使用接口定义
 * - 注意类型安全性
 */

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
	MinPrice *int   `json:"min_price"`
	MaxPrice *int   `json:"max_type"`
}

// Error BiddingError 竞价错误
type Error struct {
	message string
}

// NewBiddingError 创建新的竞价错误
func NewBiddingError(message string) *Error {
	return &Error{message: message}
}

// Error 实现 error 接口
func (e *Error) Error() string {
	return e.message
}
