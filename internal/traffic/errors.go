package traffic

import "errors"

var (
	// ErrInvalidRequestID 表示请求ID无效
	ErrInvalidRequestID = errors.New("无效的请求ID")

	// ErrInvalidUserID 表示用户ID无效
	ErrInvalidUserID = errors.New("无效的用户ID")

	// ErrInvalidDeviceID 表示设备ID无效
	ErrInvalidDeviceID = errors.New("无效的设备ID")

	// ErrInvalidIP 表示IP地址无效
	ErrInvalidIP = errors.New("无效的IP地址")

	// ErrNoAdSlots 表示没有广告位
	ErrNoAdSlots = errors.New("没有广告位信息")

	// ErrTooManyAdSlots 表示广告位数量过多
	ErrTooManyAdSlots = errors.New("广告位数量超过限制")

	// ErrInvalidAdSlot 表示广告位信息无效
	ErrInvalidAdSlot = errors.New("无效的广告位信息")

	// ErrInvalidSlotID 表示广告位ID无效
	ErrInvalidSlotID = errors.New("无效的广告位ID")

	// ErrInvalidAdSlotSize 表示广告位尺寸无效
	ErrInvalidAdSlotSize = errors.New("无效的广告位尺寸")

	// ErrInvalidAdSlotPrice 表示广告位价格无效
	ErrInvalidAdSlotPrice = errors.New("无效的广告位价格")

	// ErrInvalidAdSlotPosition 表示广告位位置无效
	ErrInvalidAdSlotPosition = errors.New("无效的广告位位置")

	// ErrInvalidAdType 表示广告类型无效
	ErrInvalidAdType = errors.New("无效的广告类型")

	// ErrRTAError 表示RTA服务错误
	ErrRTAError = errors.New("RTA服务错误")

	// ErrBiddingError 表示竞价服务错误
	ErrBiddingError = errors.New("竞价服务错误")

	// ErrBudgetExceeded 表示预算超限
	ErrBudgetExceeded = errors.New("预算已超限")

	// ErrServiceUnavailable 表示服务不可用
	ErrServiceUnavailable = errors.New("服务暂时不可用")

	// ErrRequestTimeout 表示请求超时
	ErrRequestTimeout = errors.New("请求处理超时")

	// ErrRateLimited 表示请求被限流
	ErrRateLimited = errors.New("请求被限流")

	// ErrInvalidRequestFormat 表示请求格式无效
	ErrInvalidRequestFormat = errors.New("无效的请求格式")

	// ErrInvalidResponseFormat 表示响应格式无效
	ErrInvalidResponseFormat = errors.New("无效的响应格式")
) 