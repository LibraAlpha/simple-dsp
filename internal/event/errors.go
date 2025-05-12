package event

import "errors"

var (
	// ErrInvalidRequestID 表示请求ID无效
	ErrInvalidRequestID = errors.New("无效的请求ID")

	// ErrInvalidUserID 表示用户ID无效
	ErrInvalidUserID = errors.New("无效的用户ID")

	// ErrInvalidAdID 表示广告ID无效
	ErrInvalidAdID = errors.New("无效的广告ID")

	// ErrInvalidSlotID 表示广告位ID无效
	ErrInvalidSlotID = errors.New("无效的广告位ID")

	// ErrInvalidWinPrice 表示无效的成交价格
	ErrInvalidWinPrice = errors.New("无效的成交价格")

	// ErrInvalidEventType 表示无效的事件类型
	ErrInvalidEventType = errors.New("无效的事件类型")

	// ErrEventProcessFailed 表示事件处理失败
	ErrEventProcessFailed = errors.New("事件处理失败")

	// ErrStatsNotFound 表示统计数据不存在
	ErrStatsNotFound = errors.New("统计数据不存在")
) 